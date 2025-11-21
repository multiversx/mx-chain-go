package state

import (
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/state/lastSnapshotMarker"
	"github.com/multiversx/mx-chain-go/state/stateAccesses"
	"github.com/multiversx/mx-chain-go/state/stateMetrics"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/state/syncer"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever"
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
	ChainHandler             chainData.ChainHandler
}

type stateComponentsFactory struct {
	config                   config.Config
	core                     factory.CoreComponentsHolder
	statusCore               factory.StatusCoreComponentsHolder
	storageService           dataRetriever.StorageService
	processingMode           common.NodeProcessingMode
	shouldSerializeSnapshots bool
	chainHandler             chainData.ChainHandler
}

// stateComponents struct holds the state components of the MultiversX protocol
type stateComponents struct {
	peerAccounts             state.AccountsAdapter
	accountsAdapter          state.AccountsAdapter
	accountsAdapterAPI       state.AccountsAdapter
	accountsAdapterProposal  state.AccountsAdapter
	accountsRepository       state.AccountsRepository
	triesContainer           common.TriesHolder
	trieStorageManagers      map[string]common.StorageManager
	missingTrieNodesNotifier common.MissingTrieNodesNotifier
	trieLeavesRetriever      common.TrieLeavesRetriever
	stateAccessesCollector   state.StateAccessesCollector
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
	}, nil
}

type accountsAdapterCreationResult struct {
	accountsAdapter         state.AccountsAdapter
	accountsAdapterAPI      state.AccountsAdapter
	accountsAdapterProposal state.AccountsAdapter
	accountsRepository      state.AccountsRepository
}

// Create creates the state components
func (scf *stateComponentsFactory) Create() (*stateComponents, error) {
	triesContainer, trieStorageManagers, err := trieFactory.CreateTriesComponentsForShardId(
		scf.config,
		scf.core,
		scf.storageService,
		scf.statusCore.StateStatsHandler(),
	)
	if err != nil {
		return nil, err
	}

	stateAccessesCollector, err := scf.createStateAccessesCollector()
	if err != nil {
		return nil, err
	}

	accCreationResult, err := scf.createAccountsAdapters(triesContainer, stateAccessesCollector)
	if err != nil {
		return nil, err
	}

	peerAdapter, err := scf.createPeerAdapter(triesContainer)
	if err != nil {
		return nil, err
	}

	trieLeavesRetriever, err := scf.createTrieLeavesRetriever(trieStorageManagers[dataRetriever.UserAccountsUnit.String()])
	if err != nil {
		return nil, err
	}

	return &stateComponents{
		peerAccounts:             peerAdapter,
		accountsAdapter:          accCreationResult.accountsAdapter,
		accountsAdapterAPI:       accCreationResult.accountsAdapterAPI,
		accountsAdapterProposal:  accCreationResult.accountsAdapterProposal,
		accountsRepository:       accCreationResult.accountsRepository,
		triesContainer:           triesContainer,
		trieStorageManagers:      trieStorageManagers,
		missingTrieNodesNotifier: syncer.NewMissingTrieNodesNotifier(),
		trieLeavesRetriever:      trieLeavesRetriever,
		stateAccessesCollector:   stateAccessesCollector,
	}, nil
}

func (scf *stateComponentsFactory) createTrieLeavesRetriever(trieStorage common.TrieStorageInteractor) (common.TrieLeavesRetriever, error) {
	if !scf.config.TrieLeavesRetrieverConfig.Enabled {
		return leavesRetriever.NewDisabledLeavesRetriever(), nil
	}

	return leavesRetriever.NewLeavesRetriever(
		trieStorage,
		scf.core.InternalMarshalizer(),
		scf.core.Hasher(),
		scf.config.TrieLeavesRetrieverConfig.MaxSizeInBytes,
	)
}

func (scf *stateComponentsFactory) createStateAccessesCollector() (state.StateAccessesCollector, error) {
	if len(scf.config.StateAccessesCollectorConfig.TypesToCollect) == 0 {
		return disabled.NewDisabledStateAccessesCollector(), nil
	}

	collectRead, collectWrite, err := parseStateChangesTypesToCollect(scf.config.StateAccessesCollectorConfig.TypesToCollect)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state changes types to collect: %w", err)
	}

	var opts []stateAccesses.CollectorOption
	if collectRead {
		opts = append(opts, stateAccesses.WithCollectRead())
	}
	if collectWrite {
		opts = append(opts, stateAccesses.WithCollectWrite())
	}
	if scf.config.StateAccessesCollectorConfig.WithAccountChanges {
		opts = append(opts, stateAccesses.WithAccountChanges())
	}

	storer, err := scf.getStorerForCollector()
	if err != nil {
		return nil, err
	}

	return stateAccesses.NewCollector(storer, opts...)
}

