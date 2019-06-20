package smartContract

import (
	"github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/process"
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

// GetSeparator returns the separator used for parsing the data
func (at *atArgumentParser) GetSeparator() string {
	return atSep
}

// GetStorageUpdates parse data into storage updates
func (at *atArgumentParser) GetStorageUpdates(data []byte) ([]*vmcommon.StorageUpdate, error) {
	splitString := strings.Split(string(data), atSep)
	if len(splitString) == 0 || len(splitString[0]) == 0 {
		return nil, process.ErrStringSplitFailed
	}

	if len(splitString)%2 != 0 {
		return nil, process.ErrInvalidDataInput
	}

	storageUpdates := make([]*vmcommon.StorageUpdate, 0)
	for i := 0; i < len(splitString); i += 2 {
		storageUpdate := &vmcommon.StorageUpdate{Offset: []byte(splitString[i]), Data: []byte(splitString[i+1])}
		storageUpdates = append(storageUpdates, storageUpdate)
	}

	return storageUpdates, nil
}

// CreateDataFromStorageUpdate creates storage update from data
func (at *atArgumentParser) CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) []byte {
	data := []byte{}
	for i := 0; i < len(storageUpdates); i++ {
		storageUpdate := storageUpdates[i]
		data = append(data, storageUpdate.Offset...)
		data = append(data, []byte(at.GetSeparator())...)
		data = append(data, storageUpdate.Data...)

		if i < len(storageUpdates)-1 {
			data = append(data, []byte(at.GetSeparator())...)
		}
	}
	return data
}
