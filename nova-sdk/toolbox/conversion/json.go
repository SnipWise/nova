package conversion

import (
	"bytes"
	"encoding/json"
)

// JsonStringToMap parses a JSON string and converts it to a map with string keys and any values
func JsonStringToMap(jsonString string) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func AnyToMap(data any) (map[string]any, error) {
	// Marshal en JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Unmarshal vers map
	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FromJSON converts a JSON string to a struct of type T
func FromJSON[T any](jsonString string) (T, error) {
	var result T
	err := json.Unmarshal([]byte(jsonString), &result)
	return result, err
}

func PrettyPrint(jsonStr string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(jsonStr), "", "  "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
