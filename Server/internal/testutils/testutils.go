package testutils

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func Expect(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedBody string) {
	if w.Code != expectedStatus {
		t.Fatalf("Expected status code %v, got status code %v", expectedStatus, w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != expectedBody {
		t.Fatalf("Expected body `%v`, got `%v`", expectedBody, body)
	}
}
/*
func ExpectEquals(t *testing.T, expected string, v string) {
	if expected != v {
		t.Fatalf(`Expected "%v" got "%v"`, expected, v)
	}
}*/

func ExpectEquals(t *testing.T, expected interface{}, v interface{}, error string) {
	if expected != v {
		t.Fatalf(`%s, Expected "%v" got "%v"`, error, expected, v)
	}
}

func RunAgainstRealDatabase() bool {
	return true // os.Getenv("WINDIR") == "C:\\WINDOWS"
}

func FailOnError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s, %v", message, err)
	}
}