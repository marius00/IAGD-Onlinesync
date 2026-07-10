package delete

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/util"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteAccount(t *testing.T) {
	userDb := storage.UserDb{}
	itemDb := storage.ItemDb{}
	characterDb := storage.CharacterDb{}
	authDb := storage.AuthDb{}

	email := fmt.Sprintf("%s@example.com", uuid.NewV4().String())
	ts := util.GetCurrentTimestamp()
	userId := CreateTestUser(t, email)
	defer userDb.Purge(userId)
	defer itemDb.Purge(email)
	storage.Preload()

	expected := storage.JsonItem{
		Id:         uuid.NewV4().String(),
		Ts:         ts,
		BaseRecord: "my base record",
	}
	deletingThis := storage.JsonItem{
		Id:         uuid.NewV4().String(),
		Ts:         ts,
		BaseRecord: "my base record2",
	}

	inputItems, _ := itemDb.ToInputItems([]storage.JsonItem{expected, deletingThis})
	err := itemDb.Insert(email, inputItems)
	assert.NoErrorf(t, err, "Error inserting item")

	err = itemDb.Delete(context.Background(), email, []string{deletingThis.Id}, ts)
	assert.NoErrorf(t, err, "Expected no error")

	entry := storage.CharacterEntry{
		Name:     "Pete",
		Filename: "fileynamey",
	}

	err = characterDb.Insert(email, entry)
	assert.NoErrorf(t, err, "Expected no error")

	accessToken := uuid.NewV4().String()
	err = authDb.StoreSuccessfulAuth(email, userId, "key", accessToken)
	assert.NoErrorf(t, err, "Expected no error")

	{
		items, err := itemDb.List(context.Background(), email, 0)
		assert.NoErrorf(t, err, "Expected no error")
		assert.Len(t, items, 1, "Expected 1 item")

		deletedItems, err := itemDb.ListDeletedItems(email, 0)
		assert.NoErrorf(t, err, "Expected no error")
		assert.Len(t, deletedItems, 1, "Expected 1 deleted item")

		user, err := userDb.Get(userId)
		assert.NoErrorf(t, err, "Expected no error")
		assert.NotNil(t, user, "Expected a user")

		characters, err := characterDb.List(email)
		assert.NoErrorf(t, err, "Expected no error")
		assert.Len(t, characters, 1, "Expected 1 character")

		latestAuthToken := authDb.GetLatestAuthToken(email)
		assert.Equal(t, accessToken, latestAuthToken)
	}

	CallEndpoint(t, userId, email)

	{
		items, err := itemDb.List(context.Background(), email, 0)
		assert.NoErrorf(t, err, "Expected no error")
		assert.Len(t, items, 0, "Expected zero items")

		deletedItems, err := itemDb.ListDeletedItems(email, 0)
		assert.NoErrorf(t, err, "Expected no error")
		assert.Len(t, deletedItems, 0, "Expected zero deleted items")

		user, err := userDb.Get(userId)
		assert.NoErrorf(t, err, "Expected no error")
		assert.Nil(t, user, "Expected no user")

		characters, err := characterDb.List(email)
		assert.NoErrorf(t, err, "Expected no error")
		assert.Len(t, characters, 0, "Expected no character")

		latestAuthToken := authDb.GetLatestAuthToken(email)
		assert.Equal(t, "", latestAuthToken)
	}

	// TODO: Verify no auth token
}

func CallEndpoint(t *testing.T, userId config.UserId, email string) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// Create a new http.Request
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	// Set the RemoteAddr of the request
	req.RemoteAddr = "192.168.1.1:12345" // Example IP and port

	// Assign the modified request to the Gin context
	ctx.Request = req
	ctx.Set("AuthUserKey", userId)
	ctx.Set("AuthEmailKey", email)
	ProcessRequest(ctx)

	assert.Equalf(t, http.StatusOK, w.Code, "Expected status code OK")
}

// Create a clean user for tests
func CreateTestUser(t *testing.T, email string) config.UserId {
	userDb := storage.UserDb{}
	itemDb := storage.ItemDb{}

	userId, err := userDb.Insert(storage.UserEntry{
		Email: email,
	})
	if err != nil {
		t.Fatalf("Error inserting user, %v", err)
	}

	// Ensure we have no left-over data for this user
	if err := itemDb.Purge(email); err != nil {
		t.Fatal("Failed to purge user")
	}

	return userId
}
