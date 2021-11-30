package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/gorilla/mux"
	"github.com/nickbryan/go-template/service/app"
)

// Server defines a HTTP server for handling Rest requests.
type Server struct {
	environment *app.Environment
	router      *mux.Router
}

// NewServer initialises a new Server with a router.
func NewServer(e *app.Environment) *Server {
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := gabs.New()

		if _, err := msg.SetP("resource not found", "error.message"); err != nil {
			e.Logger().Error(fmt.Sprintf("set json path failed in not found handler: %s", err))
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		(&responder{
			ResponseWriter: w,
			logger:         e.Logger(),
		}).Respond(http.StatusNotFound, msg.Data())
	})

	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := gabs.New()

		if _, err := msg.SetP(fmt.Sprintf("method %s is not allowed", r.Method), "error.message"); err != nil {
			e.Logger().Error(fmt.Sprintf("set json path failed in method not allowed handler: %s", err))
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		(&responder{
			ResponseWriter: w,
			logger:         e.Logger(),
		}).Respond(http.StatusMethodNotAllowed, msg.Data())
	})

	return &Server{
		e,
		router,
	}
}

// Start the server and listen for incoming requests.
func (s *Server) Start() error {
	conf := s.environment.Config()

	srv := &http.Server{
		Addr:         conf.Server.Address,
		WriteTimeout: conf.Server.WriteTimeout * time.Second,
		ReadTimeout:  conf.Server.ReadTimeout * time.Second,
		IdleTimeout:  conf.Server.IdleTimeout * time.Second,
		Handler:      s,
	}

	// Here we start the web server listing for connections and serving responses. When the server
	// closes we may get back an error so we create a channel that allows us to receive that error
	// to be reported on later. If the error is http.ErrServerClosed then we ignore it as we expect
	// that to happen at some point. We start this in a separate go routine to allow us to block on
	// an os.Interrupt signal later.
	errChan := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("ann error occured on ListenAndServe: %w", err)
		}

		close(errChan)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan // Block awaiting an interrupt signal

	// Once we receive an interrupt signal, we know it is time to shut down the web server. We will
	// allow x seconds for graceful shutdown before we force the close. This is handled through
	// our context.
	ctx, cancel := context.WithTimeout(context.Background(), conf.Server.ShutdownTimeout*time.Second)
	defer cancel()

	// Shutting down the server should trigger our http.ErrServerClosed that we ignore
	// in the above go routine.
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("an error occurred on server shutdown: %w", err)
	}

	// Here we return the result of our error channel from earlier. If there was no error then we will receive nil
	// otherwise we will receive the specified error.
	return <-errChan
}

// ServeHTTP requests via the internal router.
// This is what allows us to use our Server struct as the http.Server Handler in the Start method.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// RegisterHandlers with the router. This allows a Handler to define their route with the router.
func (s *Server) RegisterHandlers(handlers ...Handler) {
	for _, h := range handlers {
		h.AddRoute(s.router, s.environment)
	}
}
