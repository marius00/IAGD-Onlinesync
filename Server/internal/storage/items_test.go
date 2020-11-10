package storage

import (
	"github.com/marmyr/myservice/endpoints/testutils"
	"testing"
)

func TestSanitizePartition(t *testing.T) {
	testutils.ExpectEquals(t, "b:c", SanitizePartition("a:b:c"))
	testutils.ExpectEquals(t, "c", SanitizePartition("c"))
	testutils.ExpectEquals(t, "c", SanitizePartition("b:c"))
	testutils.ExpectEquals(t, "a:b:c", SanitizePartition("x:a:b:c"))
}

func TestApplyOwner(t *testing.T) {
	testutils.ExpectEquals(t, "a:b:c", ApplyOwner(Partition{Partition:"b:c"}, "a"))

	initial := "b:c"
	owner := "owner@example.com"
	combined := ApplyOwner(Partition{Partition:initial}, owner)
	testutils.ExpectEquals(t, owner + ":" + initial, combined)

	sanitized := SanitizePartition(combined)
	testutils.ExpectEquals(t, initial, sanitized)
}