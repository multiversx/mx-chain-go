package parsers

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/keyValStorage"
)

type trieLeafParserV1 struct {
}

// NewTrieLeafParserV1 creates a new instance of trieLeafParserV1
func NewTrieLeafParserV1() *trieLeafParserV1 {
	return &trieLeafParserV1{}
}

// ParseLeaf returns the given key an value as a KeyValStorage
func (tlp *trieLeafParserV1) ParseLeaf(trieKey []byte, trieVal []byte) (core.KeyValueHolder, error) {
	return keyValStorage.NewKeyValStorage(trieKey, trieVal), nil
}

// IsInterfaceNil returns nil if there is no value under the interface
func (tlp *trieLeafParserV1) IsInterfaceNil() bool {
	return tlp == nil
}
