package models

import (
	"github.com/andriyg76/bgl/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IdAndCode represents a pair of MongoDB ObjectID and its string code representation
type IdAndCode struct {
	ID   primitive.ObjectID
	Code string
}

// Stringify returns the code string representation
func (ic *IdAndCode) Stringify() string {
	return ic.Code
}

// NewIdAndCode creates a new IdAndCode from ObjectID
func NewIdAndCode(id primitive.ObjectID) *IdAndCode {
	return &IdAndCode{
		ID:   id,
		Code: utils.IdToCode(id),
	}
}

// NewIdAndCodeFromCode creates a new IdAndCode from code string
func NewIdAndCodeFromCode(code string) (*IdAndCode, error) {
	id, err := utils.CodeToID(code)
	if err != nil {
		return nil, err
	}
	return &IdAndCode{
		ID:   id,
		Code: code,
	}, nil
}

