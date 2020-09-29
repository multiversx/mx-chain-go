package factory

import (
	"fmt"
	"math"
	"os"
	"path"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
	"github.com/ElrondNetwork/elrond-go/update/storing"
	"github.com/ElrondNetwork/elrond-go/update/sync"
)

// ArgsExporter is the argument structure to create a new exporter
type ArgsExporter struct {
	CoreComponents           process.CoreComponentsHolder
	CryptoComponents         process.CryptoComponentsHolder
	HeaderValidator          epochStart.HeaderValidator
	DataPool                 dataRetriever.PoolsHolder
	StorageService           dataRetriever.StorageService
	RequestHandler           process.RequestHandler
	ShardCoordinator         sharding.Coordinator
	Messenger                p2p.Messenger
	ActiveAccountsDBs        map[state.AccountsDbIdentifier]state.AccountsAdapter
	ExistingResolvers        dataRetriever.ResolversContainer
	ExportFolder             string
	ExportTriesStorageConfig config.StorageConfig
	ExportStateStorageConfig config.StorageConfig
	ExportStateKeysConfig    config.StorageConfig
	MaxTrieLevelInMemory     uint
	WhiteListHandler         process.WhiteListHandler
	WhiteListerVerifiedTxs   process.WhiteListHandler
	InterceptorsContainer    process.InterceptorsContainer
	NodesCoordinator         sharding.NodesCoordinator
	HeaderSigVerifier        process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier  process.HeaderIntegrityVerifier
	ValidityAttester         process.ValidityAttester
	InputAntifloodHandler    process.P2PAntifloodHandler
	OutputAntifloodHandler   process.P2PAntifloodHandler
	Rounder                  process.Rounder
}

type exportHandlerFactory struct {
	CoreComponents           process.CoreComponentsHolder
	CryptoComponents         process.CryptoComponentsHolder
	headerValidator          epochStart.HeaderValidator
	dataPool                 dataRetriever.PoolsHolder
	storageService           dataRetriever.StorageService
	requestHandler           process.RequestHandler
	shardCoordinator         sharding.Coordinator
	messenger                p2p.Messenger
	activeAccountsDBs        map[state.AccountsDbIdentifier]state.AccountsAdapter
	exportFolder             string
	exportTriesStorageConfig config.StorageConfig
	exportStateStorageConfig config.StorageConfig
	exportStateKeysConfig    config.StorageConfig
	maxTrieLevelInMemory     uint
	whiteListHandler         process.WhiteListHandler
	whiteListerVerifiedTxs   process.WhiteListHandler
	interceptorsContainer    process.InterceptorsContainer
	existingResolvers        dataRetriever.ResolversContainer
	epochStartTrigger        epochStart.TriggerHandler
	accounts                 state.AccountsAdapter
	nodesCoordinator         sharding.NodesCoordinator
	headerSigVerifier        process.InterceptedHeaderSigVerifier
	headerIntegrityVerifier  process.HeaderIntegrityVerifier
	validityAttester         process.ValidityAttester
	resolverContainer        dataRetriever.ResolversContainer
	inputAntifloodHandler    process.P2PAntifloodHandler
	outputAntifloodHandler   process.P2PAntifloodHandler
	rounder                  process.Rounder
}

