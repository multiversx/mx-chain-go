package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	chainData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	factoryState "github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
)

// TODO: merge this with data components

// StateComponentsFactoryArgs holds the arguments needed for creating a state components factory
type StateComponentsFactoryArgs struct {
	Config           config.Config
	EnableEpochs     config.EnableEpochs
	ShardCoordinator sharding.Coordinator
	Core             CoreComponentsHolder
	StorageService   dataRetriever.StorageService
	ProcessingMode   common.NodeProcessingMode
	ChainHandler     chainData.ChainHandler
}

type stateComponentsFactory struct {
	config           config.Config
	shardCoordinator sharding.Coordinator
	core             CoreComponentsHolder
	storageService   dataRetriever.StorageService
	enableEpochs     config.EnableEpochs
	processingMode   common.NodeProcessingMode
	chainHandler     chainData.ChainHandler
}

// stateComponents struct holds the state components of the Elrond protocol
type stateComponents struct {
	peerAccounts        state.AccountsAdapter
	accountsAdapter     state.AccountsAdapter
	accountsAdapterAPI  state.AccountsAdapter
	triesContainer      common.TriesHolder
	trieStorageManagers map[string]common.StorageManager
}

// NewStateComponentsFactory will return a new instance of stateComponentsFactory
func NewStateComponentsFactory(args StateComponentsFactoryArgs) (*stateComponentsFactory, error) {
	if args.Core == nil {
		return nil, errors.ErrNilCoreComponents
	}
	if check.IfNil(args.Core.Hasher()) {
		return nil, errors.ErrNilHasher
	}
	if check.IfNil(args.Core.InternalMarshalizer()) {
		return nil, errors.ErrNilMarshalizer
	}
	if check.IfNil(args.Core.PathHandler()) {
		return nil, errors.ErrNilPathHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, errors.ErrNilShardCoordinator
	}
	if check.IfNil(args.StorageService) {
		return nil, errors.ErrNilStorageService
	}
	if check.IfNil(args.ChainHandler) {
		return nil, errors.ErrNilBlockChainHandler
	}

	return &stateComponentsFactory{
		config:           args.Config,
		shardCoordinator: args.ShardCoordinator,
		core:             args.Core,
		storageService:   args.StorageService,
		enableEpochs:     args.EnableEpochs,
		processingMode:   args.ProcessingMode,
		chainHandler:     args.ChainHandler,
	}, nil
}

// Create creates the state components
func (scf *stateComponentsFactory) Create() (*stateComponents, error) {
	triesContainer, trieStorageManagers, err := trieFactory.CreateTriesComponentsForShardId(
		scf.config,
		scf.core,
		scf.storageService,
	)
	if err != nil {
		return nil, err
	}

	accountsAdapter, accountsAdapterAPI, err := scf.createAccountsAdapters(triesContainer)
	if err != nil {
		return nil, err
	}

	peerAdapter, err := scf.createPeerAdapter(triesContainer)
	if err != nil {
		return nil, err
	}

	return &stateComponents{
		peerAccounts:        peerAdapter,
		accountsAdapter:     accountsAdapter,
		accountsAdapterAPI:  accountsAdapterAPI,
		triesContainer:      triesContainer,
		trieStorageManagers: trieStorageManagers,
	}, nil
}

