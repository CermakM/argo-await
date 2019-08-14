package common

import "encoding/json"

// FormatJSON formats an object and returns JSON string
func FormatJSON(obj interface{}) string {
	jsonified, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(jsonified)
}
