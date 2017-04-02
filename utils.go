package sep

import "strconv"

func DetectType(data string) string {
	if len(data) == 0 {
		return "string"
	}
	if data[0] == '[' {
		return "list"
	}
	if data[0] == '{' {
		return "map"
	}
	if data[0] == '*' {
		return "variable"
	}
	_, err := strconv.Atoi(data)
	if err == nil {
		return "int"
	}
	return "string"
}

// Normalize the [s and ]s into {s and }s for maps
func NormalizeMapString(data string) string {
	if len(data) < 3 {
		return "{}"
	}
	if data == "[]" {
		return "{}"
	}
	if data[0] == '[' && data[len(data) - 1] == ']' {
		data = data[1:len(data) - 1]
		data = "{" + data + "}"
	}
	return data
}
