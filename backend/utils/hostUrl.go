package utils

import (
	"github.com/andriyg76/glog"
	"net/http"
	"os"
	"strings"
	"sync"
)

var config = struct {
	hostUrl string
}{
	hostUrl: os.Getenv("HOST_URL"),
}

var initOnce sync.Once
var hostUrl string

func GetHostUrl(r *http.Request) string {
	initOnce.Do(func() {
		source := ""
		if config.hostUrl != "" {
			source = "config"
			hostUrl = config.hostUrl
		} else if r.Header.Get("Origin") != "" {
			source = "origin"
			hostUrl = r.Header.Get("Origin")
		} else if r.Header.Get("X-Forwarded-Proto") != "" && r.Header.Get("X-Forwarded-Host") != "" {
			source = "proxy header"
			scheme := r.Header.Get("X-Forwarded-Proto")
			host := r.Header.Get("X-Forwarded-Host")
			hostUrl = scheme + "://" + host
		} else {
			source = "request attributes"
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			hostUrl = scheme + "://" + r.Host
		}
		hostUrl = strings.TrimSuffix(hostUrl, "/")
		glog.Info("Host url resolved/configured via %s to: %s", source, hostUrl)
	})
	return hostUrl
}