func (scf *stateComponentsFactory) getStorerForCollector() (state.StateAccessesStorer, error) {
	if !scf.config.StateAccessesCollectorConfig.SaveToStorage {
		return disabled.NewDisabledStateAccessesStorer(), nil
	}

	storer, err := scf.storageService.GetStorer(dataRetriever.StateAccessesUnit)
	if err != nil {
		return nil, fmt.Errorf("failed to get storer for state accesses: %w", err)
	}

	return stateAccesses.NewStateAccessesStorer(storer, scf.core.InternalMarshalizer())
}

func (scf *stateComponentsFactory) createSnapshotManager(
	accountFactory state.AccountFactory,
	stateMetrics state.StateMetrics,
	iteratorChannelsProvider state.IteratorChannelsProvider,
) (state.SnapshotsManager, error) {
	if !scf.config.StateTriesConfig.SnapshotsEnabled {
		return disabled.NewDisabledSnapshotsManager(), nil
	}

	argsSnapshotsManager := state.ArgsNewSnapshotsManager{
		ShouldSerializeSnapshots: scf.shouldSerializeSnapshots,
		ProcessingMode:           scf.processingMode,
		Marshaller:               scf.core.InternalMarshalizer(),
		AddressConverter:         scf.core.AddressPubKeyConverter(),
		ProcessStatusHandler:     scf.core.ProcessStatusHandler(),
		StateMetrics:             stateMetrics,
		ChannelsProvider:         iteratorChannelsProvider,
		AccountFactory:           accountFactory,
		LastSnapshotMarker:       lastSnapshotMarker.NewLastSnapshotMarker(),
		StateStatsHandler:        scf.statusCore.StateStatsHandler(),
	}
	return state.NewSnapshotsManager(argsSnapshotsManager)
}

