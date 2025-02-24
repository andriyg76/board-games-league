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

func Map[U, V any](ts []U, f func(U) V) []V {
	us := make([]V, len(ts))
	for i, t := range ts {
		us[i] = f(t)
	}
	return us
}
