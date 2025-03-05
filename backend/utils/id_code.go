package utils

import (
	"encoding/base64"
	"errors"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/url"
)

// IdToCode converts a MongoDB ObjectID to a URL-safe base64 encoded string.
// The resulting string is safe to use in URLs as it's URL-encoded.
//
// Parameters:
//   - id: primitive.ObjectID to be converted
//
// Returns:
//   - string: URL-safe base64 encoded representation of the ID
func IdToCode(id primitive.ObjectID) string {
	return url.QueryEscape(base64.StdEncoding.EncodeToString(id[:]))
}

// CodeToID converts a URL-safe base64 encoded string back to a MongoDB ObjectID.
// It first URL-decodes the string and then decodes the base64 representation.
//
// Parameters:
//   - code: URL-safe base64 encoded string to convert
//
// Returns:
//   - primitive.ObjectID: the decoded MongoDB ObjectID
//   - error: if the code is invalid or cannot be decoded
func CodeToID(code string) (primitive.ObjectID, error) {
	unescapedCode, err := url.QueryUnescape(code)
	if err != nil {
		return primitive.NilObjectID, err
	}

	bytes, err := base64.StdEncoding.DecodeString(unescapedCode)
	if err != nil {
		return primitive.NilObjectID, err
	} else if len(bytes) != 12 {
		return primitive.NilObjectID, errors.New("invalid code format")
	}

	var id primitive.ObjectID
	copy(id[:], bytes)
	return id, nil
}

// GetIDFromChiURL extracts a MongoDB ObjectID from a Chi URL parameter.
// It handles URL-decoding and base64 decoding of the parameter value.
//
// Parameters:
//   - r: HTTP request containing the URL parameters
//   - codeParam: name of the URL parameter containing the encoded ID
//
// Returns:
//   - primitive.ObjectID: the decoded MongoDB ObjectID
//   - error: if the parameter is invalid or cannot be decoded
func GetIDFromChiURL(r *http.Request, codeParam string) (primitive.ObjectID, error) {
	code := chi.URLParam(r, codeParam)
	return CodeToID(code)
}
