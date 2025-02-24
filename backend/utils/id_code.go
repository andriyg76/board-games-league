package utils

import (
	"encoding/base64"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IdToCode converts a MongoDB ObjectID to a string code.
func IdToCode(id primitive.ObjectID) string {
	return base64.StdEncoding.EncodeToString(id[:])
}

// CodeToID converts a UUEncoded string code to a MongoDB ObjectID.
func CodeToID(code string) (primitive.ObjectID, error) {
	data, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if len(data) != 12 {
		return primitive.NilObjectID, errors.New("invalid code format")
	}
	var id primitive.ObjectID
	copy(id[:], data)
	return id, nil
}
