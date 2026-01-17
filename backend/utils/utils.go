package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/andriyg76/glog"
	"net/http"
	"reflect"
	"runtime/debug"
)

func LogAndWriteHTTPError(r *http.Request, w http.ResponseWriter, statusCode int, err error, message string, a ...interface{}) {
	message2 := fmt.Sprintf(message, a...)
	logMessage := message2
	if err != nil {
		logMessage = fmt.Sprintf("%s: %v", message2, err)
	}

	requestInfo := "request=<nil>"
	requestIDInfo := "request_id=<none>"
	userInfo := "user_code=<anonymous>"
	if r != nil {
		if r.URL != nil {
			requestInfo = fmt.Sprintf("%s %s", r.Method, r.URL.String())
		} else {
			requestInfo = fmt.Sprintf("%s <nil-url>", r.Method)
		}

		if requestID := requestIDFromRequest(r); requestID != "" {
			requestIDInfo = fmt.Sprintf("request_id=%s", requestID)
		}

		if userCode := userCodeFromContext(r.Context()); userCode != "" {
			userInfo = fmt.Sprintf("user_code=%s", userCode)
		}
	}

	stack := string(debug.Stack())
	_ = glog.Error("%s | status=%d | %s | %s | %s | stack=%s", logMessage, statusCode, requestInfo, requestIDInfo, userInfo, stack)
	http.Error(w, message2, statusCode)
}

func requestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if requestID := r.Header.Get("X-Request-Id"); requestID != "" {
		return requestID
	}
	return r.Header.Get("X-Request-ID")
}

func userCodeFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	user := ctx.Value("user")
	if user == nil {
		return ""
	}

	value := reflect.ValueOf(user)
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return ""
	}
	field := value.FieldByName("Code")
	if !field.IsValid() || field.Kind() != reflect.String {
		return ""
	}
	return field.String()
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
