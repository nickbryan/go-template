package rest

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	"github.com/nickbryan/go-template/service/app"
	"go.uber.org/zap"
)

// Responder wraps a http.ResponseWriter so that we can add our own helper methods.
type Responder interface {
	http.ResponseWriter

	// Respond will write an encoded json string to the http.ResponseWriter if data is not nil.
	// A http.StatusInternalServerError will be written if setting of the json values fails.
	Respond(status int, data interface{})

	// RespondError will write a formatted list of validation errors to the http.ResponseWriter.
	// A http.StatusInternalServerError will be written if setting of the json values fails.
	RespondError(status int, err error)

	// RespondValidationFailed will write the give error message to the http.ResponseWriter as formatted JSON.
	// A http.StatusInternalServerError will be written if setting of the json values fails.
	RespondValidationFailed(errors validation.Errors)
}

// ServiceFunc is our handler function definition, so that handlers can access the configured Responder and Request.
type ServiceFunc func(w Responder, r Request)

// Handler is responsible for defining a HTTP request route and corresponding handler.
type Handler struct {
	// Route receives a route to modify, like adding path, methods, etc.
	Route func(r *mux.Route)

	// Middleware allows wrapping the Func in middleware handlers.
	Middleware func(next ServiceFunc) ServiceFunc

	// Func will be registered with the router.
	Func ServiceFunc
}

// AddRoute adds the handler's route the to the router.
func (h Handler) AddRoute(r *mux.Router, e *app.Environment) {
	fnc := h.Func

	if h.Middleware != nil {
		fnc = h.Middleware(fnc)
	}

	h.Route(r.NewRoute().HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recoverPanicMiddleware(fnc, e)(
			&responder{
				ResponseWriter: w,
				logger:         e.Logger(),
			},
			Request{r},
		)
	}))
}

// ErrUnknown will be logged when the panic recovery has an unknown type.
var ErrUnknown = errors.New("unknown error")

func recoverPanicMiddleware(next ServiceFunc, e *app.Environment) ServiceFunc {
	return func(w Responder, r Request) {
		defer func() {
			if rec := recover(); rec != nil {
				var err error

				switch e := rec.(type) {
				case string:
					err = errors.New(e)
				case error:
					err = e
				default:
					err = ErrUnknown
				}

				w.WriteHeader(http.StatusInternalServerError)
				e.Logger().Error("application panicked", zap.Error(err))
			}
		}()

		next(w, r)
	}
}
