package auth

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
)

// MockExternalAuthProvider implements ExternalAuthProvider interface
type MockExternalAuthProvider struct {
	mock.Mock
}

func (m *MockExternalAuthProvider) BeginUserAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusPermanentRedirect)
	w.Header().Add("Location", fmt.Sprintf("http://google.com/auth?state=%s", r.URL.Query().Get("state")))
}

func (m *MockExternalAuthProvider) CompleteUserAuthHandler(_ http.ResponseWriter, r *http.Request) (ExternalUser, error) {
	args := m.Called(r.Context(), r)
	return args.Get(0).(ExternalUser), args.Error(1)
}

func (m *MockExternalAuthProvider) LogoutHandler(_ http.ResponseWriter, r *http.Request) error {
	args := m.Called(r.Context(), r)
	return args.Error(0)
}
