package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	factoryState "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

//TODO: merge this with data components

// StateComponentsFactoryArgs holds the arguments needed for creating a state components factory
type StateComponentsFactoryArgs struct {
	Config           config.Config
	ShardCoordinator sharding.Coordinator
	Core             *CoreComponents
	Tries            *TriesComponents
	PathManager      storage.PathManagerHandler
}

type stateComponentsFactory struct {
	config           config.Config
	shardCoordinator sharding.Coordinator
	core             *CoreComponents
	tries            *TriesComponents
	pathManager      storage.PathManagerHandler
}

// NewStateComponentsFactory will return a new instance of stateComponentsFactory
func NewStateComponentsFactory(args StateComponentsFactoryArgs) (*stateComponentsFactory, error) {
	if check.IfNil(args.PathManager) {
		return nil, ErrNilPathManager
	}
	if args.Core == nil {
		return nil, ErrNilCoreComponents
	}
	if args.Tries == nil {
		return nil, ErrNilTriesComponents
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}

	return &stateComponentsFactory{
		config:           args.Config,
		core:             args.Core,
		tries:            args.Tries,
		pathManager:      args.PathManager,
		shardCoordinator: args.ShardCoordinator,
	}, nil
}

// Create creates the state components
func (scf *stateComponentsFactory) Create() (*StateComponents, error) {
	processPubkeyConverter, err := factoryState.NewPubkeyConverter(scf.config.AddressPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for ProcessPubkeyConverter: %s", ErrPubKeyConverterCreation, err.Error())
	}

	validatorPubkeyConverter, err := factoryState.NewPubkeyConverter(scf.config.ValidatorPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for ValidatorPubkeyConverter: %s", ErrPubKeyConverterCreation, err.Error())
	}

	accountFactory := factoryState.NewAccountCreator()
	merkleTrie := scf.tries.TriesContainer.Get([]byte(factory.UserAccountTrie))
	accountsAdapter, err := state.NewAccountsDB(merkleTrie, scf.core.Hasher, scf.core.InternalMarshalizer, accountFactory)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrAccountsAdapterCreation, err.Error())
	}

	accountsAdapterAPI, err := state.NewAccountsDB(merkleTrie, scf.core.Hasher, scf.core.InternalMarshalizer, accountFactory)
	if err != nil {
		return nil, fmt.Errorf("accounts adapter API: %w: %s", ErrAccountsAdapterCreation, err.Error())
	}

	accountFactory = factoryState.NewPeerAccountCreator()
	merkleTrie = scf.tries.TriesContainer.Get([]byte(factory.PeerAccountTrie))
	peerAdapter, err := state.NewPeerAccountsDB(merkleTrie, scf.core.Hasher, scf.core.InternalMarshalizer, accountFactory)
	if err != nil {
		return nil, err
	}

	return &StateComponents{
		PeerAccounts:             peerAdapter,
		AddressPubkeyConverter:   processPubkeyConverter,
		ValidatorPubkeyConverter: validatorPubkeyConverter,
		AccountsAdapter:          accountsAdapter,
		AccountsAdapterAPI:       accountsAdapterAPI,
	}, nil
}
