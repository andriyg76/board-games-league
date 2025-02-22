package utils

import (
	"encoding/base64"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ConvertIDToCode converts a MongoDB ObjectID to a string code.
func ConvertIDToCode(id primitive.ObjectID) string {
	return base64.StdEncoding.EncodeToString(id[:])
}

// ConvertCodeToID converts a UUEncoded string code to a MongoDB ObjectID.
func ConvertCodeToID(code string) (primitive.ObjectID, error) {
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

type IdCode struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	Code string             `bson:"-" json:"code"`
}

func (r IdCode) ConvertIDToCode() {
	r.Code = ConvertIDToCode(r.ID)
}

func (r IdCode) ConvertCodeToID() error {
	if id, err := ConvertCodeToID(r.Code); err != nil {
		return err
	} else {
		r.ID = id
		return err
	}
}
