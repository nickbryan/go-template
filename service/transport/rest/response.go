package rest

import (
	"encoding/json"
	"net/http"

	"github.com/Jeffail/gabs"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"go.uber.org/zap"
)

type responder struct {
	http.ResponseWriter
	logger *zap.Logger
}

func (r *responder) Respond(status int, data interface{}) {
	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(r).Encode(data); err != nil {
			r.logger.Error("unable to encode response", zap.Error(err))
			r.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (r *responder) RespondError(status int, err error) {
	body := gabs.New()

	r.logger.Error("responding application error", zap.Error(err), zap.Int("status_code", status))

	if _, err := body.SetP(err.Error(), "error.message"); err != nil {
		r.logger.Error("unable to set json body when responding error message", zap.Error(err))
		r.WriteHeader(http.StatusInternalServerError)

		return
	}

	r.Respond(status, body.Data())
}

func (r *responder) RespondValidationFailed(errors validation.Errors) {
	body := gabs.New()

	if _, err := body.SetP("request contains invalid fields", "error.message"); err != nil {
		r.logger.Error("unable to set json body when setting validation failed message", zap.Error(err))
		r.WriteHeader(http.StatusInternalServerError)

		return
	}

	if _, err := body.SetP(errors, "error.validation_errors"); err != nil {
		r.logger.Error("unable to set json body when responding validation failed errors", zap.Error(err))
		r.WriteHeader(http.StatusInternalServerError)

		return
	}

	r.Respond(http.StatusBadRequest, body.Data())
}
