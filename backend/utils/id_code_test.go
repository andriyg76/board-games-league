package utils

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestIdToCode(t *testing.T) {
	tests := []struct {
		name string
		id   primitive.ObjectID
	}{
		{
			name: "regular ObjectID",
			id:   primitive.NewObjectID(),
		},
		{
			name: "zero ObjectID",
			id:   primitive.NilObjectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := IdToCode(tt.id)
			assert.NotEmpty(t, code)

			// Verify roundtrip
			decodedID, err := CodeToID(code)
			assert.NoError(t, err)
			assert.Equal(t, tt.id, decodedID)
		})
	}
}

func TestCodeToID(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
		id      primitive.ObjectID
	}{
		{
			name:    "valid code",
			code:    IdToCode(primitive.NewObjectID()),
			wantErr: false,
		},
		{
			name:    "URL encoded code",
			code:    "Z71x74W1r/XKUsjq",
			wantErr: false,
			id:      primitive.ObjectID{0x67, 0xbd, 0x71, 0xef, 0x85, 0xb5, 0xaf, 0xf5, 0xca, 0x52, 0xc8, 0xea},
		},
		{
			name:    "invalid base64",
			code:    "invalid-base64",
			wantErr: true,
		},
		{
			name:    "empty string",
			code:    "",
			wantErr: true,
		},
		{
			name:    "wrong length",
			code:    "AA==",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := CodeToID(tt.code)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.id, id)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetIDFromChiURL(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		paramVal  string
		wantErr   bool
	}{
		{
			name:      "valid parameter",
			paramName: "code",
			paramVal:  IdToCode(primitive.NewObjectID()),
			wantErr:   false,
		},
		{
			name:      "missing parameter",
			paramName: "code",
			paramVal:  "",
			wantErr:   true,
		},
		{
			name:      "invalid code",
			paramName: "code",
			paramVal:  "invalid-code",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Chi router and context
			r := chi.NewRouter()
			r.Get("/{code}", func(w http.ResponseWriter, r *http.Request) {
				id, err := GetIDFromChiURL(r, tt.paramName)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Equal(t, primitive.NilObjectID, id)
				} else {
					assert.NoError(t, err)
					assert.NotEqual(t, primitive.NilObjectID, id)
				}
			})

			// Create test request
			req := httptest.NewRequest("GET", "/"+tt.paramVal, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		})
	}
}
