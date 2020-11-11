package save

import "testing"


func TestShouldRejectPartitionsInInput(t *testing.T) {
	m := make([]map[string]interface{}, 1)
	m[0] = make(map[string]interface{})
	m[0]["id"] = "ok"
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