package customers_test

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/nickbryan/go-template/service/app"
	"github.com/nickbryan/go-template/service/infrastructure/postgres"
	"github.com/nickbryan/go-template/service/infrastructure/postgres/postgrestest"
	"github.com/nickbryan/go-template/service/transport/rest/customers"
	"github.com/nickbryan/go-template/service/transport/rest/resttest"
)

func genLongPassword(t *testing.T, ln uint) string {
	t.Helper()

	letter := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, ln)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))] //nolint:gosec
	}

	return string(b)
}

func TestCustomerCreateHandler(t *testing.T) {
	t.Parallel()

	type payload map[string]interface{}

	tests := []struct {
		name   string
		url    string
		method string
		input  payload
		setup  func(db *app.DB)
		assert func(data *gabs.Container, resp *httptest.ResponseRecorder, db *app.DB)
	}{
		{
			name:   "create with empty payload returns error",
			url:    "/customers",
			method: http.MethodPost,
			input:  payload{},
			assert: func(data *gabs.Container, resp *httptest.ResponseRecorder, _ *app.DB) {
				assert.Equal(t, http.StatusBadRequest, resp.Code)
				assert.Equal(
					t,
					"cannot be blank",
					data.Path("error.validation_errors.username").Data().(string),
				)
				assert.Equal(
					t,
					"cannot be blank",
					data.Path("error.validation_errors.password").Data().(string),
				)
			},
		},
		{
			name:   "create with empty username returns error",
			url:    "/customers",
			method: http.MethodPost,
			input:  payload{"username": "", "password": "test123"},
			assert: func(data *gabs.Container, resp *httptest.ResponseRecorder, _ *app.DB) {
				assert.Equal(t, http.StatusBadRequest, resp.Code)
				assert.Equal(
					t,
					"cannot be blank",
					data.Path("error.validation_errors.username").Data().(string),
					data,
				)
			},
		},
		{
			name:   "create with empty password returns error",
			url:    "/customers",
			method: http.MethodPost,
			input:  payload{"username": "test@example.org", "password": ""},
			assert: func(data *gabs.Container, resp *httptest.ResponseRecorder, _ *app.DB) {
				assert.Equal(t, http.StatusBadRequest, resp.Code)
				assert.Equal(
					t,
					"cannot be blank",
					data.Path("error.validation_errors.password").Data().(string),
					data,
				)
			},
		},
		{
			name:   "create with invalid email for username returns error",
			url:    "/customers",
			method: http.MethodPost,
			input:  payload{"username": "this is not an email", "password": "test123"},
			assert: func(data *gabs.Container, resp *httptest.ResponseRecorder, _ *app.DB) {
				assert.Equal(t, http.StatusBadRequest, resp.Code)
				assert.Equal(
					t,
					"must be a valid email address",
					data.Path("error.validation_errors.username").Data().(string),
				)
			},
		},
		{
			name:   "create with password too short returns error",
			url:    "/customers",
			method: http.MethodPost,
			input:  payload{"username": "test@example.org", "password": "123"},
			assert: func(data *gabs.Container, resp *httptest.ResponseRecorder, _ *app.DB) {
				assert.Equal(t, http.StatusBadRequest, resp.Code)
				assert.Equal(
					t,
					"the length must be between 6 and 256",
					data.Path("error.validation_errors.password").Data().(string),
				)
			},
		},
		{
			name:   "create with password too long returns error",
			url:    "/customers",
			method: http.MethodPost,
			input:  payload{"username": "test@example.org", "password": genLongPassword(t, 257)},
			assert: func(data *gabs.Container, resp *httptest.ResponseRecorder, _ *app.DB) {
				assert.Equal(t, http.StatusBadRequest, resp.Code)
				assert.Equal(
					t,
					"the length must be between 6 and 256",
					data.Path("error.validation_errors.password").Data().(string),
				)
			},
		},
		{
			name:   "customers can be created",
			url:    "/customers",
			method: http.MethodPost,
			input:  payload{"username": "test@example.org", "password": "Sup3rS3cr3t"},
			assert: func(data *gabs.Container, resp *httptest.ResponseRecorder, db *app.DB) {
				assert.Equal(t, http.StatusCreated, resp.Code)
				postgrestest.AssertDatabaseHas(t, db, "customers", map[string]string{
					"username": "test@example.org",
				})
			},
		},
		{
			name:   "trying to create user that already exists returns an error",
			url:    "/customers",
			method: http.MethodPost,
			input:  payload{"username": "existing@example.org", "password": "Sup3rS3cr3t"},
			setup: func(db *app.DB) {
				query := db.QB().
					Insert("customers").
					Columns("uuid", "username", "password", "created_at", "updated_at").
					Values(uuid.New(), "existing@example.org", "abc123", time.Now(), time.Now())

				sql, args, err := query.ToSql()
				if err != nil {
					t.Fatalf("unable to convert query to SQL: %v", err)
				}

				_, err = db.Conn().Exec(context.Background(), sql, args...)
				if err != nil {
					t.Fatalf("unable to execute query: %v", err)
				}
			},
			assert: func(data *gabs.Container, resp *httptest.ResponseRecorder, db *app.DB) {
				assert.Equal(t, http.StatusBadRequest, resp.Code)
				assert.Equal(
					t,
					"customers already exists with the given username",
					data.Path("error.validation_errors.username").Data().(string),
				)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testEnv := app.NewTestEnvironment(t, true)

			if tc.setup != nil {
				tc.setup(testEnv.DB())
			}

			data, resp := resttest.RequestWithData(
				t,
				tc.method,
				tc.url,
				customers.NewCreateHandler(postgres.NewCustomerRepository(testEnv.DB())),
				tc.input,
				testEnv,
			)

			tc.assert(data, resp, testEnv.DB())
		})
	}
}
