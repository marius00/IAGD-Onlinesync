package login

import "testing"

func TestPincodeIs9Digits(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := generateRandomCode()
		if len(code) != 9 {
			t.Fatalf("Expected code of length %d, got length %d (%s)", 9, len(code), code)
		}
	}
}
