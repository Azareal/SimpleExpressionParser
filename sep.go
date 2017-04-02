package sep

import "strings"
import "strconv"
import "unicode"
import "errors"

type Datastore interface {
	SetVar(name, value string) error
	GetVar(name string) (string, bool)
	DeleteVar(name string) error
}

type ArbitraryBlock struct {
	Name string
	Contents string
	Type int // 0: Unknown, 1: int, 2: string, 3: list, 4: map, 5: variable, 6: function, 7: operator, 8: literal
	Extra interface{}
}

func HandleArbitraryCommands(command string, server_data Datastore, extra_data ...interface{}) (out string) {
	var currentBlock ArbitraryBlock
	var blocks []ArbitraryBlock
	var ntype int
	for i:=0;i < len(command);i++ {
		char := command[i]
		switch(ntype) {
			case 0: // Unknown
				if '0' <= char && char <= '9' {
					currentBlock = ArbitraryBlock{Contents:string(char),Type:1}
					ntype = 1
				} else if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') {
					currentBlock = ArbitraryBlock{Contents:string(char),Type:8}
					ntype = 8
				} else if char == '"' {
					currentBlock = ArbitraryBlock{Type:2}
					ntype = 2
				} else if char == '[' {
					currentBlock = ArbitraryBlock{Contents:string(char),Type:2}
					ntype = 3
				} else if char == '{' {
					currentBlock = ArbitraryBlock{Contents:string(char),Type:2}
					ntype = 4
				} else if char == '*' {
					currentBlock = ArbitraryBlock{Type:5}
					ntype = 5
				} else if char == '+' || char == '-' || char == '=' || char == '/' || char == '!' || char == '&' || char == '|' {
					currentBlock = ArbitraryBlock{Contents:string(char),Type:7}
					ntype = 7
				} else if !unicode.IsSpace(rune(char)) {
					return "Illegal character in arbitrary expression"
				}
			case 1: // Integer
				if !('0' <= char && char <= '9') {
					blocks = append(blocks, currentBlock)
					ntype = 0
				} else {
					currentBlock.Contents += string(char)
				}
			case 2: // String
				if char == '"' {
					blocks = append(blocks, currentBlock)
					ntype = 0
				} else {
					currentBlock.Contents += string(char)
				}
			case 3: // List
				currentBlock.Contents += string(char)
				if char == ']' {
					blocks = append(blocks, currentBlock)
					ntype = 0
				}
			case 4: // Map
				currentBlock.Contents += string(char)
				if char == '}' {
					blocks = append(blocks, currentBlock)
					ntype = 0
				}
			case 5: // Variable
				if char == ' ' {
					var err error
					currentBlock.Name = currentBlock.Contents
					currentBlock.Contents, err = ResolveVariable(currentBlock.Name,server_data)
					if err != nil {
						return err.Error()
					}
					blocks = append(blocks, currentBlock)
					ntype = 0
				} else {
					currentBlock.Contents += string(char)
				}
			case 6: // Functions
				if char == ')' {
					res, err := ResolveArbitraryFunction(currentBlock.Name,currentBlock.Contents,server_data,1)
					if err != nil {
						return err.Error()
					}
					currentBlock.Contents = res
					currentBlock.Type = 2
					blocks = append(blocks, currentBlock)
					ntype = 0
				}
				currentBlock.Contents += string(char)
			case 7: // Operators
				if char != '+' && char != '-' && char != '=' && char != '/' && char != '!' && char != '&' && char != '|' {
					blocks = append(blocks, currentBlock)
					ntype = 0
				} else {
					currentBlock.Contents += string(char)
				}
			case 8: // Literals
				if !('a' <= char && char <= 'z') && !('A' <= char && char <= 'Z') && char != '(' {
					currentBlock.Contents = strings.ToLower(currentBlock.Contents)
					if currentBlock.Contents == "true" || currentBlock.Contents == "false" {
						currentBlock.Type = 2
					}
					blocks = append(blocks, currentBlock)
					ntype = 0
				}
				if char == '(' {
					currentBlock.Name = currentBlock.Contents
					currentBlock.Contents = ""
					currentBlock.Type = 6
					ntype = 6
				} else {
					currentBlock.Contents += string(char)
				}
		}
	}
	if ntype != 0 {
		if ntype == 6 {
			return "there's an unclosed function call x.x"
		}
		if ntype == 5 {
			var err error
			currentBlock.Name = currentBlock.Contents
			currentBlock.Contents, err = ResolveVariable(currentBlock.Name,server_data)
			if err != nil {
				return err.Error()
			}
		}
		if ntype == 8 {
			currentBlock.Contents = strings.ToLower(currentBlock.Contents)
			if currentBlock.Contents == "true" || currentBlock.Contents == "false" {
				currentBlock.Type = 2
			}
		}
		blocks = append(blocks, currentBlock)
	}
	
	var outbuf_cursor int = 0
	var outbuf map[int]string = make(map[int]string)
	var boolInvert bool
	blockCount := len(blocks)
	for index := 0;index < blockCount;index++ {
		block := blocks[index]
		if block.Type == 1 || block.Type == 2 || block.Type == 5 {
			if index > 0 {
				if boolInvert {
					block.Contents = strings.ToLower(block.Contents)
					if block.Contents == "true" || block.Contents == "1" || block.Contents == "yes" {
						block.Contents = "false"
					} else if block.Contents == "false" || block.Contents == "0" || block.Contents == "no" {
						block.Contents = "true"
					}
				}
				
				prevtype := blocks[index - 1].Type
				if prevtype == 2 || prevtype == 1 { // Append to the previous string or int
					outbuf[outbuf_cursor] = outbuf[outbuf_cursor] + block.Contents
				} else {
					outbuf_cursor++
					outbuf[outbuf_cursor] = block.Contents
				}
			} else {
				outbuf[0] = block.Contents
			}
		} else if block.Type == 7 {
			if (blockCount - 1) <= index {
				return "Missing a right operand in the arbitrary expression"
			}
			if block.Contents == "!" {
				continue
			}
			if index == 0 {
				return "Missing a left operand in the arbitrary expression"
			}
			
			prevtype := blocks[index - 1].Type
			if prevtype == 7 {
				return "You cannot have an operator next to another operator in an arbitrary expression"
			}
			
			switch(block.Contents) {
				case "+": return "+ not implemented"
				case "-": return "- not implemented"
				case "=": return "= not implemented"
				case "/": return "/ not implemented"
				case "==": return "== not implemented"
				case "&&":
					previtem := blocks[index - 1]
					nextitem := blocks[index + 1]
					previtem_s, success := NormalizeBool(previtem.Contents)
					if !success {
						return "cannot coerce to bool"
					}
					nextitem_s, success := NormalizeBool(nextitem.Contents)
					if !success {
						return "cannot coerce to bool"
					}
					
					if previtem_s == "true" && nextitem_s == "true" {
						outbuf[outbuf_cursor] = "true"
					} else {
						outbuf[outbuf_cursor] = "false"
					}
					index++
				case "||":
					previtem := blocks[index - 1]
					nextitem := blocks[index + 1]
					previtem_s, success := NormalizeBool(previtem.Contents)
					if !success {
						return "cannot coerce string to bool"
					}
					nextitem_s, success := NormalizeBool(nextitem.Contents)
					if !success {
						return "cannot coerce string to bool"
					}
					
					if previtem_s == "true" || nextitem_s == "true" {
						outbuf[outbuf_cursor] = "true"
					} else {
						outbuf[outbuf_cursor] = "false"
					}
					index++
				default:
					return "Invalid operator"
			}
		} else {
			return "Unable to reduce to string"
		}
	}
	
	for _, item := range outbuf {
		out += item
	}
	return out
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
