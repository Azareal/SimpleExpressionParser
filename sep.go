package sep

import "strconv"
import "errors"

type Datastore interface {
	SetVar(name, value string) error
	GetVar(name string) (string, bool)
	DeleteVar(name string) error
}

func ResolveVariable(data string, server_data Datastore) (result string, err error) {
	if len(data) < 3 {
		return "", errors.New("variable name too short x.x")
	}
	if data[0] == '*' {
		data = data[1:]
	}
	
	// Validate the variable and break it up into chunks...
	var parts []string
	var buffer string
	var in_brackets bool
	for _, char := range data {
		if !('a' <= char && char <= 'z') && !('A' <= char && char <= 'Z') && !('0' <= char && char <= '9') && char != '[' && char != ']' && char != '.' && char != '_' && char != '-' {
			return "", errors.New("invalid character in variable name")
		}
		
		if in_brackets {
			if char == ']' {
				if buffer != "" {
					parts = append(parts,buffer)
					buffer = ""
				}
				in_brackets = false
				continue
			}
		}
		
		if char == '[' {
			if buffer != "" {
				parts = append(parts,buffer)
				buffer = ""
			}
			in_brackets = true
		} else if char == '.' {
			if buffer != "" {
				parts = append(parts,buffer)
				buffer = ""
			}
		} else {
			buffer += string(char)
		}
	}
	if buffer != "" {
		parts = append(parts,buffer)
	}
	
	partCount := len(parts)
	data_value, ok := server_data.GetVar(parts[0])
	if !ok {
		return "", errors.New("this variable doesn't exist o.o")
	}
	data_type := DetectType(data_value)
	if data_type == "string" && partCount == 2 {
		conv_data, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", errors.New("indices can only be integers x.x")
		}
		if len(data_value) <= conv_data {
			return "", errors.New("index out of range")
		}
		return string(data_value[conv_data]), nil
	}
	
	if partCount > 1 && data_type != "map" && data_type != "list" && data_type != "variable" {
		return "", errors.New("type " + data_type + " cannot have subelements")
	} else {
		return data_value, nil
	}
	
	var n int
	for _, part := range parts {
		if n > 5 {
			return "", errors.New("too many nested subelements")
		}
		switch(data_type) {
			case "map":
				pmap, err := MapParser(data_value)
				if err != nil {
					return "", errors.New("invalid map x.x")
				}
				pdata, ok := pmap[part]
				if !ok {
					return "", errors.New("field does not exist in map")
				}
				data_value = pdata
			case "list":
				plist, err := ListParser(data_value)
				if err != nil {
					return "", errors.New("invalid list x.x")
				}
				list_index, err := strconv.Atoi(part)
				if err != nil {
					return "", errors.New("list indices must be integers o.o")
				}
				if len(plist) <= list_index {
					return "", errors.New("index out of range o.o")
				}
				data_value = plist[list_index]
			case "variable":
				data_value, ok := server_data.GetVar(parts[0])
				if !ok {
					return "", errors.New("subvariable doesn't exist o.o")
				}
				data_type = DetectType(data_value)
		}
		n++
	}
	
	return data_value, nil
}
