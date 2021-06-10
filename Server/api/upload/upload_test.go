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

	expected := `One or more items is missing the property "id"`
	if err := validate(m); err != expected {
		t.Fatalf("Unexpected error: `%s`, expected `%s`", err, expected)
	}
}

func TestShouldRejectEmptyLists(t *testing.T) {
	var m []storage.JsonItem

	expected := `Input array is empty, no items provided`
	if err := validate(m); err != expected {
		t.Fatalf("Unexpected error: `%s`, expected `%s`", err, expected)
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

	expected := `The field "id" must be of length 32 or longer.`
	if err := validate(m); err != expected {
		t.Fatalf("Unexpected error: `%s`, expected `%s`", err, expected)
	}
}

func TestShouldRejectTooLongBaseRecord(t *testing.T) {
	m := []storage.JsonItem{
		{
			Id: "AC361743-67FA-4693-BA2D-D5CFBC0BE8C2",
			BaseRecord: "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
			Seed: 12345,
			StackCount: 1,
		},
	}

	expected := `Item with id="AC361743-67FA-4693-BA2D-D5CFBC0BE8C2" has a one or more invalid records`
	if err := validate(m); err != expected {
		t.Fatalf("Unexpected error: `%s`, expected `%s`", err, expected)
	}
}

func TestShouldRejectMangledRecords(t *testing.T) {
	m := []storage.JsonItem{
		{
			Id: "AC361743-67FA-4693-BA2D-D5CFBC0BE8C2",
			BaseRecord: "records/items/lootaffixes/suffix/a038b_off_dmg�ther_07_je.dbr",
			Seed: 12345,
			StackCount: 1,
		},
	}

	expected := `Item with id="AC361743-67FA-4693-BA2D-D5CFBC0BE8C2" has a one or more invalid records`
	if err := validate(m); err != expected {
		t.Fatalf("Unexpected error: `%s`, expected `%s`", err, expected)
	}
}