// NewExportHandlerFactory creates an exporter factory
func NewExportHandlerFactory(args ArgsExporter) (*exportHandlerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.CoreComponents) {
		return nil, update.ErrNilCoreComponents
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, update.ErrNilCryptoComponents
	}
	if check.IfNil(args.CoreComponents.Hasher()) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.CoreComponents.InternalMarshalizer()) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.HeaderValidator) {
		return nil, update.ErrNilHeaderValidator
	}
	if check.IfNil(args.CoreComponents.Uint64ByteSliceConverter()) {
		return nil, update.ErrNilUint64Converter
	}
	if check.IfNil(args.DataPool) {
		return nil, update.ErrNilDataPoolHolder
	}
	if check.IfNil(args.StorageService) {
		return nil, update.ErrNilStorage
	}
	if check.IfNil(args.RequestHandler) {
		return nil, update.ErrNilRequestHandler
	}
	if check.IfNil(args.Messenger) {
		return nil, update.ErrNilMessenger
	}
	if args.ActiveAccountsDBs == nil {
		return nil, update.ErrNilAccounts
	}
	if check.IfNil(args.WhiteListHandler) {
		return nil, update.ErrNilWhiteListHandler
	}
	if check.IfNil(args.WhiteListerVerifiedTxs) {
		return nil, update.ErrNilWhiteListHandler
	}
	if check.IfNil(args.InterceptorsContainer) {
		return nil, update.ErrNilInterceptorsContainer
	}
	if check.IfNil(args.ExistingResolvers) {
		return nil, update.ErrNilResolverContainer
	}
	if check.IfNil(args.CryptoComponents.MultiSigner()) {
		return nil, update.ErrNilMultiSigner
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, update.ErrNilNodesCoordinator
	}
	if check.IfNil(args.CryptoComponents.TxSingleSigner()) {
		return nil, update.ErrNilSingleSigner
	}
	if check.IfNil(args.CoreComponents.AddressPubKeyConverter()) {
		return nil, update.ErrNilPubKeyConverter
	}
	if check.IfNil(args.CoreComponents.ValidatorPubKeyConverter()) {
		return nil, update.ErrNilPubKeyConverter
	}
	if check.IfNil(args.CryptoComponents.BlockSignKeyGen()) {
		return nil, update.ErrNilBlockKeyGen
	}
	if check.IfNil(args.CryptoComponents.TxSignKeyGen()) {
		return nil, update.ErrNilKeyGenerator
	}
	if check.IfNil(args.CryptoComponents.BlockSigner()) {
		return nil, update.ErrNilBlockSigner
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return nil, update.ErrNilHeaderSigVerifier
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return nil, update.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(args.ValidityAttester) {
		return nil, update.ErrNilValidityAttester
	}
	if check.IfNil(args.CoreComponents.TxMarshalizer()) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.InputAntifloodHandler) {
		return nil, update.ErrNilAntiFloodHandler
	}
	if check.IfNil(args.OutputAntifloodHandler) {
		return nil, update.ErrNilAntiFloodHandler
	}
	if check.IfNil(args.Rounder) {
		return nil, update.ErrNilRounder
	}

	e := &exportHandlerFactory{
		CoreComponents:           args.CoreComponents,
		CryptoComponents:         args.CryptoComponents,
		headerValidator:          args.HeaderValidator,
		dataPool:                 args.DataPool,
		storageService:           args.StorageService,
		requestHandler:           args.RequestHandler,
		shardCoordinator:         args.ShardCoordinator,
		messenger:                args.Messenger,
		activeAccountsDBs:        args.ActiveAccountsDBs,
		exportFolder:             args.ExportFolder,
		exportTriesStorageConfig: args.ExportTriesStorageConfig,
		exportStateStorageConfig: args.ExportStateStorageConfig,
		exportStateKeysConfig:    args.ExportStateKeysConfig,
		interceptorsContainer:    args.InterceptorsContainer,
		whiteListHandler:         args.WhiteListHandler,
		whiteListerVerifiedTxs:   args.WhiteListerVerifiedTxs,
		existingResolvers:        args.ExistingResolvers,
		accounts:                 args.ActiveAccountsDBs[state.UserAccountsState],
		nodesCoordinator:         args.NodesCoordinator,
		headerSigVerifier:        args.HeaderSigVerifier,
		headerIntegrityVerifier:  args.HeaderIntegrityVerifier,
		validityAttester:         args.ValidityAttester,
		inputAntifloodHandler:    args.InputAntifloodHandler,
		outputAntifloodHandler:   args.OutputAntifloodHandler,
		maxTrieLevelInMemory:     args.MaxTrieLevelInMemory,
		rounder:                  args.Rounder,
	}

	return e, nil
}

