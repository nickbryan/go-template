package rest_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/nickbryan/go-template/service/app"
	"github.com/nickbryan/go-template/service/transport/rest"
	"github.com/nickbryan/go-template/service/transport/rest/resttest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestPanicIsRecovered(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		handlerFunc rest.ServiceFunc
		assert      func(resp *httptest.ResponseRecorder, logs *observer.ObservedLogs)
	}{
		{
			name: "with string panic",
			handlerFunc: func(w rest.Responder, r rest.Request) {
				panic("something really bad happened")
			},
			assert: func(resp *httptest.ResponseRecorder, logs *observer.ObservedLogs) {
				assert.Equal(t, http.StatusInternalServerError, resp.Code)
				assert.Equal(t, 1, logs.Len(), logs.All())
				assert.Equal(t, "application panicked", logs.All()[0].Message)
				assert.Equal(t, "error", logs.All()[0].Context[0].Key)
				assert.Equal(t, "something really bad happened", logs.All()[0].Context[0].Interface.(error).Error())
			},
		},
		{
			name: "with error panic",
			handlerFunc: func(w rest.Responder, r rest.Request) {
				panic(errors.New("some really bad error"))
			},
			assert: func(resp *httptest.ResponseRecorder, logs *observer.ObservedLogs) {
				assert.Equal(t, http.StatusInternalServerError, resp.Code)
				assert.Equal(t, 1, logs.Len(), logs.All())
				assert.Equal(t, "error", logs.All()[0].Context[0].Key)
				assert.Equal(t, "some really bad error", logs.All()[0].Context[0].Interface.(error).Error())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			core, logs := observer.New(zap.DebugLevel)
			logger := zap.New(core)

			testEnv := app.NewTestEnvironmentWithLogger(t, logger, false)

			_, resp := resttest.Request(
				t,
				http.MethodGet,
				"/test-panic-recovery",
				rest.Handler{
					Route: func(r *mux.Route) {
						r.Path("/test-panic-recovery").Methods(http.MethodGet)
					},
					Func: tc.handlerFunc,
				},
				testEnv,
			)

			if err := logger.Sync(); err != nil {
				t.Fatalf("logger sync failed: %v", err)
			}

			tc.assert(resp, logs)
		})
	}
}