func (scf *stateComponentsFactory) createAccountsAdapters(triesContainer common.TriesHolder) (state.AccountsAdapter, state.AccountsAdapter, error) {
	accountFactory := factoryState.NewAccountCreator()
	merkleTrie := triesContainer.Get([]byte(trieFactory.UserAccountTrie))
	storagePruning, err := scf.newStoragePruningManager()
	if err != nil {
		return nil, nil, err
	}

	argsProcessingAccountsDB := state.ArgsAccountsDB{
		Trie:                  merkleTrie,
		Hasher:                scf.core.Hasher(),
		Marshaller:            scf.core.InternalMarshalizer(),
		AccountFactory:        accountFactory,
		StoragePruningManager: storagePruning,
		ProcessingMode:        scf.processingMode,
		ProcessStatusHandler:  scf.core.ProcessStatusHandler(),
	}
	accountsAdapter, err := state.NewAccountsDB(argsProcessingAccountsDB)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	argsAPIAccountsDB := state.ArgsAccountsDB{
		Trie:                  merkleTrie,
		Hasher:                scf.core.Hasher(),
		Marshaller:            scf.core.InternalMarshalizer(),
		AccountFactory:        accountFactory,
		StoragePruningManager: storagePruning,
		ProcessingMode:        scf.processingMode,
		ProcessStatusHandler:  scf.core.ProcessStatusHandler(),
	}
	accountsAdapterAPI, err := state.NewAccountsDB(argsAPIAccountsDB)
	if err != nil {
		return nil, nil, fmt.Errorf("accounts adapter API: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	wrapper, err := state.NewAccountsDBApi(accountsAdapterAPI, scf.chainHandler)
	if err != nil {
		return nil, nil, fmt.Errorf("accounts adapter API: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	return accountsAdapter, wrapper, nil
}

func (scf *stateComponentsFactory) createPeerAdapter(triesContainer common.TriesHolder) (state.AccountsAdapter, error) {
	accountFactory := factoryState.NewPeerAccountCreator()
	merkleTrie := triesContainer.Get([]byte(trieFactory.PeerAccountTrie))
	storagePruning, err := scf.newStoragePruningManager()
	if err != nil {
		return nil, err
	}

	argsProcessingPeerAccountsDB := state.ArgsAccountsDB{
		Trie:                  merkleTrie,
		Hasher:                scf.core.Hasher(),
		Marshaller:            scf.core.InternalMarshalizer(),
		AccountFactory:        accountFactory,
		StoragePruningManager: storagePruning,
		ProcessingMode:        scf.processingMode,
		ProcessStatusHandler:  scf.core.ProcessStatusHandler(),
	}
	peerAdapter, err := state.NewPeerAccountsDB(argsProcessingPeerAccountsDB)
	if err != nil {
		return nil, err
	}

	return peerAdapter, nil
}

func (scf *stateComponentsFactory) newStoragePruningManager() (state.StoragePruningManager, error) {
	args := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: scf.config.EvictionWaitingList.RootHashesSize,
		HashesSize:     scf.config.EvictionWaitingList.HashesSize,
	}
	trieEvictionWaitingList, err := evictionWaitingList.NewMemoryEvictionWaitingList(args)
	if err != nil {
		return nil, err
	}

	storagePruning, err := storagePruningManager.NewStoragePruningManager(
		trieEvictionWaitingList,
		scf.config.TrieStorageManagerConfig.PruningBufferLen,
	)
	if err != nil {
		return nil, err
	}

	return storagePruning, nil
}

// Close closes all underlying components that need closing
func (pc *stateComponents) Close() error {
	errString := ""

	err := pc.accountsAdapter.Close()
	if err != nil {
		errString += fmt.Errorf("accountsAdapter close failed: %w ", err).Error()
	}

	err = pc.accountsAdapterAPI.Close()
	if err != nil {
		errString += fmt.Errorf("accountsAdapterAPI close failed: %w ", err).Error()
	}

	err = pc.peerAccounts.Close()
	if err != nil {
		errString += fmt.Errorf("peerAccounts close failed: %w ", err).Error()
	}

	tries := pc.triesContainer.GetAll()
	for _, trie := range tries {
		err = trie.Close()
		if err != nil {
			errString += fmt.Errorf("trie close failed: %w ", err).Error()
		}
	}

	for _, trieStorageManager := range pc.trieStorageManagers {
		err = trieStorageManager.Close()
		if err != nil {
			errString += fmt.Errorf("trieStorageManager close failed: %w ", err).Error()
		}
	}

	if len(errString) != 0 {
		return fmt.Errorf("state components close failed: %s", errString)
	}
	return nil
}
