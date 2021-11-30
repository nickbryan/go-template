package health

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nickbryan/go-template/service/transport/rest"
)

// NewCheckHandler returns a handler for health checks.
// The handler returns status 200 and a `{"status":  "ok"}` payload.
func NewCheckHandler() rest.Handler {
	type response struct {
		Status string `json:"status"`
	}

	return rest.Handler{
		Route: func(r *mux.Route) {
			r.Path("/health").Methods(http.MethodGet)
		},
		Func: func(w rest.Responder, r rest.Request) {
			w.Respond(http.StatusOK, response{Status: "ok"})
		},
	}
}
