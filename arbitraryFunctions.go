package sep

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
