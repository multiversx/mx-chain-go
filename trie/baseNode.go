package trie

import (
	"sync"
	
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

type baseNode struct {
	mutex  sync.RWMutex
	hash   []byte
	dirty  bool
	marsh  marshal.Marshalizer
	hasher hashing.Hasher
}

func (bn *baseNode) getHash() []byte {
	//TODO add mutex protection for all methods

	return bn.hash
}

func (bn *baseNode) setGivenHash(hash []byte) {
	bn.hash = hash
}

func (bn *baseNode) isDirty() bool {
	return bn.dirty
}

func (bn *baseNode) setDirty(dirty bool) {
	bn.dirty = dirty
}

func (bn *baseNode) getMarshalizer() marshal.Marshalizer {
	return bn.marsh
}

func (bn *baseNode) setMarshalizer(marshalizer marshal.Marshalizer) {
	bn.marsh = marshalizer
}

func (bn *baseNode) getHasher() hashing.Hasher {
	return bn.hasher
}

func (bn *baseNode) setHasher(hasher hashing.Hasher) {
	bn.hasher = hasher
}
