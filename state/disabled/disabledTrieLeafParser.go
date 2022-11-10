package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/keyValStorage"
)

type disabledTrieLeafParser struct {
}

// NewDisabledTrieLeafParser creates a new instance of disabledTrieLeafParser
func NewDisabledTrieLeafParser() *disabledTrieLeafParser {
	return &disabledTrieLeafParser{}
}

// ParseLeaf returns the given key an value as a KeyValStorage
func (tlp *disabledTrieLeafParser) ParseLeaf(trieKey []byte, trieVal []byte) (core.KeyValueHolder, error) {
	return keyValStorage.NewKeyValStorage(trieKey, trieVal), nil
}

// IsInterfaceNil returns nil if there is no value under the interface
func (tlp *disabledTrieLeafParser) IsInterfaceNil() bool {
	return tlp == nil
}
