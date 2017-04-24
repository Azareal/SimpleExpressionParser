package sep

import "fmt"
import "strings"
import "strconv"
import "unicode"
import "errors"

type Datastore interface {
	VarExists(name string) bool
	SetVar(name, value string) error
	GetVar(name string) (string, bool)
	DeleteVar(name string) error
}

type ArbitraryOptions struct {
	Comments bool
	Multiline bool
	//EnableStoreVarRead bool
	//EnableStoreVarWrite bool
}

type ArbitraryBlock struct {
	Name string
	Contents string
	Type int // 0: Unknown, 1: int, 2: string, 3: list, 4: map, 5: variable, 6: function, 7: operator, 8: literal, 9: comment
	Extra interface{}
}

func HandleArbitraryCommands(command string, ds Datastore, extra_data ...interface{}) (out string, lastIndex int, err error) {
	return parseArbitraryBlock(command, ds, ArbitraryOptions{Comments:true,Multiline:true}, 0, extra_data...)
}

func parseArbitraryBlock(command string, ds Datastore, options ArbitraryOptions, n int, extra_data ...interface{}) (out string, lastIndex int, err error) {
	if n > 5 {
		return "", -1, errors.New("too many nested calls x.x")
	}
	lastIndex = -1
	
	var currentBlock ArbitraryBlock
	var blocks []ArbitraryBlock
	var ntype, brace_count int
	var last_if int8 // Ternary value. -1, 0, 1. Needed to tell an else node whether it should eat it's block or not
	
CharLoop:
	for i:=0;i < len(command);i++ {
		char := command[i]
		switch(ntype) {
			case 0: // Unknown node
				// Sniff the type of the next node
				switch {
				case '0' <= char && char <= '9':
					currentBlock = ArbitraryBlock{Contents:string(char),Type:1}
					ntype = 1
				case ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z'):
					currentBlock = ArbitraryBlock{Contents:string(char),Type:8}
					ntype = 8
				case char == '"':
					currentBlock = ArbitraryBlock{Type:2}
					ntype = 2
				case char == '[':
					currentBlock = ArbitraryBlock{Contents:string(char),Type:2}
					ntype = 3
				case char == '{':
					currentBlock = ArbitraryBlock{Contents:string(char),Type:2}
					ntype = 4
				case char == '*':
					currentBlock = ArbitraryBlock{Type:5}
					ntype = 5
				case char == '+' || char == '-' || char == '=' || char == '!' || char == '&' || char == '|':
					currentBlock = ArbitraryBlock{Contents:string(char),Type:7}
					ntype = 7
				case char == '#' || char == '/':
					if options.Comments {
						if !options.Multiline {
							break CharLoop
						}
						currentBlock = ArbitraryBlock{Contents:string(char),Type:9}
						ntype = 9
					}
				case char == ':': // Magic break character, needed for Go style switches
					lastIndex = i
					break CharLoop
				case !unicode.IsSpace(rune(char)) && char != '`':
					return "", i, errors.New("Illegal character '" + string(char) + "' in arbitrary expression")
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
					currentBlock.Name = currentBlock.Contents
					currentBlock.Contents, err = ResolveVariable(currentBlock.Name,ds)
					if err != nil {
						return "", i, err
					}
					blocks = append(blocks, currentBlock)
					ntype = 0
				} else {
					currentBlock.Contents += string(char)
				}
			case 6: // Functions
				if char == '(' {
					brace_count++
				}
				if char == ')' {
					if brace_count == 0 {
						// Is this a control structure...?
						if currentBlock.Name == "if" {
							res, _, err := parseArbitraryBlock(currentBlock.Contents, ds, options, n + 1, extra_data...)
							if err != nil {
								return "", i, err
							}
							
							norm, ok := NormalizeBool(res)
							if !ok {
								return "", i, errors.New("if statements only accept boolean values")
							}
							
							// Eat everything upto the next line
							if norm == "false" {
								for ; i < len(command);i++ {
									if command[i] == 10 {
										break
									}
								}
								last_if = -1
							} else {
								last_if = 1
							}
							ntype = 0
							continue
						} else if currentBlock.Name == "switch" {
							switch_text, _, err := parseArbitraryBlock(currentBlock.Contents, ds, options, n + 1, extra_data...)
							if err != nil {
								return "", i, err
							}
							fmt.Println("switch_text",switch_text)
							
							var brace_count int = -1
							var inner_text, default_text string
							for i++; i < len(command);i++ {
								char = command[i]
								if char == '{' {
									brace_count++
									fmt.Println("brace_count++",brace_count)
								} else if char == '}' {
									if brace_count == 0 {
										fmt.Println("brace break")
										i = skipCurrentBlock(command,i)
										break
									}
									brace_count--
									fmt.Println("brace_count--",brace_count)
								} else if char == ':' && inner_text != "" {
									inner_text = strings.TrimSpace(inner_text)
									if len(command) > (i+1) {
										i++
									}
									fmt.Println("label",inner_text)
									
									startIndex := i
									if inner_text != "default" {
										res, _, err := parseArbitraryBlock(inner_text, ds, options, n + 1, extra_data...)
										if err != nil {
											return "", i, err
										}
										fmt.Println("res_label",res)
										
										i = skipUntilChar(command,i,',')
										inner_text = command[startIndex:i]
										if res == switch_text {
											i = skipCurrentBlock(command,i)
											break
										}
									} else {
										i = skipUntilChar(command,i,',')
										default_text = command[startIndex:i]
										fmt.Println("default_text",default_text)
									}
									inner_text = ""
								} else if char != ':' {
									inner_text += string(char)
								}
							}
							
							if brace_count < 0 {
								return "", i, errors.New("You missed an opening brace for or within a switch block")
							}
							
							fmt.Println("post inner_text",inner_text)
							fmt.Println("post default_text",default_text)
							if inner_text == "" && default_text != "" {
								inner_text = default_text
							}
							if inner_text == "" {
								ntype = 0
								continue
							}
							
							res, _, err := parseArbitraryBlock(inner_text, ds, options, n + 1, extra_data...)
							if err != nil {
								fmt.Println("Err:",err.Error())
								fmt.Println("Data:",inner_text)
								fmt.Println("Result:",res)
								return "", i, err
							}
							
							currentBlock.Contents = res
							currentBlock.Type = 2
							blocks = append(blocks, currentBlock)
							ntype = 0
							continue
						}
						
						res, err := ResolveArbitraryFunction(currentBlock.Name,currentBlock.Contents,ds,n,extra_data...)
						if err != nil {
							return "", i, err
						}
						
						currentBlock.Contents = res
						currentBlock.Type = 2
						blocks = append(blocks, currentBlock)
						ntype = 0
					} else {
						brace_count--
					}
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
					if currentBlock.Contents == "as" {
						currentBlock.Type = 7
					}
					
					if currentBlock.Contents == "else" {
						if last_if == 0 {
							return "", i, errors.New("There's no if statement to match this else to!")
						}
						
						// Eat everything upto the next line
						if last_if == 1 {
							for ; i < len(command);i++ {
								if command[i] == 10 {
									break
								}
							}
						}
						ntype = 0
						last_if = 0
						continue
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
			case 9: // Comments
				switch(currentBlock.Contents) {
					case "/":
						if char == '/' {
							currentBlock.Contents = "//"
						} else if char == '*' {
							currentBlock.Contents = "/*"
						} else {
							i--
							currentBlock.Type = 7
							ntype = 7
						}
					case "#","//":
						if char == 10 { // Newline character
							ntype = 0
						}
					case "/*":
						if char == '*' && ((i + 1) < len(command)) && command[i + 1] == '/' {
							ntype = 0
							i++
						}
				}
		}
	}
	if ntype != 0 {
		if ntype == 6 {
			return "", lastIndex, errors.New("there's an unclosed function call x.x")
		}
		
		if ntype == 5 {
			var err error
			currentBlock.Name = currentBlock.Contents
			currentBlock.Contents, err = ResolveVariable(currentBlock.Name,ds)
			if err != nil {
				return "", lastIndex, err
			}
		}
		
		if ntype == 8 {
			currentBlock.Contents = strings.ToLower(currentBlock.Contents)
			if currentBlock.Contents == "true" || currentBlock.Contents == "false" {
				currentBlock.Type = 2
			}
			if currentBlock.Contents == "as" {
				currentBlock.Type = 7
			}
		}
		
		if ntype != 9 {
			blocks = append(blocks, currentBlock)
		}
	}
	
	var outbuf_cursor int // = 0
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
				return "", lastIndex, errors.New("Missing a right operand in the arbitrary expression")
			}
			if block.Contents == "!" {
				continue
			}
			if index == 0 {
				return "", lastIndex, errors.New("Missing a left operand in the arbitrary expression")
			}
			
			prevtype := blocks[index - 1].Type
			if prevtype == 7 {
				return "", lastIndex, errors.New("You cannot have an operator next to another operator in an arbitrary expression")
			}
			
			switch(block.Contents) {
				case "+": return "", lastIndex, errors.New("+ not implemented")
				case "-": return "", lastIndex, errors.New("- not implemented")
				case "=": return "", lastIndex, errors.New("= not implemented")
				case "as": return "", lastIndex, errors.New("as not implemented")
				case "++": return "", lastIndex, errors.New("++ not implemented")
				case "--": return "", lastIndex, errors.New("-- not implemented")
				case "+=": return "", lastIndex, errors.New("+= not implemented")
				case "-=": return "", lastIndex, errors.New("-= not implemented")
				case "/": return "", lastIndex, errors.New("/ not implemented")
				case "==": return "", lastIndex, errors.New("== not implemented")
				case "&&":
					previtem := blocks[index - 1]
					nextitem := blocks[index + 1]
					
					previtem_s, success := NormalizeBool(previtem.Contents)
					if !success {
						return "", lastIndex, errors.New("cannot coerce to bool")
					}
					
					nextitem_s, success := NormalizeBool(nextitem.Contents)
					if !success {
						return "", lastIndex, errors.New("cannot coerce to bool")
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
						return "", lastIndex, errors.New("cannot coerce string to bool")
					}
					
					nextitem_s, success := NormalizeBool(nextitem.Contents)
					if !success {
						return "", lastIndex, errors.New("cannot coerce string to bool")
					}
					
					if previtem_s == "true" || nextitem_s == "true" {
						outbuf[outbuf_cursor] = "true"
					} else {
						outbuf[outbuf_cursor] = "false"
					}
					index++
				default:
					return "", lastIndex, errors.New("Invalid operator")
			}
		} else {
			//fmt.Println(ntype)
			//fmt.Println(block)
			return "", lastIndex, errors.New("Unable to reduce to string")
		}
	}
	
	for _, item := range outbuf {
		out += item
	}
	return out, lastIndex, nil
}

