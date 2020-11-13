package parsers

import (
	"encoding/hex"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type storageUpdatesParser struct {
}

// NewStorageUpdatesParser creates a new parser
func NewStorageUpdatesParser() *storageUpdatesParser {
	return &storageUpdatesParser{}
}

// GetStorageUpdates parse data into storage updates
func (parser *storageUpdatesParser) GetStorageUpdates(data string) ([]*vmcommon.StorageUpdate, error) {
	data = trimLeadingSeparatorChar(data)

	tokens, err := tokenize(data)
	if err != nil {
		return nil, err
	}
	err = requireNumTokensIsEven(tokens)
	if err != nil {
		return nil, err
	}

	storageUpdates := make([]*vmcommon.StorageUpdate, 0, len(tokens))
	for i := 0; i < len(tokens); i += 2 {
		offset, err := decodeToken(tokens[i])
		if err != nil {
			return nil, err
		}

		value, err := decodeToken(tokens[i+1])
		if err != nil {
			return nil, err
		}

		storageUpdate := &vmcommon.StorageUpdate{Offset: offset, Data: value}
		storageUpdates = append(storageUpdates, storageUpdate)
	}

	return storageUpdates, nil
}

// CreateDataFromStorageUpdate creates storage update from data
func (parser *storageUpdatesParser) CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) string {
	data := ""
	for i := 0; i < len(storageUpdates); i++ {
		storageUpdate := storageUpdates[i]
		data = data + hex.EncodeToString(storageUpdate.Offset)
		data = data + atSeparator
		data = data + hex.EncodeToString(storageUpdate.Data)

		if i < len(storageUpdates)-1 {
			data = data + atSeparator
		}
	}
	return data
}

// IsInterfaceNil returns true if there is no value under the interface
func (parser *storageUpdatesParser) IsInterfaceNil() bool {
	return parser == nil
}