// Create makes a new export handler
func (e *exportHandlerFactory) Create() (update.ExportHandler, error) {
	err := e.prepareFolders(e.exportFolder)
	if err != nil {
		return nil, err
	}

	argsPeerMiniBlocksSyncer := shardchain.ArgPeerMiniBlockSyncer{
		MiniBlocksPool: e.dataPool.MiniBlocks(),
		Requesthandler: e.requestHandler,
	}
	peerMiniBlocksSyncer, err := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlocksSyncer)
	if err != nil {
		return nil, err
	}
	argsEpochTrigger := shardchain.ArgsShardEpochStartTrigger{
		Marshalizer:          e.CoreComponents.InternalMarshalizer(),
		Hasher:               e.CoreComponents.Hasher(),
		HeaderValidator:      e.headerValidator,
		Uint64Converter:      e.CoreComponents.Uint64ByteSliceConverter(),
		DataPool:             e.dataPool,
		Storage:              e.storageService,
		RequestHandler:       e.requestHandler,
		EpochStartNotifier:   notifier.NewEpochStartSubscriptionHandler(),
		Epoch:                0,
		Validity:             process.MetaBlockValidity,
		Finality:             process.BlockFinality,
		PeerMiniBlocksSyncer: peerMiniBlocksSyncer,
		Rounder:              e.rounder,
		AppStatusHandler:     e.CoreComponents.StatusHandler(),
	}
	epochHandler, err := shardchain.NewEpochStartTrigger(&argsEpochTrigger)
	if err != nil {
		return nil, err
	}

	argsDataTrieFactory := ArgsNewDataTrieFactory{
		StorageConfig:        e.exportTriesStorageConfig,
		SyncFolder:           e.exportFolder,
		Marshalizer:          e.CoreComponents.InternalMarshalizer(),
		Hasher:               e.CoreComponents.Hasher(),
		ShardCoordinator:     e.shardCoordinator,
		MaxTrieLevelInMemory: e.maxTrieLevelInMemory,
	}
	dataTriesContainerFactory, err := NewDataTrieFactory(argsDataTrieFactory)
	if err != nil {
		return nil, err
	}
	dataTries, err := dataTriesContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	argsResolvers := ArgsNewResolversContainerFactory{
		ShardCoordinator:           e.shardCoordinator,
		Messenger:                  e.messenger,
		Marshalizer:                e.CoreComponents.InternalMarshalizer(),
		DataTrieContainer:          dataTries,
		ExistingResolvers:          e.existingResolvers,
		NumConcurrentResolvingJobs: 100,
		InputAntifloodHandler:      e.inputAntifloodHandler,
		OutputAntifloodHandler:     e.outputAntifloodHandler,
	}
	resolversFactory, err := NewResolversContainerFactory(argsResolvers)
	if err != nil {
		return nil, err
	}
	e.resolverContainer, err = resolversFactory.Create()
	if err != nil {
		return nil, err
	}

	argsAccountsSyncers := ArgsNewAccountsDBSyncersContainerFactory{
		TrieCacher:           e.dataPool.TrieNodes(),
		RequestHandler:       e.requestHandler,
		ShardCoordinator:     e.shardCoordinator,
		Hasher:               e.CoreComponents.Hasher(),
		Marshalizer:          e.CoreComponents.InternalMarshalizer(),
		TrieStorageManager:   dataTriesContainerFactory.TrieStorageManager(),
		WaitTime:             time.Minute,
		MaxTrieLevelInMemory: e.maxTrieLevelInMemory,
	}
	accountsDBSyncerFactory, err := NewAccountsDBSContainerFactory(argsAccountsSyncers)
	if err != nil {
		return nil, err
	}
	accountsDBSyncerContainer, err := accountsDBSyncerFactory.Create()
	if err != nil {
		return nil, err
	}

	argsNewHeadersSync := sync.ArgsNewHeadersSyncHandler{
		StorageService:   e.storageService,
		Cache:            e.dataPool.Headers(),
		Marshalizer:      e.CoreComponents.InternalMarshalizer(),
		Hasher:           e.CoreComponents.Hasher(),
		EpochHandler:     epochHandler,
		RequestHandler:   e.requestHandler,
		Uint64Converter:  e.CoreComponents.Uint64ByteSliceConverter(),
		ShardCoordinator: e.shardCoordinator,
	}
	epochStartHeadersSyncer, err := sync.NewHeadersSyncHandler(argsNewHeadersSync)
	if err != nil {
		return nil, err
	}

	argsNewSyncAccountsDBsHandler := sync.ArgsNewSyncAccountsDBsHandler{
		AccountsDBsSyncers: accountsDBSyncerContainer,
		ActiveAccountsDBs:  e.activeAccountsDBs,
	}
	epochStartTrieSyncer, err := sync.NewSyncAccountsDBsHandler(argsNewSyncAccountsDBsHandler)
	if err != nil {
		return nil, err
	}

	argsMiniBlockSyncer := sync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        e.storageService.GetStorer(dataRetriever.MiniBlockUnit),
		Cache:          e.dataPool.MiniBlocks(),
		Marshalizer:    e.CoreComponents.InternalMarshalizer(),
		RequestHandler: e.requestHandler,
	}
	epochStartMiniBlocksSyncer, err := sync.NewPendingMiniBlocksSyncer(argsMiniBlockSyncer)
	if err != nil {
		return nil, err
	}

	argsPendingTransactions := sync.ArgsNewPendingTransactionsSyncer{
		DataPools:      e.dataPool,
		Storages:       e.storageService,
		Marshalizer:    e.CoreComponents.InternalMarshalizer(),
		RequestHandler: e.requestHandler,
	}
	epochStartTransactionsSyncer, err := sync.NewPendingTransactionsSyncer(argsPendingTransactions)
	if err != nil {
		return nil, err
	}

	argsSyncState := sync.ArgsNewSyncState{
		Headers:      epochStartHeadersSyncer,
		Tries:        epochStartTrieSyncer,
		MiniBlocks:   epochStartMiniBlocksSyncer,
		Transactions: epochStartTransactionsSyncer,
	}
	stateSyncer, err := sync.NewSyncState(argsSyncState)
	if err != nil {
		return nil, err
	}

	keysStorer, err := createStorer(e.exportStateKeysConfig, e.exportFolder)
	if err != nil {
		return nil, fmt.Errorf("%w while creating keys storer", err)
	}
	keysVals, err := createStorer(e.exportStateStorageConfig, e.exportFolder)
	if err != nil {
		return nil, fmt.Errorf("%w while creating keys-values storer", err)
	}

	arg := storing.ArgHardforkStorer{
		KeysStore:   keysStorer,
		KeyValue:    keysVals,
		Marshalizer: e.CoreComponents.InternalMarshalizer(),
	}
	hs, err := storing.NewHardforkStorer(arg)

	argsExporter := genesis.ArgsNewStateExporter{
		ShardCoordinator:         e.shardCoordinator,
		StateSyncer:              stateSyncer,
		Marshalizer:              e.CoreComponents.InternalMarshalizer(),
		HardforkStorer:           hs,
		Hasher:                   e.CoreComponents.Hasher(),
		ExportFolder:             e.exportFolder,
		ValidatorPubKeyConverter: e.CoreComponents.ValidatorPubKeyConverter(),
		AddressPubKeyConverter:   e.CoreComponents.AddressPubKeyConverter(),
		GenesisNodesSetupHandler: e.CoreComponents.GenesisNodesSetup(),
	}
	exportHandler, err := genesis.NewStateExporter(argsExporter)
	if err != nil {
		return nil, err
	}

	e.epochStartTrigger = epochHandler
	err = e.createInterceptors()
	if err != nil {
		return nil, err
	}

	return exportHandler, nil
}

