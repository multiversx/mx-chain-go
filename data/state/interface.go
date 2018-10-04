package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

var DefHasher hashing.Hasher
var DefMarsh marshal.Marshalizer

func init() {
	DefHasher = &mock.MockHasher{}
	DefMarsh = &mock.MockMarshalizer{}
}

type Accounter interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	Commit() error
	Undo() error
	Root() []byte
}
