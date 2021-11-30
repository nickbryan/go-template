package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Request wraps the http.Request so that we can add custom methods.
type Request struct {
	*http.Request
}

// Decode de-serialises the JSON body of the request into the passed destination object.
func (r Request) Decode(dest interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dest)
	if err != nil {
		return fmt.Errorf("unable to decode request: %w", err)
	}

	return nil
}
