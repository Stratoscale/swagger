package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"github.com/Stratoscale/swagger/example/auth"
	"github.com/Stratoscale/swagger/example/models"
	"github.com/Stratoscale/swagger/example/restapi"
	"github.com/Stratoscale/swagger/example/restapi/operations/pet"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const target = "http://petstore.org/api"

type mocks struct {
	pet   restapi.MockPetAPI
	store restapi.MockStoreAPI
}

func (m *mocks) assertExpectations(t *testing.T) {
	m.pet.AssertExpectations(t)
	m.store.AssertExpectations(t)
}

func TestHTTPHandler(t *testing.T) {
	t.Parallel()

	// declare the test cases
	tests := []struct {
		name     string
		req      *http.Request
		cookie   string
		prepare  func(*testing.T, *mocks)
		wantCode int
		wantBody []byte
	}{
		{
			name:     "get pet by anonymous should be unauthorized",
			req:      httptest.NewRequest(http.MethodGet, target+"/pets/1", nil),
			wantCode: http.StatusUnauthorized,
		},
		{
			name:   "get pet by a member",
			req:    httptest.NewRequest(http.MethodGet, target+"/pets/1", nil),
			cookie: `{"id":1,"role":"member"}`,
			prepare: func(t *testing.T, m *mocks) {
				m.pet.On("PetGet", mock.Anything, mock.Anything).
					Return(&pet.PetGetOK{Payload: &models.Pet{ID: 1, Name: swag.String("kitty")}}).
					Once()
			},
			wantCode: http.StatusOK,
			wantBody: []byte(`{"id":1,"name":"kitty","tags":null}`),
		},
		{
			name:     "pet create by anonymous should be unauthorized",
			req:      httptest.NewRequest(http.MethodPost, target+"/pets", bytes.NewBufferString(`{"name":"kitty"}`)),
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "pet create by a member should be forbidden",
			req:      httptest.NewRequest(http.MethodPost, target+"/pets", bytes.NewBufferString(`{"name":"kitty"}`)),
			cookie:   `{"id":1,"role":"member"}`,
			wantCode: http.StatusForbidden,
		},
		{
			name:   "pet create by an admin",
			req:    httptest.NewRequest(http.MethodPost, target+"/pets", bytes.NewBufferString(`{"name":"kitty"}`)),
			cookie: `{"id":1,"role":"admin"}`,
			prepare: func(t *testing.T, m *mocks) {
				m.pet.On("PetCreate", mock.Anything, mock.Anything).
					Return(&pet.PetCreateCreated{Payload: &models.Pet{ID: 1, Name: swag.String("kitty")}}).
					Once()
			},
			wantCode: http.StatusCreated,
			wantBody: []byte(`{"id":1,"name":"kitty","tags":null}`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				resp  = httptest.NewRecorder()
				mocks mocks
			)

			h, err := restapi.Handler(restapi.Config{
				PetAPI:     &mocks.pet,
				StoreAPI:   &mocks.store,
				AuthToken:  auth.Token,
				Authorizer: auth.Request,
				Logger:     t.Logf,
			})
			require.Nil(t, err)

			tt.req.Header.Set("Content-Type", "application/json")
			tt.req.Header.Set("Cookie", tt.cookie)

			// prepare mocks
			if tt.prepare != nil {
				tt.prepare(t, &mocks)
			}

			h.ServeHTTP(resp, tt.req)

			t.Logf("Got response for request %s %s: %d %s", tt.req.Method, tt.req.URL, resp.Code, resp.Body.String())

			// assert the response expectations
			assert.Equal(t, tt.wantCode, resp.Code)
			if tt.wantBody != nil {
				assert.JSONEq(t, string(tt.wantBody), resp.Body.String())
			}

			mocks.assertExpectations(t)
		})
	}
}