func (scf *stateComponentsFactory) createAccountsAdapters(triesContainer common.TriesHolder, StateAccessesCollector state.StateAccessesCollector) (*accountsAdapterCreationResult, error) {
	argsAccCreator := factoryState.ArgsAccountCreator{
		Hasher:                 scf.core.Hasher(),
		Marshaller:             scf.core.InternalMarshalizer(),
		EnableEpochsHandler:    scf.core.EnableEpochsHandler(),
		StateAccessesCollector: StateAccessesCollector,
	}
	accountFactory, err := factoryState.NewAccountCreator(argsAccCreator)
	if err != nil {
		return nil, err
	}

	merkleTrie := triesContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	storagePruning, err := scf.newStoragePruningManager()
	if err != nil {
		return nil, err
	}

	argStateMetrics := stateMetrics.ArgsStateMetrics{
		SnapshotInProgressKey:   common.MetricAccountsSnapshotInProgress,
		LastSnapshotDurationKey: common.MetricLastAccountsSnapshotDurationSec,
		SnapshotMessage:         stateMetrics.UserTrieSnapshotMsg,
	}
	sm, err := stateMetrics.NewStateMetrics(argStateMetrics, scf.statusCore.AppStatusHandler())
	if err != nil {
		return nil, err
	}

	snapshotsManager, err := scf.createSnapshotManager(accountFactory, sm, iteratorChannelsProvider.NewUserStateIteratorChannelsProvider())
	if err != nil {
		return nil, err
	}

	argsProcessingAccountsDB := state.ArgsAccountsDB{
		Trie:                   merkleTrie,
		Hasher:                 scf.core.Hasher(),
		Marshaller:             scf.core.InternalMarshalizer(),
		AccountFactory:         accountFactory,
		StoragePruningManager:  storagePruning,
		AddressConverter:       scf.core.AddressPubKeyConverter(),
		SnapshotsManager:       snapshotsManager,
		StateAccessesCollector: StateAccessesCollector,
	}
	accountsAdapter, err := state.NewAccountsDB(argsProcessingAccountsDB)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	argsAPIAccCreator := factoryState.ArgsAccountCreator{
		Hasher:                 scf.core.Hasher(),
		Marshaller:             scf.core.InternalMarshalizer(),
		EnableEpochsHandler:    scf.core.EnableEpochsHandler(),
		StateAccessesCollector: disabled.NewDisabledStateAccessesCollector(),
	}
	accountFactoryAPI, err := factoryState.NewAccountCreator(argsAPIAccCreator)
	if err != nil {
		return nil, err
	}

	argsAPIAccountsDB := state.ArgsAccountsDB{
		Trie:                   merkleTrie,
		Hasher:                 scf.core.Hasher(),
		Marshaller:             scf.core.InternalMarshalizer(),
		AccountFactory:         accountFactoryAPI,
		StoragePruningManager:  storagePruning,
		AddressConverter:       scf.core.AddressPubKeyConverter(),
		SnapshotsManager:       disabled.NewDisabledSnapshotsManager(),
		StateAccessesCollector: disabled.NewDisabledStateAccessesCollector(),
	}

	accountsAdapterApiOnFinal, err := factoryState.CreateAccountsAdapterAPIOnFinal(argsAPIAccountsDB, scf.chainHandler)
	if err != nil {
		return nil, fmt.Errorf("accounts adapter API on final: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	accountsAdapterApiOnCurrent, err := factoryState.CreateAccountsAdapterAPIOnCurrent(argsAPIAccountsDB, scf.chainHandler)
	if err != nil {
		return nil, fmt.Errorf("accounts adapter API on current: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	accountsAdapterApiOnHistorical, err := factoryState.CreateAccountsAdapterAPIOnHistorical(argsAPIAccountsDB)
	if err != nil {
		return nil, fmt.Errorf("accounts adapter API on historical: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	argsAccountsRepository := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accountsAdapterApiOnFinal,
		CurrentStateAccountsWrapper:    accountsAdapterApiOnCurrent,
		HistoricalStateAccountsWrapper: accountsAdapterApiOnHistorical,
	}

	accountsRepository, err := state.NewAccountsRepository(argsAccountsRepository)
	if err != nil {
		return nil, fmt.Errorf("accountsRepository: %w", err)
	}

	accountsProposal, err := state.NewAccountsDB(argsAPIAccountsDB)
	if err != nil {
		return nil, fmt.Errorf("%w in state.NewAccountsDB", err)
	}

	response := &accountsAdapterCreationResult{
		accountsAdapter:         accountsAdapter,
		accountsAdapterAPI:      accountsRepository.GetCurrentStateAccountsWrapper(),
		accountsAdapterProposal: accountsProposal,
		accountsRepository:      accountsRepository,
	}

	return response, nil
}

func (scf *stateComponentsFactory) createPeerAdapter(triesContainer common.TriesHolder) (state.AccountsAdapter, error) {
	accountFactory := factoryState.NewPeerAccountCreator()
	merkleTrie := triesContainer.Get([]byte(dataRetriever.PeerAccountsUnit.String()))
	storagePruning, err := scf.newStoragePruningManager()
	if err != nil {
		return nil, err
	}

	argStateMetrics := stateMetrics.ArgsStateMetrics{
		SnapshotInProgressKey:   common.MetricPeersSnapshotInProgress,
		LastSnapshotDurationKey: common.MetricLastPeersSnapshotDurationSec,
		SnapshotMessage:         stateMetrics.PeerTrieSnapshotMsg,
	}
	sm, err := stateMetrics.NewStateMetrics(argStateMetrics, scf.statusCore.AppStatusHandler())
	if err != nil {
		return nil, err
	}

	snapshotManager, err := scf.createSnapshotManager(accountFactory, sm, iteratorChannelsProvider.NewPeerStateIteratorChannelsProvider())
	if err != nil {
		return nil, err
	}

	// TODO: also collect state accesses for the peer trie
	argsProcessingPeerAccountsDB := state.ArgsAccountsDB{
		Trie:                   merkleTrie,
		Hasher:                 scf.core.Hasher(),
		Marshaller:             scf.core.InternalMarshalizer(),
		AccountFactory:         accountFactory,
		StoragePruningManager:  storagePruning,
		AddressConverter:       scf.core.AddressPubKeyConverter(),
		SnapshotsManager:       snapshotManager,
		StateAccessesCollector: disabled.NewDisabledStateAccessesCollector(),
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

func parseStateChangesTypesToCollect(stateChangesTypes []string) (collectRead bool, collectWrite bool, err error) {
	types := sanitizeActionTypes(data.ActionType_value)
	for _, stateChangeType := range stateChangesTypes {
		if value, ok := types[strings.ToLower(stateChangeType)]; ok {
			switch value {
			case 0:
				collectRead = true

			case 1:
				collectWrite = true
			}
		} else {
			return false, false, fmt.Errorf("unknown action type %s", stateChangeType)
		}
	}

	return collectRead, collectWrite, nil
}

func sanitizeActionTypes(actionTypes map[string]int32) map[string]int32 {
	sanitizedActionTypes := make(map[string]int32, len(actionTypes))

	for actionType, value := range actionTypes {
		sanitizedActionTypes[strings.ToLower(actionType)] = value
	}

	return sanitizedActionTypes
}
