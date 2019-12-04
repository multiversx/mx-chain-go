package smartContract

import (
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-vm-common"
)

type atArgumentParser struct {
	arguments [][]byte
	code      []byte
}

const atSep = "@"

// NewAtArgumentParser creates a new argument parser implementation that splits arguments by @ character
func NewAtArgumentParser() (process.ArgumentsParser, error) {
	return &atArgumentParser{}, nil
}

// ParseData creates the code and the arguments from the input data
// format: code@arg1@arg2@arg3...
// Until the first @ all the bytes are for the code / function
// after that every argument start with an @
func (at *atArgumentParser) ParseData(data string) error {
	splitString := strings.Split(data, atSep)
	if len(splitString) == 0 || len(splitString[0]) == 0 {
		return process.ErrStringSplitFailed
	}

	var err error
	code := []byte(splitString[0])
	arguments := make([][]byte, len(splitString)-1)
	for i := 1; i < len(splitString); i++ {
		arguments[i-1], err = hex.DecodeString(splitString[i])
		if err != nil {
			return err
		}
	}

	at.code = code
	at.arguments = arguments
	return nil
}

// GetArguments returns the arguments from the parsed data
func (at *atArgumentParser) GetArguments() ([][]byte, error) {
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

// GetFunction returns the function from the parsed data
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
func (at *atArgumentParser) GetStorageUpdates(data string) ([]*vmcommon.StorageUpdate, error) {
	splitString := strings.Split(data, atSep)
	if len(splitString) == 0 || len(splitString[0]) == 0 {
		return nil, process.ErrStringSplitFailed
	}

	if len(splitString)%2 != 0 {
		return nil, process.ErrInvalidDataInput
	}

	storageUpdates := make([]*vmcommon.StorageUpdate, 0)
	for i := 0; i < len(splitString); i += 2 {
		offset, err := hex.DecodeString(splitString[i])
		if err != nil {
			return nil, err
		}

		value, err := hex.DecodeString(splitString[i+1])
		if err != nil {
			return nil, err
		}

		storageUpdate := &vmcommon.StorageUpdate{Offset: offset, Data: value}
		storageUpdates = append(storageUpdates, storageUpdate)
	}

	return storageUpdates, nil
}

// CreateDataFromStorageUpdate creates storage update from data
func (at *atArgumentParser) CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) string {
	data := ""
	for i := 0; i < len(storageUpdates); i++ {
		storageUpdate := storageUpdates[i]
		data = data + hex.EncodeToString(storageUpdate.Offset)
		data = data + at.GetSeparator()
		data = data + hex.EncodeToString(storageUpdate.Data)

		if i < len(storageUpdates)-1 {
			data = data + at.GetSeparator()
		}
	}
	return data
}

// IsInterfaceNil returns true if there is no value under the interface
func (at *atArgumentParser) IsInterfaceNil() bool {
	if at == nil {
		return true
	}
	return false
}
