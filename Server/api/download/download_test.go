package download

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)


/*
func TestShouldReturn400OnMissingPartitionKey(t *testing.T) {
	w := HostEndpoint(testHandler(), "/")
	testutils.Expect(t, w, 400, `{"msg":"The query parameter \"partition\" is missing"}`)
}

func TestShouldReturn400OnInvalidPartitionKey(t *testing.T) {
	testutils.Expect(t, HostEndpoint(testHandler(), "/?partition=2020:40"), 400, `{"msg":"The query parameter \"partition\" is invalid"}`)
	testutils.Expect(t, HostEndpoint(testHandler(), "/?partition=2019:40:1"), 400, `{"msg":"The query parameter \"partition\" is invalid"}`)
	testutils.Expect(t, HostEndpoint(testHandler(), "/?partition=3000:40:1"), 400, `{"msg":"The query parameter \"partition\" is invalid"}`)
	testutils.Expect(t, HostEndpoint(testHandler(), "/?partition=2020:xx:1"), 400, `{"msg":"The query parameter \"partition\" is invalid"}`)
}

func TestShouldReturn404OnMissingPartition(t *testing.T) {
	testutils.Expect(t, HostEndpoint(testHandler(), "/?partition=2020:04:1"), 404, `{"msg":"Partition does not exist"}`)
}

func TestShouldSucceedFetchingItemsAndDeletedEntries(t *testing.T) {
	// TODO: Some more details here might be nice.. nuances..
	testutils.Expect(t, HostEndpoint(testHandler(), "/?partition=2020:05:1"), 200, `{"items":[{"BaseRecord":"stuff/here/etc","id":"123456"}],"deleted":[]}`)
}
// Ensures that the context is Aborted when it's supposed to. Mocks a "happy day 200 OK" return when not aborted.
func testHandler() gin.HandlerFunc {
	partitionDb := storage.InMemoryPartitionDb{
		Entries: map[string]storage.Partition {
			"download@example.com@@@2020:05:1": {Partition: "download@example.com:2020:05:1", NumItems: 5, IsActive: false, Email: "download@example.com"},
		},
	}

	deletedItemDb := storage.InMemoryDeletedItemDb {}

	item := storage.Item {
		"BaseRecord": "stuff/here/etc",
		"id": "123456",
	}

	itemDb := storage.InMemoryItemDb {
		Entries:map[string][]storage.Item {
			"download@example.com@@@2020:05:1": {
				item,
			},
		},
	}

	return func(c *gin.Context) {
		c.Set(routing.AuthUserKey, "download@example.com")
		p := processRequest(&itemDb)
		p(c)
	}
}

*/
func HostEndpoint(f gin.HandlerFunc, url string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	r := gin.Default()
	r.GET("/", f)

	r.ServeHTTP(w, req)

	return w
}