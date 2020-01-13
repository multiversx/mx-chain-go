package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
	containers "github.com/ElrondNetwork/elrond-go/update/container"
)

type ArgsNewActiveAccountHandlersFactory struct {
	PeerAccounts     state.AccountsAdapter
	AccountsAdapter  state.AccountsAdapter
	ShardCoordinator sharding.Coordinator
}

type activeAccountsContainerFactory struct {
	peerAccounts     state.AccountsAdapter
	accountsAdapter  state.AccountsAdapter
	shardCoordinator sharding.Coordinator
}

func NewActiveAccountsContainerFactory(args ArgsNewActiveAccountHandlersFactory) (*activeAccountsContainerFactory, error) {
	if check.IfNil(args.PeerAccounts) {
		return nil, process.ErrNilPeerAccountsAdapter
	}
	if check.IfNil(args.AccountsAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, sharding.ErrNilShardCoordinator
	}

	a := &activeAccountsContainerFactory{
		peerAccounts:     args.PeerAccounts,
		accountsAdapter:  args.AccountsAdapter,
		shardCoordinator: args.ShardCoordinator,
	}

	return a, nil
}

func (a *activeAccountsContainerFactory) Create() (update.AccountsHandlerContainer, error) {
	container := containers.NewAccountsHandlerContainer()

	accAdapterIdentifier := update.CreateTrieIdentifier(a.shardCoordinator.SelfId(), factory.UserAccount)
	err := container.Add(accAdapterIdentifier, a.accountsAdapter)
	if err != nil {
		return nil, err
	}

	accAdapterIdentifier = update.CreateTrieIdentifier(a.shardCoordinator.SelfId(), factory.ValidatorAccount)
	err = container.Add(accAdapterIdentifier, a.peerAccounts)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (a *activeAccountsContainerFactory) IsInterfaceNil() bool {
	return a == nil
}
