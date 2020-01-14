package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgsNewActiveTrieContainerFactory the input arguments for the factory
type ArgsNewActiveTrieContainerFactory struct {
	PeerAccounts 	 state.AccountsAdapter
	UserAccounts 	 state.AccountsAdapter
	ShardCoordinator sharding.Coordinator
}

type activeTrieContainerFactory struct {
	peerAccounts state.AccountsAdapter
	userAccounts state.AccountsAdapter
	shardCoordinator sharding.Coordinator
}

// NewActiveTrieContainerFactory create a new active trie container factory
func NewActiveTrieContainerFactory(args ArgsNewActiveTrieContainerFactory) (*activeTrieContainerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, data.ErrNilShardCoordinator
	}
	if check.IfNil(args.PeerAccounts) {
		return nil, data.ErrNilPeerAccounts
	}
	if check.IfNil(args.UserAccounts) {
		return nil, data.ErrNilUserAccounts
	}

	a := &activeTrieContainerFactory{
		peerAccounts:     args.PeerAccounts,
		userAccounts:     args.UserAccounts,
		shardCoordinator: args.ShardCoordinator,
	}

	return a, nil
}

func (a *activeTrieContainerFactory) Create() (state.TriesHolder, error) {
	activeTries := state.NewDataTriesHolder()

	userAccountTrie := a.userAccounts.
}

func (a *activeTrieContainerFactory) IsInterfaceNil() bool {
	return a == nil
}
