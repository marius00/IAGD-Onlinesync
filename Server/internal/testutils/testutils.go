package testutils

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// IsolateStorage points STORAGE_PATH at a fresh temp directory (if not already
// set) so tests never touch the production /storage mount and parallel test
// binaries don't share a core.db. Call from a package's TestMain.
func IsolateStorage() {
	if os.Getenv("STORAGE_PATH") == "" {
		dir, err := os.MkdirTemp("", "iagdbackup-test-*")
		if err != nil {
			panic(err)
		}
		os.Setenv("STORAGE_PATH", dir)
	}
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