package sep

import "strings"
import "errors"

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
