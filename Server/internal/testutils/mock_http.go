package testutils

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
)

type HeaderEntry struct {
	Name string
	Value string
}



func HostEndpoint(f gin.HandlerFunc, body string, headers []HeaderEntry) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/", f)
	for _, h := range headers {
		req.Header.Set(h.Name, h.Value)
	}

	r.ServeHTTP(w, req)

	return w
}
func HostGetEndpoint(f gin.HandlerFunc, url string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	r := gin.Default()
	r.GET("/", f)

	r.ServeHTTP(w, req)

	return w
}