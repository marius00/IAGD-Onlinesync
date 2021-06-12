package session

import (
	"encoding/json"
	"fmt"
	"github.com/marmyr/iagdbackup/api/session/auth"
	"github.com/marmyr/iagdbackup/api/session/login"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/testutils"
	"go.uber.org/zap"
	"testing"
)

type LoginReturnType struct {
	Key string `json:"key"`
}

func getPincode(t *testing.T) (string, string) {
	var pinCode string
	sendMail := func(logger zap.Logger, recipient string, code string) error {
		pinCode = code
		return nil
	}

	// Read the "key" param provided in json
	handler := login.ProcessRequestInternal(sendMail)
	resp := testutils.HostGetEndpoint(handler, "/?email=pincode@example.com")
	if resp.Code != 200 {
		t.Fatalf("Expected status code 200 requesting pincode, got status %d, %s", resp.Code, resp.Body.String())
	}

	var loginRet LoginReturnType
	testutils.FailOnError(t, json.Unmarshal([]byte(resp.Body.String()), &loginRet), "Error decoding JSON")

	return loginRet.Key, pinCode
}


type AuthReturnType struct {
	Token string `json:"token"`
}

func TestPincodeLoginFlow(t *testing.T) {
	throttleDb := storage.ThrottleDb{}
	testutils.FailOnError(t, throttleDb.Purge("sendmail:pincode@example.com", "sendmail:"), "Error purging records")
	testutils.FailOnError(t, throttleDb.Purge("", "verifyKey:"), "Error purging records")

	key, pinCode := getPincode(t)
	body := fmt.Sprintf("key=%s&code=%s", key, pinCode)
	resp := testutils.HostEndpoint(auth.ProcessRequest, body, nil)
	if resp.Code != 200 {
		t.Fatalf("Expected status code 200 on auth, got status %d, %s", resp.Code, resp.Body.String())
	}

	var ret AuthReturnType
	testutils.FailOnError(t, json.Unmarshal([]byte(resp.Body.String()), &ret), "Error decoding JSON")
	if len(ret.Token) != 36 {
		t.Fatalf("Expected token length 36, got token length %d, %s", len(ret.Token), ret.Token)
	}
}
