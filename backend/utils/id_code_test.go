package utils

import (
	"encoding/base64"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestConvertIDToCode(t *testing.T) {
	id := primitive.NewObjectID()
	expectedCode := base64.StdEncoding.EncodeToString(id[:])

	code := ConvertIDToCode(id)
	if code != expectedCode {
		t.Fatalf("expected %v, got %v", expectedCode, code)
	}
}

func TestConvertCodeToID(t *testing.T) {
	id := primitive.NewObjectID()
	code := base64.StdEncoding.EncodeToString(id[:])
	expectedID := id

	decodedID, err := ConvertCodeToID(code)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if decodedID != expectedID {
		t.Fatalf("expected %v, got %v", expectedID, decodedID)
	}
}

func TestConvertCodeToID_InvalidCode(t *testing.T) {
	code := "invalid_code"
	_, err := ConvertCodeToID(code)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestConvertCodeToID_IncompleteCode(t *testing.T) {
	code := "short"
	_, err := ConvertCodeToID(code)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
