package customers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	"github.com/nickbryan/go-template/service/app"
	"github.com/nickbryan/go-template/service/domain/customer"
	"github.com/nickbryan/go-template/service/transport/rest"
)

var errUserExists = errors.New("customers already exists with the given username")

type usernameUniqueRule struct {
	storage customer.Repository
	ctx     context.Context
}

func (r usernameUniqueRule) Validate(value interface{}) error {
	c, err := r.storage.FindByUsername(r.ctx, value.(string))
	if err != nil {
		return validation.NewInternalError(err)
	}

	if c != nil {
		return errUserExists
	}

	return nil
}

// NewCreateHandler creates a new handler for creating customers.
func NewCreateHandler(repo customer.Repository) rest.Handler {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	const (
		minPassLen = 6
		maxPassLen = 256 // Stops ddos through long password hashing
	)

	return rest.Handler{
		Route: func(r *mux.Route) {
			r.Path("/customers").Methods(http.MethodPost)
		},
		Func: func(w rest.Responder, r rest.Request) {
			var req request

			if err := r.Decode(&req); err != nil {
				w.RespondError(http.StatusBadRequest, err)

				return
			}

			if errs := app.Validate(&req,
				validation.Field(&req.Username, validation.Required, is.Email, usernameUniqueRule{repo, r.Context()}),
				validation.Field(&req.Password, validation.Required, validation.Length(minPassLen, maxPassLen)),
			); errs != nil {
				w.RespondValidationFailed(errs)

				return
			}

			cust, err := customer.New(req.Username, req.Password)
			if err != nil && errors.Is(err, customer.ErrGeneratePassword) {
				// TODO: handle error
			}

			if err := repo.Add(r.Context(), cust); err != nil {
				// TODO: handle error
			}

			w.WriteHeader(http.StatusCreated)
		},
	}
}
