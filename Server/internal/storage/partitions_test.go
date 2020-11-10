package storage

import (
	"github.com/marmyr/myservice/endpoints/testutils"
	"testing"
	"time"
)

func TestGeneratePartitionKeyFirstOfMonth(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	when := time.Date(2018, time.April, 1, 12, 0, 0, 0, loc)
	testutils.ExpectEquals(t, "2018:13:01", GeneratePartitionKey(when, 1))
}

func TestGeneratePartitionKeyStartOfWeek(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	when := time.Date(2018, time.April, 2, 12, 0, 0, 0, loc)
	testutils.ExpectEquals(t, "2018:14:15", GeneratePartitionKey(when, 15))
}

func TestGeneratePartitionKeyExceedingIterations(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	when := time.Date(2018, time.April, 2, 12, 0, 0, 0, loc)
	testutils.ExpectEquals(t, "2018:14:1015", GeneratePartitionKey(when, 1015))
}