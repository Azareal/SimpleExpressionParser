package sep

import "strings"
import "strconv"
import "errors"

type Datastore interface {
	SetVar(name, value string) error
	GetVar(name string) (string, bool)
	DeleteVar(name string) error
}

func ListParser(data string) (output []string, err error) {
	if data == "" {
		return
	}
	if data[0] != '[' {
		return output, errors.New("I couldn't find the opening tag o.o")
	}
	if data[len(data) - 1] != ']' {
		return output, errors.New("I couldn't find the closing tag o.o")
	}
	if data == "[]" {
		return
	}
	
	var startIndex int
	var itemLen int = 0
	for i := 1;i < len(data);i++ {
		if data[i] == ',' || data[i] == ']' {
			if itemLen != 0 {
				output = append(output,data[startIndex + 1:i])
			}
			startIndex = i
		} else {
			itemLen++
		}
	}
	//fmt.Println(output)
	return
}

func MapParser(data string) (output map[string]string, err error) {
	output = make(map[string]string)
	if data == "" {
		return
	}
	if data[0] != '[' && data[0] != '{' {
		return output, errors.New("I couldn't find the opening tag o.o")
	}
	if data[len(data) - 1] != ']' && data[len(data) - 1] != '}' {
		return output, errors.New("I couldn't find the closing tag o.o")
	}
	if data == "[]" || data == "{}" {
		return
	}
	
	if data[0] == '[' && data[len(data) - 1] != ']' {
		return output, errors.New("the opening and closing tags don't match x.x")
	}
	if data[0] == '{' && data[len(data) - 1] != '}' {
		return output, errors.New("the opening and closing tags don't match x.x")
	}
	
	data = NormalizeMapString(data)
	
	var elements []string
	var startIndex int
	var itemLen int = 0
	for i := 1;i < len(data);i++ {
		if data[i] == ',' || data[i] == '}' {
			if itemLen != 0 {
				elements = append(elements,data[startIndex + 1:i])
			}
			startIndex = i
		} else {
			itemLen++
		}
	}
	//fmt.Println(elements)
	
	for _, element := range elements {
		if strings.Index(element,":") == -1 {
			return output, errors.New("I couldn't find the : seperator between an element name and an element value o.o")
		}
		
		fields := strings.SplitN(element,":",2)
		if fields[0] == "" {
			return output, errors.New("You can't have a blank name for a field x.x")
		}
		if fields[1] == "" {
			return output, errors.New("You can't have an empty field x.x")
		}
		
		_, exists := output[fields[0]]
		if exists {
			return output, errors.New("You can't have two fields with the same name")
		}
		output[fields[0]] = fields[1]
	}
	// fmt.Printf("%+v\n", output)
	return
}

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
