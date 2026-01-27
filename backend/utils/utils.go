package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"reflect"

	"github.com/andriyg76/glog"
	"github.com/andriyg76/hexerr"
	"github.com/go-chi/chi/v5/middleware"
)

func LogAndWriteHTTPError(r *http.Request, w http.ResponseWriter, statusCode int, err error, message string, a ...interface{}) {
	message2 := fmt.Sprintf(message, a...)

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

		if userCode := UserCodeFromContext(r.Context()); userCode != "" {
			userInfo = fmt.Sprintf("user_code=%s", userCode)
		}
	}

	var errDetail string
	if err != nil {
		err = hexerr.Wrap(err, message2)
		errDetail = fmt.Sprintf("%+v", err)
	}

	_ = glog.Error("%s | status=%d | %s | %s | %s | error=%s", message2, statusCode, requestInfo, requestIDInfo, userInfo, errDetail)
	http.Error(w, message2, statusCode)
}

func requestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	// First try to get from chi middleware context (most reliable)
	if r.Context() != nil {
		if requestID := middleware.GetReqID(r.Context()); requestID != "" {
			return requestID
		}
	}
	// Fallback to headers (incoming request might have X-Request-Id)
	if requestID := r.Header.Get("X-Request-Id"); requestID != "" {
		return requestID
	}
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	return ""
}

func UserCodeFromContext(ctx context.Context) string {
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
