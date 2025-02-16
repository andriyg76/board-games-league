package utils

import (
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
