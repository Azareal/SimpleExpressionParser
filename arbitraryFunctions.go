package sep

import "strings"
import "strconv"
import "time"
import "errors"

var arbitraryFunctions map[string]func(...string)(string,error) = map[string]func(...string)(string,error){
	"unixtime": func(_ ...string) (string, error) {
		return strconv.FormatInt(time.Now().Unix(),10), nil
	},
}
var arbitraryFunctionParams map[string]int = map[string]int{"unixtime":0}

func GetArbitraryFunction(name string) (func(...string)(string,error),bool) {
	res, ok := arbitraryFunctions[name]
	return res, ok
}

func SetArbitraryFunction(name string, callback func(...string)(string,error), paramCount int) {
	arbitraryFunctions[name] = callback
	arbitraryFunctionParams[name] = paramCount
}

func HasArbitraryFunction(name string) bool {
	_, ok := arbitraryFunctions[name]
	return ok
}


func ResolveArbitraryFunction(name string, paramstr string, ds Datastore, depth int) (result string, err error) {
	callback, ok := arbitraryFunctions[name]
	if !ok {
		return "", errors.New("this function doesn't exist x.x")
	}
	
	params := strings.Split(paramstr,",")
	if arbitraryFunctionParams[name] == 0 {
		return callback(params...)
	}
	
	/*for _, param := range params {
		if param[0] == '*' {
			
		} else if param[0] == '"' {
			
		} else {
			
		}
	}*/
	
	return ":soontm:", nil
}