package save

import "testing"


func TestShouldRejectPartitionsInInput(t *testing.T) {
	m := make([]map[string]interface{}, 1)
	m[0] = make(map[string]interface{})
	m[0]["id"] = "FC361743-67FC-4693-BF2D-D5CABC0BE8C2"
	m[0]["partition"] = "evil-attempt"

	if validate(m) == "" {
		t.Fatal("Expected error message, got empty/OK")
	}
}

func TestShouldRejectItemsWithoutId(t *testing.T) {
	m := make([]map[string]interface{}, 1)
	m[0] = make(map[string]interface{})
	m[0]["a"] = "stuff"
	m[0]["b"] = "other stuff"

	if err := validate(m); err != `One or more items is missing the property "id"` {
		t.Fatalf("Expected error message, got %s", err)
	}
}

func TestShouldRejectEmptyLists(t *testing.T) {
	m := make([]map[string]interface{}, 0)

	if err := validate(m); err != `Input array is empty, no items provided` {
		t.Fatalf("Expected error message, got %s", err)
	}
}

func TestShouldPassValidationWithNoErrors(t *testing.T) {
	m := make([]map[string]interface{}, 1)
	m[0] = make(map[string]interface{})
	m[0]["id"] = "FC361743-67FC-4693-BF2D-D5CABC0BE8C2"

	if validate(m) != "" {
		t.Fatal("Validation to pass")
	}
}

func TestShouldRejectTooShortId(t *testing.T) {
	m := make([]map[string]interface{}, 1)
	m[0] = make(map[string]interface{})
	m[0]["id"] = "123"

	if validate(m) != `The field "id" must be of length 32 or longer.` {
		t.Fatal("Expected error")
	}
}