package abiDecoder

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type FuncSigTypes struct {
	Type string `json:"type"`
}

func ParseFunctionSignature(signature string) ([]FunctionSignature, error) {
	//TODO: multiple nested struct
	// the parser method is a bit too simple

	re := regexp.MustCompile(`(\w+)\((.*)\)`)
	matches := re.FindStringSubmatch(signature)

	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid function signature")
	}

	functionName := matches[1]
	paramsStr := matches[2]

	var inputs []FuncSigTypes
	if paramsStr == "" {
		return []FunctionSignature{{
			Type:   "function",
			Name:   functionName,
			Inputs: inputs,
		}}, nil
	}

	params := strings.Split(paramsStr, ",")
	for _, param := range params {
		trimmedParam := strings.TrimSpace(param)
		inputs = append(inputs, FuncSigTypes{
			Type: trimmedParam,
		})
	}

	return []FunctionSignature{{
		Type:   "function",
		Name:   functionName,
		Inputs: inputs,
	}}, nil
}

type FunctionSignature struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Inputs []FuncSigTypes `json:"inputs,omitempty"`
}

func ABIByFuncSig(funcSig string) (*ABIDec, error) {
	abis, err := ParseFunctionSignature(funcSig)
	if err != nil {
		return nil, fmt.Errorf("error parsing function signature: %v", err)
	}
	jsonABI, err := json.Marshal(abis)

	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON for function signature: %v", err)
	}

	decoder, err := SetABI(string(jsonABI))
	if err != nil {
		// TODO: 类型处理
		return nil, fmt.Errorf("invalid ABI by function signature: %v", err)
	}
	return decoder, nil
}
