package sep

import "strconv"
import "time"
import "errors"
import "math/rand"

var arbitraryFunctions map[string]func(Datastore,...string)(string,error) = map[string]func(Datastore,...string)(string,error){
	"unixtime": func(_ Datastore, _ ...string) (string, error) {
		return strconv.FormatInt(time.Now().Unix(),10), nil
	},
	
	"bool": func(_ Datastore, params ...string) (string, error) {
		norm, ok := NormalizeBool(params[0])
		if !ok {
			return "", errors.New("bool() only accepts boolean values not whatever it is you're trying to pass to it")
		}
		return norm, nil
	},
	
	"rand": func(_ Datastore, params ...string) (string, error) {
		maxNumber, err := strconv.Atoi(params[0])
		if err != nil {
			return "", errors.New("rand() only accepts positive integers")
		}
		if maxNumber == 0 {
			return "0", nil
		}
		if maxNumber < 0 {
			return "", errors.New("rand() does not accept negative numbers")
		}
		rand.Seed(time.Now().Unix())
		return strconv.Itoa(rand.Intn(maxNumber)), nil
	},
	
	"len": func(ds Datastore, params ...string) (string, error) {
		data_type := DetectType(params[0])
		data_value := params[0]
		if data_type == "variable" {
			if len(params[0]) < 2 {
				return "", errors.New("super bad variable x.x")
			}
			if ds == nil {
				return "", errors.New("not a valid variable x.x")
			}
			var ok bool
			data_value, ok = ds.GetVar(data_value[1:])
			if !ok {
				return "", errors.New("not a good variable o.o")
			}
			data_type = DetectType(data_value)
		}
		
		if data_type == "int" || data_type == "string" {
			return strconv.Itoa(len(data_value)), nil
		}
		if data_type == "list" {
			plist, err := ListParser(data_value)
			if err != nil {
				return "", errors.New("bad list x.x")
			}
			return strconv.Itoa(len(plist)), nil
		}
		if data_type == "map" {
			pmap, err := MapParser(data_value)
			if err != nil {
				return "", errors.New("bad map x.x")
			}
			return strconv.Itoa(len(pmap)), nil
		}
		return "", errors.New("bad data type")
	},
}
var arbitraryFunctionParams map[string]int = map[string]int{
	"unixtime": 0,
	"bool": 1,
	"rand": 1,
	"len": 1,
}
var arbitraryFunctionsWantsRaw map[string]func(Datastore,string)(string,error) = map[string]func(Datastore,string)(string,error){
	"exists": func(ds Datastore, params string) (string, error) {
		if ds.VarExists(params) {
			return "true", nil
		}
		return "false", nil
	},
}

func GetArbitraryFunction(name string) (func(Datastore,...string)(string,error),bool) {
	res, ok := arbitraryFunctions[name]
	return res, ok
}

func SetArbitraryFunction(name string, callback func(Datastore,...string)(string,error), paramCount int) {
	arbitraryFunctions[name] = callback
	arbitraryFunctionParams[name] = paramCount
}

func HasArbitraryFunction(name string) bool {
	_, ok := arbitraryFunctions[name]
	return ok
}


func ResolveArbitraryFunction(name string, paramstr string, ds Datastore, depth int, extra_data ...interface{}) (result string, err error) {
	callback2, ok := arbitraryFunctionsWantsRaw[name]
	if ok {
		return callback2(ds,paramstr)
	}
	callback, ok := arbitraryFunctions[name]
	if !ok {
		return "", errors.New("this function doesn't exist x.x")
	}
	
	params, err := parseParams(paramstr)
	if err != nil {
		return "", err
	}
	reqParams := arbitraryFunctionParams[name]
	if reqParams == 0 {
		return callback(ds,params...)
	}
	
	if len(params) != reqParams {
		if len(params) == 0 {
			return "", errors.New("No parameters provided x.x")
		} else if len(params) > reqParams {
			return "", errors.New("Too many parameters provided x.x")
		}
		return "", errors.New("You need to provide more parameters x.x")
	}
	
	for index, param := range params {
		res, _, err := parseArbitraryBlock(param, ds, ArbitraryOptions{Comments:false}, depth + 1, extra_data...)
		if err != nil {
			return "", err
		}
		params[index] = res
	}
	
	return callback(ds, params...)
}

func parseParams(paramstr string) (params []string, err error) {
	if paramstr == "" {
		return params, nil
	}
	
	var buffer string
	for i := 0;i < len(paramstr);i++ {
		char := paramstr[i]
		if char == ',' {
			if buffer != "" {
				params = append(params,buffer)
				buffer = ""
			}
		} else if char > 32 { // 32 is whitespace. Below that are the special characters
			buffer += string(char)
		}
	}
	
	if buffer != "" {
		params = append(params,buffer)
	}
	return params, nil
}
