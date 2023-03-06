package tkt

import (
	"encoding/json"
	"strings"
)

const (
	malformedOriginalError  = "Malformed original JSON"
	malformedGeneratedError = "Malformed generated JSON"
	malformedNodeError      = "Malformed input JSON node"
	malformedFieldError     = "Malformed input JSON field"
	malformedListError      = "Malformed input JSON list"
	malformedTerminalError  = "Malformed input JSON terminal"
	sanitizedSecret         = "**********"
	maxStringFieldLength    = 200
)

var secretKeys []string

func init() {
	secretKeys = []string{
		"password",
		"ssn",
		"socialsecuritynumber",
		"fein",
		"taxid",
	}
}

func AppendSecretKeys(newKeys []string) {
	secretKeys = append(secretKeys, newKeys...)
}

func SanitizeObject(obj interface{}) string {
	rawData, err := json.Marshal(obj)
	if err != nil {
		return malformedOriginalError
	}
	return SanitizeJson(rawData)
}

func SanitizeJson(data []byte) string {
	obj := parseJsonField("", data)
	sanitized, err := json.Marshal(obj)
	if err != nil {
		return malformedGeneratedError
	}
	return string(sanitized)
}

func BuildSanitizeJson(data interface{}) json.RawMessage {
	obj := parseJsonField("", Marshal(data))
	sanitized, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return sanitized
}

func parseJsonField(key string, data []byte) interface{} {
	if len(data) < 1 {
		return malformedFieldError
	}
	switch data[0] {
	case '{':
		return parseJsonNode(data)
	case '[':
		return parseJsonList(key, data)
	default:
		return parseJsonTerminal(key, data)
	}
}

func parseJsonNode(data []byte) interface{} {
	obj := make(map[string]interface{})
	var parsed map[string]json.RawMessage
	err := json.Unmarshal(data, &parsed)
	if err != nil {
		return malformedNodeError
	}
	for key, value := range parsed {
		obj[key] = parseJsonField(strings.ToLower(key), value)
	}
	return obj
}

func parseJsonList(key string, data []byte) interface{} {
	obj := make([]interface{}, 0)
	var parsed []json.RawMessage
	err := json.Unmarshal(data, &parsed)
	if err != nil {
		return malformedListError
	}
	for _, value := range parsed {
		obj = append(obj, parseJsonField(key, value))
	}
	return obj
}

func parseJsonTerminal(key string, data []byte) interface{} {
	if len(data) < 1 {
		return malformedTerminalError
	}
	if InStringList(key, secretKeys) {
		return sanitizedSecret
	}
	if data[0] == '"' && len(data) > maxStringFieldLength {
		return string(append(data[1:maxStringFieldLength], "..."...))
	}
	var obj interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return malformedTerminalError
	}
	return obj
}
