package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// GetJsonData extracts JSON from the POST body and returns it as a key:value map
func GetJsonData(body io.ReadCloser) (map[string]interface{}, error) {
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

func GetJsonDataSlice(body io.ReadCloser) ([]map[string]interface{}, error) {
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

