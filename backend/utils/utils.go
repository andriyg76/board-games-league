package utils

import (
	"crypto/rand"
	"fmt"
	"github.com/andriyg76/glog"
	"net/http"
)

func LogAndWriteHTTPError(w http.ResponseWriter, statusCode int, err error, message string, a ...interface{}) {
	message2 := fmt.Sprintf(message, a...)
	a = append(a, err)
	_ = glog.Error(message+": %v", a...)
	http.Error(w, message2, statusCode)
}

func GenerateRandomKey(length int) []byte {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		glog.Fatal("Failed to generate random key: %v", err)
	}
	return key
}
