package smartContract

import (
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type atArgumentParser struct {
	arguments []*big.Int
	code      []byte
}

const atSep = "@"
const base = 16

func NewAtArgumentParser() (process.ArgumentsParser, error) {
	return &atArgumentParser{}, nil
}

// ParseData creates the code and the arguments from the input data
// format: code@arg1@arg2@arg3...
// Until the first @ all the bytes are for the code / function
// after that every argument start with an @
func (at *atArgumentParser) ParseData(data []byte) error {
	splitString := strings.Split(string(data), atSep)
	if len(splitString) == 0 || len(splitString[0]) == 0 {
		return process.ErrStringSplitFailed
	}

	code := []byte(splitString[0])
	arguments := make([]*big.Int, 0)
	for i := 1; i < len(splitString); i++ {
		currArg := new(big.Int)
		currArg, ok := currArg.SetString(splitString[i], base)
		if !ok {
			continue
		}

		arguments = append(arguments, currArg)
	}

	at.code = code
	at.arguments = arguments
	return nil
}

// GetArguments returns the arguments from the parsed data
func (at *atArgumentParser) GetArguments() ([]*big.Int, error) {
	if at.arguments == nil {
		return nil, process.ErrNilArguments
	}
	return at.arguments, nil
}

// GetCode returns the code from the parsed data
func (at *atArgumentParser) GetCode() ([]byte, error) {
	if at.code == nil {
		return nil, process.ErrNilCode
	}
	return at.code, nil
}

// GetFunctons returns the function from the parsed data
func (at *atArgumentParser) GetFunction() (string, error) {
	if at.code == nil {
		return "", process.ErrNilFunction
	}
	return string(at.code), nil
}
