package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/andriyg76/glog"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
)

func LogAndWriteHTTPError(r *http.Request, w http.ResponseWriter, statusCode int, err error, message string, a ...interface{}) {
	message2 := fmt.Sprintf(message, a...)
	logMessage := message2

	requestInfo := "request=<nil>"
	requestIDInfo := "request_id=<none>"
	userInfo := "user_code=<anonymous>"
	errorInfo := ""
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

	if err != nil {
		errorInfo = fmt.Sprintf("error: %v", err)
	}

	stack := filterStackTrace(string(debug.Stack()))
	_ = glog.Error("%s | %s | status=%d | %s | %s | %s | stack=%s", logMessage, errorInfo, statusCode, requestInfo, requestIDInfo, userInfo, stack)
	http.Error(w, message2, statusCode)
}

func filterStackTrace(stack string) string {
	const bglPackage = "github.com/andriyg76/bgl"
	lines := strings.Split(stack, "\n")
	var filtered []string
	var ourIndices []int

	// Знайти індекси рядків з нашим пакетом
	for i, line := range lines {
		if strings.Contains(line, bglPackage) {
			ourIndices = append(ourIndices, i)
		}
	}

	if len(ourIndices) == 0 {
		return stack
	}

	// Знайти групи послідовних індексів
	var groups [][]int
	var currentGroup []int

	for _, idx := range ourIndices {
		if len(currentGroup) == 0 || idx == currentGroup[len(currentGroup)-1]+1 {
			currentGroup = append(currentGroup, idx)
		} else {
			if len(currentGroup) > 0 {
				groups = append(groups, currentGroup)
			}
			currentGroup = []int{idx}
		}
	}
	// Додати останню групу
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	// Для кожної групи додати контекст (один рядок до і після)
	var addedIndices = make(map[int]bool)
	for _, group := range groups {
		firstIdx := group[0]
		lastIdx := group[len(group)-1]

		// Додати один рядок перед групою (якщо є і не з нашої бібліотеки)
		if firstIdx > 0 && !strings.Contains(lines[firstIdx-1], bglPackage) {
			if !addedIndices[firstIdx-1] {
				filtered = append(filtered, lines[firstIdx-1])
				addedIndices[firstIdx-1] = true
			}
		}

		// Додати всі рядки групи
		for _, idx := range group {
			if !addedIndices[idx] {
				filtered = append(filtered, lines[idx])
				addedIndices[idx] = true
			}
		}

		// Додати один рядок після групи (якщо є і не з нашої бібліотеки)
		if lastIdx < len(lines)-1 && !strings.Contains(lines[lastIdx+1], bglPackage) {
			if !addedIndices[lastIdx+1] {
				filtered = append(filtered, lines[lastIdx+1])
				addedIndices[lastIdx+1] = true
			}
		}
	}

	return strings.Join(filtered, "\n")
}

func requestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	// First check headers (incoming request might have X-Request-Id)
	if requestID := r.Header.Get("X-Request-Id"); requestID != "" {
		return requestID
	}
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	// Then check context (chi middleware.RequestID stores it there)
	if r.Context() != nil {
		if requestID := r.Context().Value("requestID"); requestID != nil {
			if id, ok := requestID.(string); ok && id != "" {
				return id
			}
		}
	}
	return ""
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
