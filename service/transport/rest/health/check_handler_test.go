package health_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nickbryan/go-template/service/app"
	"github.com/nickbryan/go-template/service/transport/rest/health"
	"github.com/nickbryan/go-template/service/transport/rest/resttest"
)

func TestHealthCheckHandler(t *testing.T) {
	t.Parallel()
	data, resp := resttest.Request(
		t,
		http.MethodGet,
		"/health",
		health.NewCheckHandler(),
		app.NewTestEnvironment(t, false),
	)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, data.Path("status").Data().(string), "ok")
}
