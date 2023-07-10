package state

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/state"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/state/syncer"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"
)

// TODO: merge this with data components

// StateComponentsFactoryArgs holds the arguments needed for creating a state components factory
type StateComponentsFactoryArgs struct {
	Config                   config.Config
	Core                     factory.CoreComponentsHolder
	StatusCore               factory.StatusCoreComponentsHolder
	StorageService           dataRetriever.StorageService
	ProcessingMode           common.NodeProcessingMode
	ShouldSerializeSnapshots bool
	SnapshotsEnabled         bool
	ChainHandler             chainData.ChainHandler
}

type stateComponentsFactory struct {
	config                   config.Config
	core                     factory.CoreComponentsHolder
	statusCore               factory.StatusCoreComponentsHolder
	storageService           dataRetriever.StorageService
	processingMode           common.NodeProcessingMode
	shouldSerializeSnapshots bool
	snapshotsEnabled         bool
	chainHandler             chainData.ChainHandler
}

// stateComponents struct holds the state components of the MultiversX protocol
type stateComponents struct {
	peerAccounts             state.AccountsAdapter
	accountsAdapter          state.AccountsAdapter
	accountsAdapterAPI       state.AccountsAdapter
	accountsRepository       state.AccountsRepository
	triesContainer           common.TriesHolder
	trieStorageManagers      map[string]common.StorageManager
	missingTrieNodesNotifier common.MissingTrieNodesNotifier
}

// NewStateComponentsFactory will return a new instance of stateComponentsFactory
func NewStateComponentsFactory(args StateComponentsFactoryArgs) (*stateComponentsFactory, error) {
	if check.IfNil(args.Core) {
		return nil, errors.ErrNilCoreComponents
	}
	if check.IfNil(args.StorageService) {
		return nil, errors.ErrNilStorageService
	}
	if check.IfNil(args.StatusCore) {
		return nil, errors.ErrNilStatusCoreComponents
	}

	return &stateComponentsFactory{
		config:                   args.Config,
		core:                     args.Core,
		statusCore:               args.StatusCore,
		storageService:           args.StorageService,
		processingMode:           args.ProcessingMode,
		shouldSerializeSnapshots: args.ShouldSerializeSnapshots,
		chainHandler:             args.ChainHandler,
		snapshotsEnabled:         args.SnapshotsEnabled,
	}, nil
}

// Create creates the state components
func (scf *stateComponentsFactory) Create() (*stateComponents, error) {
	triesContainer, trieStorageManagers, err := trieFactory.CreateTriesComponentsForShardId(
		scf.snapshotsEnabled,
		scf.config,
		scf.core,
		scf.storageService,
	)
	if err != nil {
		return nil, err
	}

	accountsAdapter, accountsAdapterAPI, accountsRepository, err := scf.createAccountsAdapters(triesContainer)
	if err != nil {
		return nil, err
	}

	peerAdapter, err := scf.createPeerAdapter(triesContainer)
	if err != nil {
		return nil, err
	}

	return &stateComponents{
		peerAccounts:             peerAdapter,
		accountsAdapter:          accountsAdapter,
		accountsAdapterAPI:       accountsAdapterAPI,
		accountsRepository:       accountsRepository,
		triesContainer:           triesContainer,
		trieStorageManagers:      trieStorageManagers,
		missingTrieNodesNotifier: syncer.NewMissingTrieNodesNotifier(),
	}, nil
}

