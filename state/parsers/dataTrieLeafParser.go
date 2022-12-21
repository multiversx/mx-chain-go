package parsers

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/state/dataTrieValue"
)

type dataTrieLeafParser struct {
	address             []byte
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
}

// NewDataTrieLeafParser returns a new instance of dataTrieLeafParser
func NewDataTrieLeafParser(address []byte, marshaller marshal.Marshalizer, enableEpochsHandler common.EnableEpochsHandler) (*dataTrieLeafParser, error) {
	if check.IfNil(marshaller) {
		return nil, errors.ErrNilMarshalizer
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}

	return &dataTrieLeafParser{
		address:             address,
		marshaller:          marshaller,
		enableEpochsHandler: enableEpochsHandler,
	}, nil
}

// ParseLeaf returns a new KeyValStorage with the actual key and value
func (tlp *dataTrieLeafParser) ParseLeaf(trieKey []byte, trieVal []byte) (core.KeyValueHolder, error) {
	if tlp.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		data := &dataTrieValue.TrieLeafData{}
		err := tlp.marshaller.Unmarshal(data, trieVal)
		if err == nil && !isEmptyTrieData(data) {
			return keyValStorage.NewKeyValStorage(data.Key, data.Value), nil
		}
	}

	suffix := append(trieKey, tlp.address...)
	value, err := common.TrimSuffixFromValue(trieVal, len(suffix))
	if err != nil {
		return nil, err
	}

	return keyValStorage.NewKeyValStorage(trieKey, value), nil
}

// TODO remove this after proper marshaller fix
func isEmptyTrieData(data *dataTrieValue.TrieLeafData) bool {
	if data == nil {
		return true
	}

	if len(data.Value) == 0 && len(data.Key) == 0 && len(data.Address) == 0 {
		return true
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (tlp *dataTrieLeafParser) IsInterfaceNil() bool {
	return tlp == nil
}