func ResolveVariable(data string, ds Datastore) (result string, err error) {
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
	data_value, ok := ds.GetVar(parts[0])
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
				data_value, ok := ds.GetVar(parts[0])
				if !ok {
					return "", errors.New("subvariable doesn't exist o.o")
				}
				data_type = DetectType(data_value)
		}
		n++
	}
	
	return data_value, nil
}

func skipBlock(data string, index int) int {
	var brace_count int
	for ;index < len(data);index++{
		char := data[index]
		if char == '{' {
			brace_count++
		} else if char == '}' {
			if brace_count == 0 {
				return index
			}
			brace_count--
		}
	}
	return index
}

func skipFunc(data string, index int) int {
	var brace_count int
	for ;index < len(data);index++{
		char := data[index]
		if char == '(' {
			brace_count++
		} else if char == ')' {
			if brace_count == 0 {
				return index
			}
			brace_count--
		}
	}
	return index
}

func skipList(data string, index int) int {
	var brace_count int
	for ;index < len(data);index++{
		char := data[index]
		if char == '[' {
			brace_count++
		} else if char == ']' {
			if brace_count == 0 {
				return index
			}
			brace_count--
		}
	}
	return index
}

func skipUntilChar(data string, index int, char byte) int {
SwitchLoop:
	for ; index < len(data);index++ {
		switch(data[index]) {
		case '{': index = skipBlock(data,index)
		case '[': index = skipList(data,index)
		case '(': index = skipFunc(data,index)
		case '}':
			if index != 0 {
				index--
			}
			break SwitchLoop
		case char: break SwitchLoop
		}
	}
	return index
}

func skipCurrentBlock(data string, index int) int {
SwitchLoop:
	for ; index < len(data);index++ {
		switch(data[index]) {
		case '{': index = skipBlock(data,index)
		case '[': index = skipList(data,index)
		case '(': index = skipFunc(data,index)
		case '}': break SwitchLoop
		}
	}
	return index
}

/*func getBlockLastIndex() (index int) {
	
}*/
