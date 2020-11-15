package utils

import (
	"io/ioutil"
	"strings"
	"testing"
)

func expected(t *testing.T, field string, expected string, got string) {
	t.Fatalf(`Expected %s=%s, got "%s" != nil`, field, expected, got)
}

func expect(t *testing.T, m map[string]interface{}, field string, expected string) {
	if m[field] != expected {
		t.Fatalf(`Expected %s=%s, got "%s" != nil`, field, expected, m[field])
	}
}

func TestDeserializeJsonItem(t *testing.T) {
	json := `
{
	"PartitionKey": "Some partition!",
	"timestamp":    13,
	"Plot":         "Nothing happens at all.",
	"Rating":       0.0
}`

	body := ioutil.NopCloser(strings.NewReader(json))
	m, err := GetJsonData(body)
	if err != nil {
		expected(t, "err", "!nil", err.Error())
	}

	expect(t, m, "PartitionKey", "Some partition!")
	if int64(m["timestamp"].(float64)) != 13 {
		t.Fatalf(`Expected %s=%v, got "%s" != nil`, "timestamp", 13, m["timestamp"])
	}
	expect(t, m, "Plot", "Nothing happens at all.")

	if m["Rating"].(float64) != 0.0 {
		t.Fatalf(`Expected %s=%v, got "%s" != nil`, "Rating", 13, m["Rating"])
	}
}

func TestDeserializeJsonItemArray(t *testing.T) {
	json := `
[{
	"PartitionKey": "Some partition!",
	"timestamp":    13,
	"Plot":         "Nothing happens at all.",
	"Rating":       0.0
}]`

	body := ioutil.NopCloser(strings.NewReader(json))
	arr, err := GetJsonDataSlice(body)
	if err != nil {
		expected(t, "err", "!nil", err.Error())
	}

	for _, m := range arr {
		expect(t, m, "PartitionKey", "Some partition!")
		if int64(m["timestamp"].(float64)) != 13 {
			t.Fatalf(`Expected %s=%v, got "%s" != nil`, "timestamp", 13, m["timestamp"])
		}
		expect(t, m, "Plot", "Nothing happens at all.")

		if m["Rating"].(float64) != 0.0 {
			t.Fatalf(`Expected %s=%v, got "%s" != nil`, "Rating", 13, m["Rating"])
		}
	}
}