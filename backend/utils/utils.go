package utils

import (
	"crypto/rand"
	"fmt"
	"github.com/andriyg76/glog"
	"net/http"
	"runtime/debug"
)

func LogAndWriteHTTPError(r *http.Request, w http.ResponseWriter, statusCode int, err error, message string, a ...interface{}) {
	message2 := fmt.Sprintf(message, a...)
	logMessage := message2
	if err != nil {
		logMessage = fmt.Sprintf("%s: %v", message2, err)
	}

	requestInfo := "request=<nil>"
	userInfo := "user=<nil>"
	if r != nil {
		if r.URL != nil {
			requestInfo = fmt.Sprintf("%s %s", r.Method, r.URL.String())
		} else {
			requestInfo = fmt.Sprintf("%s <nil-url>", r.Method)
		}

		if user := r.Context().Value("user"); user != nil {
			userInfo = fmt.Sprintf("user=%+v", user)
		} else {
			userInfo = "user=<anonymous>"
		}
	}

	stack := string(debug.Stack())
	_ = glog.Error("%s | status=%d | %s | %s | stack=%s", logMessage, statusCode, requestInfo, userInfo, stack)
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