func (scf *stateComponentsFactory) createAccountsAdapters(triesContainer common.TriesHolder) (state.AccountsAdapter, state.AccountsAdapter, state.AccountsRepository, error) {
	argsAccCreator := factoryState.ArgsAccountCreator{
		Hasher:              scf.core.Hasher(),
		Marshaller:          scf.core.InternalMarshalizer(),
		EnableEpochsHandler: scf.core.EnableEpochsHandler(),
	}
	accountFactory, err := factoryState.NewAccountCreator(argsAccCreator)
	if err != nil {
		return nil, nil, nil, err
	}

	merkleTrie := triesContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	storagePruning, err := scf.newStoragePruningManager()
	if err != nil {
		return nil, nil, nil, err
	}

	argsProcessingAccountsDB := state.ArgsAccountsDB{
		Trie:                     merkleTrie,
		Hasher:                   scf.core.Hasher(),
		Marshaller:               scf.core.InternalMarshalizer(),
		AccountFactory:           accountFactory,
		StoragePruningManager:    storagePruning,
		ProcessingMode:           scf.processingMode,
		ShouldSerializeSnapshots: scf.shouldSerializeSnapshots,
		ProcessStatusHandler:     scf.core.ProcessStatusHandler(),
		AppStatusHandler:         scf.statusCore.AppStatusHandler(),
		AddressConverter:         scf.core.AddressPubKeyConverter(),
	}
	accountsAdapter, err := state.NewAccountsDB(argsProcessingAccountsDB)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	argsAPIAccountsDB := state.ArgsAccountsDB{
		Trie:                  merkleTrie,
		Hasher:                scf.core.Hasher(),
		Marshaller:            scf.core.InternalMarshalizer(),
		AccountFactory:        accountFactory,
		StoragePruningManager: storagePruning,
		ProcessingMode:        scf.processingMode,
		ProcessStatusHandler:  scf.core.ProcessStatusHandler(),
		AppStatusHandler:      scf.statusCore.AppStatusHandler(),
		AddressConverter:      scf.core.AddressPubKeyConverter(),
	}

	accountsAdapterApiOnFinal, err := factoryState.CreateAccountsAdapterAPIOnFinal(argsAPIAccountsDB, scf.chainHandler)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("accounts adapter API on final: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	accountsAdapterApiOnCurrent, err := factoryState.CreateAccountsAdapterAPIOnCurrent(argsAPIAccountsDB, scf.chainHandler)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("accounts adapter API on current: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	accountsAdapterApiOnHistorical, err := factoryState.CreateAccountsAdapterAPIOnHistorical(argsAPIAccountsDB)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("accounts adapter API on historical: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	argsAccountsRepository := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accountsAdapterApiOnFinal,
		CurrentStateAccountsWrapper:    accountsAdapterApiOnCurrent,
		HistoricalStateAccountsWrapper: accountsAdapterApiOnHistorical,
	}

	accountsRepository, err := state.NewAccountsRepository(argsAccountsRepository)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("accountsRepository: %w", err)
	}

	return accountsAdapter, accountsRepository.GetCurrentStateAccountsWrapper(), accountsRepository, nil
}

func (scf *stateComponentsFactory) createPeerAdapter(triesContainer common.TriesHolder) (state.AccountsAdapter, error) {
	accountFactory := factoryState.NewPeerAccountCreator()
	merkleTrie := triesContainer.Get([]byte(dataRetriever.PeerAccountsUnit.String()))
	storagePruning, err := scf.newStoragePruningManager()
	if err != nil {
		return nil, err
	}

	argsProcessingPeerAccountsDB := state.ArgsAccountsDB{
		Trie:                     merkleTrie,
		Hasher:                   scf.core.Hasher(),
		Marshaller:               scf.core.InternalMarshalizer(),
		AccountFactory:           accountFactory,
		StoragePruningManager:    storagePruning,
		ProcessingMode:           scf.processingMode,
		ShouldSerializeSnapshots: scf.shouldSerializeSnapshots,
		ProcessStatusHandler:     scf.core.ProcessStatusHandler(),
		AppStatusHandler:         scf.statusCore.AppStatusHandler(),
		AddressConverter:         scf.core.AddressPubKeyConverter(),
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

	err = pc.accountsRepository.Close()
	if err != nil {
		errString += fmt.Errorf("accountsRepository close failed: %w ", err).Error()
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
