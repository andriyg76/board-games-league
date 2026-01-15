package utils

import (
	"errors"
	"math/big"

	"net/http"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// encodeBase58 encodes a byte slice to a base58 string
func encodeBase58(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Convert bytes to big.Int
	num := new(big.Int).SetBytes(data)
	zero := big.NewInt(0)
	base := big.NewInt(58)

	// Count leading zeros
	leadingZeros := 0
	for leadingZeros < len(data) && data[leadingZeros] == 0 {
		leadingZeros++
	}

	// Encode
	var result []byte
	for num.Cmp(zero) > 0 {
		mod := new(big.Int)
		num.DivMod(num, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	// Add leading zeros
	for i := 0; i < leadingZeros; i++ {
		result = append(result, base58Alphabet[0])
	}

	// Reverse the result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// decodeBase58 decodes a base58 string to a byte slice
func decodeBase58(s string) ([]byte, error) {
	if len(s) == 0 {
		return nil, errors.New("empty base58 string")
	}

	// Create reverse alphabet map
	alphabetMap := make(map[byte]int64)
	for i, char := range base58Alphabet {
		alphabetMap[byte(char)] = int64(i)
	}

	// Decode
	num := big.NewInt(0)
	base := big.NewInt(58)

	for _, char := range []byte(s) {
		value, ok := alphabetMap[char]
		if !ok {
			return nil, errors.New("invalid base58 character")
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(value))
	}

	// Convert to bytes
	bytes := num.Bytes()

	// Handle leading zeros
	leadingZeros := 0
	for leadingZeros < len(s) && s[leadingZeros] == base58Alphabet[0] {
		leadingZeros++
	}

	// Add leading zeros back
	if leadingZeros > 0 {
		result := make([]byte, leadingZeros+len(bytes))
		copy(result[leadingZeros:], bytes)
		return result, nil
	}

	return bytes, nil
}

// IdToCode converts a MongoDB ObjectID to a base58 encoded string.
// Base58 is URL-safe and does not require URL encoding.
//
// Parameters:
//   - id: primitive.ObjectID to be converted
//
// Returns:
//   - string: base58 encoded representation of the ID
func IdToCode(id primitive.ObjectID) string {
	return encodeBase58(id[:])
}

// CodeToID converts a base58 encoded string back to a MongoDB ObjectID.
//
// Parameters:
//   - code: base58 encoded string to convert
//
// Returns:
//   - primitive.ObjectID: the decoded MongoDB ObjectID
//   - error: if the code is invalid or cannot be decoded
func CodeToID(code string) (primitive.ObjectID, error) {
	if code == "" {
		return primitive.NilObjectID, errors.New("empty code")
	}

	bytes, err := decodeBase58(code)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if len(bytes) != 12 {
		return primitive.NilObjectID, errors.New("invalid code format: expected 12 bytes")
	}

	var id primitive.ObjectID
	copy(id[:], bytes)
	return id, nil
}

// GetIDFromChiURL extracts a MongoDB ObjectID from a Chi URL parameter.
// It handles base58 decoding of the parameter value.
//
// Parameters:
//   - r: HTTP request containing the URL parameters
//   - codeParam: name of the URL parameter containing the base58 encoded ID
//
// Returns:
//   - primitive.ObjectID: the decoded MongoDB ObjectID
//   - error: if the parameter is invalid or cannot be decoded
func GetIDFromChiURL(r *http.Request, codeParam string) (primitive.ObjectID, error) {
	code := chi.URLParam(r, codeParam)
	return CodeToID(code)
}
