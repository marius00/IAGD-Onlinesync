package upload

import (
	"github.com/marmyr/iagdbackup/internal/storage"
	"testing"
)

func TestShouldRejectPartitionsInInput(t *testing.T) {
	m := []storage.JsonItem{
		{ Id: "FC361743-67FC-4693-BF2D-D5CABC0BE8C2", UserId: 12345,},
	}

	if validate(m) == "" {
		t.Fatal("Expected error message, got empty/OK")
	}
}

func TestShouldRejectItemsWithoutId(t *testing.T) {
	m := []storage.JsonItem{
		{},
	}

	if err := validate(m); err != `One or more items is missing the property "id"` {
		t.Fatalf("Expected error message, got %s", err)
	}
}

func TestShouldRejectEmptyLists(t *testing.T) {
	var m []storage.JsonItem

	if err := validate(m); err != `Input array is empty, no items provided` {
		t.Fatalf("Expected error message, got %s", err)
	}
}

func TestShouldPassValidationWithNoErrors(t *testing.T) {
	m := []storage.JsonItem{
		{
			Id: "FC361743-67FC-4693-BF2D-D5CABC0BE8C2",
			BaseRecord: "my base record",
			Seed: 12345,
			StackCount: 1,
		},
	}

	if err := validate(m); err != "" {
		t.Fatalf("Validation failed with error %s", err)
	}
}

func TestShouldRejectTooShortId(t *testing.T) {
	m := []storage.JsonItem{
		{Id: "123",},
	}

	if validate(m) != `The field "id" must be of length 32 or longer.` {
		t.Fatal("Expected error")
	}
}
