package login

import (
	"strings"
	"testing"
)

func TestPincodeIs9Digits(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := generateRandomCode()
		if len(code) != 9 {
			t.Fatalf("Expected code of length %d, got length %d (%s)", 9, len(code), code)
		}
	}
}

// TODO: Ideally this should test one of the places that strings.ToLower is actually used on an email, but for now this will let me sleep at night.
func TestGolangToLower(t *testing.T) {
	if strings.ToLower("John@Gıthub.com") == strings.ToLower("John@Github.com") {
		t.Fatal("......")
	}
	if strings.ToLower("ß") == strings.ToLower("SS") {
		t.Fatal("......")
	}
}