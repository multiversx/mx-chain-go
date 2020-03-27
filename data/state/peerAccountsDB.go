package state

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// PeerAccountsDB will save and synchronize data from peer processor, plus will synchronize with nodesCoordinator
type PeerAccountsDB struct {
	*AccountsDB
}

// NewPeerAccountsDB creates a new account manager
func NewPeerAccountsDB(
	trie data.Trie,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory AccountFactory,
) (*PeerAccountsDB, error) {
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(accountFactory) {
		return nil, ErrNilAccountFactory
	}

	return &PeerAccountsDB{
		&AccountsDB{
			mainTrie:       trie,
			hasher:         hasher,
			marshalizer:    marshalizer,
			accountFactory: accountFactory,
			entries:        make([]JournalEntry, 0),
			dataTries:      NewDataTriesHolder(),
			mutOp:          sync.RWMutex{},
		},
	}, nil
}
