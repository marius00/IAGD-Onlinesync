package testutils

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type HeaderEntry struct {
	Name string
	Value string
}

func HostEndpoint(f gin.HandlerFunc, body string, headers []HeaderEntry) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/", f)
	for _, h := range headers {
		req.Header.Set(h.Name, h.Value)
	}

	r.ServeHTTP(w, req)

	return w
}


type DummyAuthorizer struct {}
func (*DummyAuthorizer) IsValid(email string, token string) (bool, error) {
	return email == "test@example.com" && token == "123456", nil
}

func Expect(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedBody string) {
	if w.Code != expectedStatus {
		t.Fatalf("Expected status code %v, got status code %v", expectedStatus, w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != expectedBody {
		t.Fatalf("Expected body `%v`, got `%v`", expectedBody, body)
	}
}

func ExpectEquals(t *testing.T, expected string, v string) {
	if expected != v {
		t.Fatalf(`Expected "%v" got "%v"`, expected, v)
	}
}

func RunAgainstRealDatabase() bool {
	return true // os.Getenv("WINDIR") == "C:\\WINDOWS"
}