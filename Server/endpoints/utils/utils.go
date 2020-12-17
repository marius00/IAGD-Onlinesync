package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// GetJSONData extracts JSON from the POST body and returns it as a key:value map
func GetJSONData(body io.Reader) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	data, err := ioutil.ReadAll(body)

	if err != nil {
		return nil, err
	}

	if e := json.Unmarshal(data, &jsonMap); e != nil {
		return nil, e
	}

	return jsonMap, nil
}

func GetJSONDataSlice(body io.Reader) ([]map[string]interface{}, error) {
	jsonMap := make([]map[string]interface{}, 0)
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if e := json.Unmarshal(data, &jsonMap); e != nil {
		return nil, e
	}

	return jsonMap, nil
}

