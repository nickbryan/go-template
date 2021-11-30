package resttest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/nickbryan/go-template/service/app"
	"github.com/nickbryan/go-template/service/transport/rest"
)

// Request wraps RequestWithBody passing nil as the body.
func Request(
	t *testing.T,
	method, url string,
	handler rest.Handler,
	e *app.Environment,
) (*gabs.Container, *httptest.ResponseRecorder) {
	t.Helper()

	return RequestWithData(t, method, url, handler, nil, e)
}

// RequestWithData returns the decoded json response as ResponseData along with the httptest.ResponseRecorder.
// If any errors are returned during the execution of the request the test will me marked as failed and exit
// immediately.
func RequestWithData(
	t *testing.T,
	method, url string,
	handler rest.Handler,
	data interface{},
	e *app.Environment,
) (*gabs.Container, *httptest.ResponseRecorder) {
	t.Helper()

	body, err := json.Marshal(data)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("unable to marshal input data: %v", err))
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("unable to create request: %v", err))
	}

	resp := httptest.NewRecorder()

	r := mux.NewRouter()
	handler.AddRoute(r, e)
	r.ServeHTTP(resp, req)

	if resp.Body.Len() == 0 {
		return nil, resp
	}

	parsed, err := gabs.ParseJSON(resp.Body.Bytes())
	if err != nil {
		assert.FailNow(
			t,
			fmt.Sprintf(
				"unable to unmarshal response body: %v, body is: %v",
				err,
				resp.Body,
			),
		)
	}

	return parsed, resp
}
