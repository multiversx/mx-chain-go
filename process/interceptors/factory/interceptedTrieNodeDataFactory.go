package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

type interceptedTrieNodeDataFactory struct {
	trie        data.Trie
	marshalizer marshal.Marshalizer
	hasher      hashing.Hasher
}

// NewInterceptedTrieNodeDataFactory creates an instance of interceptedTrieNodeDataFactory
func NewInterceptedTrieNodeDataFactory(
	argument *ArgInterceptedDataFactory,
) (*interceptedTrieNodeDataFactory, error) {

	if argument == nil {
		return nil, process.ErrNilArguments
	}
	if check.IfNil(argument.Trie) {
		return nil, process.ErrNilTrie
	}
	if check.IfNil(argument.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.Hasher) {
		return nil, process.ErrNilHasher
	}

	return &interceptedTrieNodeDataFactory{
		trie:        argument.Trie,
		marshalizer: argument.Marshalizer,
		hasher:      argument.Hasher,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (sidf *interceptedTrieNodeDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	return trie.NewInterceptedTrieNode(buff, sidf.trie.Database(), sidf.marshalizer, sidf.hasher)

}

// IsInterfaceNil returns true if there is no value under the interface
func (sidf *interceptedTrieNodeDataFactory) IsInterfaceNil() bool {
	if sidf == nil {
		return true
	}
	return false
}
