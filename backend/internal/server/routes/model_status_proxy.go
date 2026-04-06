package routes

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterModelStatusProxyRoutes(r *gin.Engine) {
	if strings.ToLower(strings.TrimSpace(os.Getenv("MODEL_STATUS_ENABLED"))) != "true" {
		return
	}

	targetURL := strings.TrimSpace(os.Getenv("MODEL_STATUS_INTERNAL_URL"))
	if targetURL == "" {
		targetURL = "http://127.0.0.1:3001"
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if req.Header.Get("X-Forwarded-Proto") == "" {
			if req.TLS != nil {
				req.Header.Set("X-Forwarded-Proto", "https")
			} else {
				req.Header.Set("X-Forwarded-Proto", "http")
			}
		}
	}
	proxy.ErrorHandler = func(cw http.ResponseWriter, _ *http.Request, _ error) {
		cw.Header().Set("Content-Type", "application/json; charset=utf-8")
		cw.WriteHeader(http.StatusBadGateway)
		_, _ = cw.Write([]byte(`{"error":"Model Status sidecar unavailable"}`))
	}

	handler := func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}

	r.Any("/status", handler)
	r.Any("/status/*path", handler)
}