func (e *exportHandlerFactory) prepareFolders(folder string) error {
	err := os.RemoveAll(folder)
	if err != nil {
		return err
	}

	return os.MkdirAll(folder, os.ModePerm)
}

func (e *exportHandlerFactory) createInterceptors() error {
	argsInterceptors := ArgsNewFullSyncInterceptorsContainerFactory{
		CoreComponents:          e.CoreComponents,
		CryptoComponents:        e.CryptoComponents,
		Accounts:                e.accounts,
		ShardCoordinator:        e.shardCoordinator,
		NodesCoordinator:        e.nodesCoordinator,
		Messenger:               e.messenger,
		Store:                   e.storageService,
		DataPool:                e.dataPool,
		MaxTxNonceDeltaAllowed:  math.MaxInt32,
		TxFeeHandler:            &disabled.FeeHandler{},
		BlockBlackList:          timecache.NewTimeCache(time.Second),
		HeaderSigVerifier:       e.headerSigVerifier,
		HeaderIntegrityVerifier: e.headerIntegrityVerifier,
		SizeCheckDelta:          math.MaxUint32,
		ValidityAttester:        e.validityAttester,
		EpochStartTrigger:       e.epochStartTrigger,
		WhiteListHandler:        e.whiteListHandler,
		WhiteListerVerifiedTxs:  e.whiteListerVerifiedTxs,
		InterceptorsContainer:   e.interceptorsContainer,
		AntifloodHandler:        e.inputAntifloodHandler,
	}
	fullSyncInterceptors, err := NewFullSyncInterceptorsContainerFactory(argsInterceptors)
	if err != nil {
		return err
	}

	interceptorsContainer, err := fullSyncInterceptors.Create()
	if err != nil {
		return err
	}

	e.interceptorsContainer = interceptorsContainer
	return nil
}

func createStorer(storageConfig config.StorageConfig, folder string) (storage.Storer, error) {
	dbConfig := storageFactory.GetDBFromConfig(storageConfig.DB)
	dbConfig.FilePath = path.Join(folder, storageConfig.DB.FilePath)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(storageConfig.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(storageConfig.Bloom),
	)
	if err != nil {
		return nil, err
	}

	return accountsTrieStorage, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *exportHandlerFactory) IsInterfaceNil() bool {
	return e == nil
}
