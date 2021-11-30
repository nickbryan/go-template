package rest_test

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/nickbryan/go-template/service/app"
	"github.com/nickbryan/go-template/service/transport/rest"
)

func testHandler(called *bool) rest.Handler {
	return rest.Handler{
		Route: func(r *mux.Route) {
			*called = true
		},
	}
}

func TestServer(t *testing.T) {
	t.Parallel()

	t.Run("can register handlers", func(t *testing.T) {
		t.Parallel()

		called := false

		testEnv := app.NewTestEnvironment(t, false)
		s := rest.NewServer(testEnv)
		s.RegisterHandlers(testHandler(&called))

		assert.True(t, called, "the RegisterRoutes method was not called on mock")
	})
}
