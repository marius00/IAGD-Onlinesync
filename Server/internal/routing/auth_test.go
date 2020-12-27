package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/testutils"
	"net/http"
	"testing"
)

func TestMissingAuthHeaderShouldReturn401(t *testing.T) {
	headers := []testutils.HeaderEntry{
		{
			Name:  "X-Api-User",
			Value: "user@example.com",
		},
	}
	w := testutils.HostEndpoint(processRequest(true, t), "", headers)

	expected := `{"msg":"API: Authorization header missing"}`
	testutils.Expect(t, w, 401, expected)
}

func TestMissingUserHeaderShouldReturn401(t *testing.T) {
	headers := []testutils.HeaderEntry{
		{
			Name:  "Authorization",
			Value: "token",
		},
	}

	w := testutils.HostEndpoint(processRequest(true, t), "", headers)
	testutils.Expect(t, w, 401, `{"msg":"API: X-Api-User header missing"}`)
}

func TestInvalidTokenShouldReturn401(t *testing.T) {
	headers := []testutils.HeaderEntry{
		{
			Name:  "Authorization",
			Value: "token",
		},
		{
			Name:  "X-Api-User",
			Value: "user@example.com",
		},
	}

	w := testutils.HostEndpoint(processRequest(true, t), "", headers)
	testutils.Expect(t, w, 401, `{"msg":"API: Authorization token invalid"}`)
}

func TestValidTokenShouldReturn200(t *testing.T) {
	headers := []testutils.HeaderEntry{
		{
			Name:  "Authorization",
			Value: "123456",
		},
		{
			Name:  "X-Api-User",
			Value: "test@example.com",
		},
	}

	w := testutils.HostEndpoint(processRequest(false, t), "", headers)
	testutils.Expect(t, w, 200, `{"msg":"Everything went OK"}`)
}

// Ensures that the context is Aborted when it's supposed to. Mocks a "happy day 200 OK" return when not aborted.
func processRequest(isAborted bool, t *testing.T) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHandler := authorizedHandler(&testutils.DummyAuthorizer{}, &testutils.DummyThrottler{})
		authHandler(c)
		if c.IsAborted() != isAborted {
			t.Fatalf("Expected context IsAborted=%v got IsAborted=%v", c.IsAborted(), isAborted)
		}

		if !c.IsAborted() {
			user, _ := c.Get(AuthUserKey)
			if user != "test@example.com" {
				t.Fatalf(`Expected user to be "test@example.com", got "%s"`, user)
			}

			c.JSON(http.StatusOK, gin.H{"msg": "Everything went OK"})
		}
	}
}
