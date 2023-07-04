package parsers

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
)

type mainTrieLeafParser struct {
}

// NewMainTrieLeafParser creates a new instance of mainTrieLeafParser
func NewMainTrieLeafParser() *mainTrieLeafParser {
	return &mainTrieLeafParser{}
}

// ParseLeaf returns the given key an value as a KeyValStorage
func (tlp *mainTrieLeafParser) ParseLeaf(trieKey []byte, trieVal []byte, _ core.TrieNodeVersion) (core.KeyValueHolder, error) {
	return keyValStorage.NewKeyValStorage(trieKey, trieVal), nil
}

// IsInterfaceNil returns nil if there is no value under the interface
func (tlp *mainTrieLeafParser) IsInterfaceNil() bool {
	return tlp == nil
}
