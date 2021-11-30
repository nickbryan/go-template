package rest_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nickbryan/go-template/service/transport/rest"
)

func TestRequestDecode(t *testing.T) {
	t.Parallel()

	type response struct {
		FieldA string `json:"field_a"`
		FieldB int    `json:"field_b"`
	}

	t.Run("decodes valid json", func(t *testing.T) {
		t.Parallel()

		var resp response

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		t.Cleanup(cancel)

		r, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/",
			bytes.NewReader([]byte(`{"field_a": "a", "field_b": 123}`)),
		)
		if err != nil {
			t.Fatalf("unable to create request: %v", err)
		}

		w := rest.Request{Request: r}

		if err := w.Decode(&resp); err != nil {
			t.Fatalf("unable to decode request in test: %v", err)
		}

		assert.Equal(t, "a", resp.FieldA)
		assert.Equal(t, 123, resp.FieldB)
	})

	t.Run("errors on invalid json", func(t *testing.T) {
		t.Parallel()

		var resp response

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		t.Cleanup(cancel)

		r, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/",
			bytes.NewReader([]byte(`not valid json`)),
		)
		if err != nil {
			assert.FailNow(t, fmt.Sprintf("unable to create request: %v", err))
		}

		w := rest.Request{Request: r}

		err = w.Decode(&resp)
		if err == nil {
			t.Fatalf("expected Decode to return false to indicate that an error occurred")
		}

		assert.Contains(t, err.Error(), "unable to decode request: ")
	})
}
