package components

import (
	"fmt"
	"math/big"
	"path/filepath"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/common/ordering"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	dbLookupFactory "github.com/multiversx/mx-chain-go/dblookupext/factory"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/factory"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors/disabled"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/trigger"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// ArgsProcessComponentsHolder will hold the components needed for process components
type ArgsProcessComponentsHolder struct {
	CoreComponents       factory.CoreComponentsHolder
	CryptoComponents     factory.CryptoComponentsHolder
	NetworkComponents    factory.NetworkComponentsHolder
	BootstrapComponents  factory.BootstrapComponentsHolder
	StateComponents      factory.StateComponentsHolder
	DataComponents       factory.DataComponentsHolder
	StatusComponents     factory.StatusComponentsHolder
	StatusCoreComponents factory.StatusCoreComponentsHolder
	NodesCoordinator     nodesCoordinator.NodesCoordinator

	EpochConfig              config.EpochConfig
	RoundConfig              config.RoundConfig
	ConfigurationPathsHolder config.ConfigurationPathsHolder
	FlagsConfig              config.ContextFlagsConfig
	ImportDBConfig           config.ImportDbConfig
	PrefsConfig              config.Preferences
	Config                   config.Config
	EconomicsConfig          config.EconomicsConfig
	SystemSCConfig           config.SystemSmartContractsConfig
}

type processComponentsHolder struct {
	closeHandler                     *closeHandler
	receiptsRepository               factory.ReceiptsRepository
	nodesCoordinator                 nodesCoordinator.NodesCoordinator
	shardCoordinator                 sharding.Coordinator
	interceptorsContainer            process.InterceptorsContainer
	fullArchiveInterceptorsContainer process.InterceptorsContainer
	resolversContainer               dataRetriever.ResolversContainer
	requestersFinder                 dataRetriever.RequestersFinder
	roundHandler                     consensus.RoundHandler
	epochStartTrigger                epochStart.TriggerHandler
	epochStartNotifier               factory.EpochStartNotifier
	forkDetector                     process.ForkDetector
	blockProcessor                   process.BlockProcessor
	blackListHandler                 process.TimeCacher
	bootStorer                       process.BootStorer
	headerSigVerifier                process.InterceptedHeaderSigVerifier
	headerIntegrityVerifier          process.HeaderIntegrityVerifier
	validatorsStatistics             process.ValidatorStatisticsProcessor
	validatorsProvider               process.ValidatorsProvider
	blockTracker                     process.BlockTracker
	pendingMiniBlocksHandler         process.PendingMiniBlocksHandler
	requestHandler                   process.RequestHandler
	txLogsProcessor                  process.TransactionLogProcessorDatabase
	headerConstructionValidator      process.HeaderConstructionValidator
	peerShardMapper                  process.NetworkShardingCollector
	fullArchivePeerShardMapper       process.NetworkShardingCollector
	fallbackHeaderValidator          process.FallbackHeaderValidator
	apiTransactionEvaluator          factory.TransactionEvaluator
	whiteListHandler                 process.WhiteListHandler
	whiteListerVerifiedTxs           process.WhiteListHandler
	historyRepository                dblookupext.HistoryRepository
	importStartHandler               update.ImportStartHandler
	requestedItemsHandler            dataRetriever.RequestedItemsHandler
	nodeRedundancyHandler            consensus.NodeRedundancyHandler
	currentEpochProvider             process.CurrentNetworkEpochProviderHandler
	scheduledTxsExecutionHandler     process.ScheduledTxsExecutionHandler
	txsSenderHandler                 process.TxsSenderHandler
	hardforkTrigger                  factory.HardforkTrigger
	processedMiniBlocksTracker       process.ProcessedMiniBlocksTracker
	esdtDataStorageHandlerForAPI     vmcommon.ESDTNFTStorageHandler
	accountsParser                   genesis.AccountsParser
	sendSignatureTracker             process.SentSignaturesTracker
}

// CreateProcessComponents will create the process components holder
func CreateProcessComponents(args ArgsProcessComponentsHolder) (factory.ProcessComponentsHandler, error) {
	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join(args.FlagsConfig.DbDir, common.DefaultDBPath), args.FlagsConfig.Version)
	if err != nil {
		return nil, err
	}
	totalSupply, ok := big.NewInt(0).SetString(args.EconomicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	if !ok {
		return nil, fmt.Errorf("can not parse total suply from economics.toml, %s is not a valid value",
			args.EconomicsConfig.GlobalSettings.GenesisTotalSupply)
	}

	mintingSenderAddress := args.EconomicsConfig.GlobalSettings.GenesisMintingSenderAddress
	argsAccountsParser := genesis.AccountsParserArgs{
		GenesisFilePath: args.ConfigurationPathsHolder.Genesis,
		EntireSupply:    totalSupply,
		MinterAddress:   mintingSenderAddress,
		PubkeyConverter: args.CoreComponents.AddressPubKeyConverter(),
		KeyGenerator:    args.CryptoComponents.TxSignKeyGen(),
		Hasher:          args.CoreComponents.Hasher(),
		Marshalizer:     args.CoreComponents.InternalMarshalizer(),
	}

	accountsParser, err := parsing.NewAccountsParser(argsAccountsParser)
	if err != nil {
		return nil, err
	}

	smartContractParser, err := parsing.NewSmartContractsParser(
		args.ConfigurationPathsHolder.SmartContracts,
		args.CoreComponents.AddressPubKeyConverter(),
		args.CryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	historyRepoFactoryArgs := &dbLookupFactory.ArgsHistoryRepositoryFactory{
		SelfShardID:              args.BootstrapComponents.ShardCoordinator().SelfId(),
		Config:                   args.Config.DbLookupExtensions,
		Hasher:                   args.CoreComponents.Hasher(),
		Marshalizer:              args.CoreComponents.InternalMarshalizer(),
		Store:                    args.DataComponents.StorageService(),
		Uint64ByteSliceConverter: args.CoreComponents.Uint64ByteSliceConverter(),
	}
	historyRepositoryFactory, err := dbLookupFactory.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
	if err != nil {
		return nil, err
	}

	whiteListRequest, err := disabled.NewDisabledWhiteListDataVerifier()
	if err != nil {
		return nil, err
	}

	whiteListerVerifiedTxs, err := disabled.NewDisabledWhiteListDataVerifier()
	if err != nil {
		return nil, err
	}

	historyRepository, err := historyRepositoryFactory.Create()
	if err != nil {
		return nil, err
	}

	requestedItemsHandler := cache.NewTimeCache(
		time.Duration(uint64(time.Millisecond) * args.CoreComponents.GenesisNodesSetup().GetRoundDuration()))

	txExecutionOrderHandler := ordering.NewOrderedCollection()

	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig:  args.EpochConfig.GasSchedule,
		ConfigDir:          args.ConfigurationPathsHolder.GasScheduleDirectoryName,
		EpochNotifier:      args.CoreComponents.EpochNotifier(),
		WasmVMChangeLocker: args.CoreComponents.WasmVMChangeLocker(),
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
	if err != nil {
		return nil, err
	}

	processArgs := processComp.ProcessComponentsFactoryArgs{
		Config:                  args.Config,
		EpochConfig:             args.EpochConfig,
		RoundConfig:             args.RoundConfig,
		PrefConfigs:             args.PrefsConfig,
		ImportDBConfig:          args.ImportDBConfig,
		AccountsParser:          accountsParser,
		SmartContractParser:     smartContractParser,
		GasSchedule:             gasScheduleNotifier,
		NodesCoordinator:        args.NodesCoordinator,
		Data:                    args.DataComponents,
		CoreData:                args.CoreComponents,
		Crypto:                  args.CryptoComponents,
		State:                   args.StateComponents,
		Network:                 args.NetworkComponents,
		BootstrapComponents:     args.BootstrapComponents,
		StatusComponents:        args.StatusComponents,
		StatusCoreComponents:    args.StatusCoreComponents,
		RequestedItemsHandler:   requestedItemsHandler,
		WhiteListHandler:        whiteListRequest,
		WhiteListerVerifiedTxs:  whiteListerVerifiedTxs,
		MaxRating:               50,
		SystemSCConfig:          &args.SystemSCConfig,
		ImportStartHandler:      importStartHandler,
		HistoryRepo:             historyRepository,
		FlagsConfig:             args.FlagsConfig,
		TxExecutionOrderHandler: txExecutionOrderHandler,
	}
	processComponentsFactory, err := processComp.NewProcessComponentsFactory(processArgs)
	if err != nil {
		return nil, fmt.Errorf("NewProcessComponentsFactory failed: %w", err)
	}

	managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedProcessComponents.Create()
	if err != nil {
		return nil, err
	}

	instance := &processComponentsHolder{
		closeHandler:                     NewCloseHandler(),
		receiptsRepository:               managedProcessComponents.ReceiptsRepository(),
		nodesCoordinator:                 managedProcessComponents.NodesCoordinator(),
		shardCoordinator:                 managedProcessComponents.ShardCoordinator(),
		interceptorsContainer:            managedProcessComponents.InterceptorsContainer(),
		fullArchiveInterceptorsContainer: managedProcessComponents.FullArchiveInterceptorsContainer(),
		resolversContainer:               managedProcessComponents.ResolversContainer(),
		requestersFinder:                 managedProcessComponents.RequestersFinder(),
		roundHandler:                     managedProcessComponents.RoundHandler(),
		epochStartTrigger:                managedProcessComponents.EpochStartTrigger(),
		epochStartNotifier:               managedProcessComponents.EpochStartNotifier(),
		forkDetector:                     managedProcessComponents.ForkDetector(),
		blockProcessor:                   managedProcessComponents.BlockProcessor(),
		blackListHandler:                 managedProcessComponents.BlackListHandler(),
		bootStorer:                       managedProcessComponents.BootStorer(),
		headerSigVerifier:                managedProcessComponents.HeaderSigVerifier(),
		headerIntegrityVerifier:          managedProcessComponents.HeaderIntegrityVerifier(),
		validatorsStatistics:             managedProcessComponents.ValidatorsStatistics(),
		validatorsProvider:               managedProcessComponents.ValidatorsProvider(),
		blockTracker:                     managedProcessComponents.BlockTracker(),
		pendingMiniBlocksHandler:         managedProcessComponents.PendingMiniBlocksHandler(),
		requestHandler:                   managedProcessComponents.RequestHandler(),
		txLogsProcessor:                  managedProcessComponents.TxLogsProcessor(),
		headerConstructionValidator:      managedProcessComponents.HeaderConstructionValidator(),
		peerShardMapper:                  managedProcessComponents.PeerShardMapper(),
		fullArchivePeerShardMapper:       managedProcessComponents.FullArchivePeerShardMapper(),
		fallbackHeaderValidator:          managedProcessComponents.FallbackHeaderValidator(),
		apiTransactionEvaluator:          managedProcessComponents.APITransactionEvaluator(),
		whiteListHandler:                 managedProcessComponents.WhiteListHandler(),
		whiteListerVerifiedTxs:           managedProcessComponents.WhiteListerVerifiedTxs(),
		historyRepository:                managedProcessComponents.HistoryRepository(),
		importStartHandler:               managedProcessComponents.ImportStartHandler(),
		requestedItemsHandler:            managedProcessComponents.RequestedItemsHandler(),
		nodeRedundancyHandler:            managedProcessComponents.NodeRedundancyHandler(),
		currentEpochProvider:             managedProcessComponents.CurrentEpochProvider(),
		scheduledTxsExecutionHandler:     managedProcessComponents.ScheduledTxsExecutionHandler(),
		txsSenderHandler:                 managedProcessComponents.TxsSenderHandler(),
		hardforkTrigger:                  managedProcessComponents.HardforkTrigger(),
		processedMiniBlocksTracker:       managedProcessComponents.ProcessedMiniBlocksTracker(),
		esdtDataStorageHandlerForAPI:     managedProcessComponents.ESDTDataStorageHandlerForAPI(),
		accountsParser:                   managedProcessComponents.AccountsParser(),
		sendSignatureTracker:             managedProcessComponents.SentSignaturesTracker(),
	}

	instance.collectClosableComponents()

	return instance, nil
}

// SentSignaturesTracker will return the send signature tracker
func (p *processComponentsHolder) SentSignaturesTracker() process.SentSignaturesTracker {
	return p.sendSignatureTracker
}

// NodesCoordinator will return the nodes coordinator
func (p *processComponentsHolder) NodesCoordinator() nodesCoordinator.NodesCoordinator {
	return p.nodesCoordinator
}

// ShardCoordinator will return the shard coordinator
func (p *processComponentsHolder) ShardCoordinator() sharding.Coordinator {
	return p.shardCoordinator
}

// InterceptorsContainer will return the interceptors container
func (p *processComponentsHolder) InterceptorsContainer() process.InterceptorsContainer {
	return p.interceptorsContainer
}

// FullArchiveInterceptorsContainer will return the full archive interceptor container
func (p *processComponentsHolder) FullArchiveInterceptorsContainer() process.InterceptorsContainer {
	return p.fullArchiveInterceptorsContainer
}

// ResolversContainer will return the resolvers container
func (p *processComponentsHolder) ResolversContainer() dataRetriever.ResolversContainer {
	return p.resolversContainer
}

// RequestersFinder will return the requesters finder
func (p *processComponentsHolder) RequestersFinder() dataRetriever.RequestersFinder {
	return p.requestersFinder
}

// RoundHandler will return the round handler
func (p *processComponentsHolder) RoundHandler() consensus.RoundHandler {
	return p.roundHandler
}

// EpochStartTrigger will return the epoch start trigger
func (p *processComponentsHolder) EpochStartTrigger() epochStart.TriggerHandler {
	return p.epochStartTrigger
}

// EpochStartNotifier will return the epoch start notifier
func (p *processComponentsHolder) EpochStartNotifier() factory.EpochStartNotifier {
	return p.epochStartNotifier
}

// ForkDetector will return the fork detector
func (p *processComponentsHolder) ForkDetector() process.ForkDetector {
	return p.forkDetector
}

// BlockProcessor will return the block processor
func (p *processComponentsHolder) BlockProcessor() process.BlockProcessor {
	return p.blockProcessor
}

// BlackListHandler will return the black list handler
func (p *processComponentsHolder) BlackListHandler() process.TimeCacher {
	return p.blackListHandler
}

// BootStorer will return the boot storer
func (p *processComponentsHolder) BootStorer() process.BootStorer {
	return p.bootStorer
}

// HeaderSigVerifier will return the header sign verifier
func (p *processComponentsHolder) HeaderSigVerifier() process.InterceptedHeaderSigVerifier {
	return p.headerSigVerifier
}

// HeaderIntegrityVerifier will return the header integrity verifier
func (p *processComponentsHolder) HeaderIntegrityVerifier() process.HeaderIntegrityVerifier {
	return p.headerIntegrityVerifier
}

// ValidatorsStatistics will return the validators statistics
func (p *processComponentsHolder) ValidatorsStatistics() process.ValidatorStatisticsProcessor {
	return p.validatorsStatistics
}

// ValidatorsProvider will return the validators provider
func (p *processComponentsHolder) ValidatorsProvider() process.ValidatorsProvider {
	return p.validatorsProvider
}

// BlockTracker will return the block tracker
func (p *processComponentsHolder) BlockTracker() process.BlockTracker {
	return p.blockTracker
}

// PendingMiniBlocksHandler will return the pending miniblocks handler
func (p *processComponentsHolder) PendingMiniBlocksHandler() process.PendingMiniBlocksHandler {
	return p.pendingMiniBlocksHandler
}

// RequestHandler will return the request handler
func (p *processComponentsHolder) RequestHandler() process.RequestHandler {
	return p.requestHandler
}

// TxLogsProcessor will return the transaction log processor
func (p *processComponentsHolder) TxLogsProcessor() process.TransactionLogProcessorDatabase {
	return p.txLogsProcessor
}

// HeaderConstructionValidator will return the header construction validator
func (p *processComponentsHolder) HeaderConstructionValidator() process.HeaderConstructionValidator {
	return p.headerConstructionValidator
}

// PeerShardMapper will return the peer shard mapper
func (p *processComponentsHolder) PeerShardMapper() process.NetworkShardingCollector {
	return p.peerShardMapper
}

// FullArchivePeerShardMapper will return the full archive peer shard mapper
func (p *processComponentsHolder) FullArchivePeerShardMapper() process.NetworkShardingCollector {
	return p.fullArchivePeerShardMapper
}

// FallbackHeaderValidator will return the fallback header validator
func (p *processComponentsHolder) FallbackHeaderValidator() process.FallbackHeaderValidator {
	return p.fallbackHeaderValidator
}

// APITransactionEvaluator will return the api transaction evaluator
func (p *processComponentsHolder) APITransactionEvaluator() factory.TransactionEvaluator {
	return p.apiTransactionEvaluator
}

// WhiteListHandler will return the white list handler
func (p *processComponentsHolder) WhiteListHandler() process.WhiteListHandler {
	return p.whiteListHandler
}

// WhiteListerVerifiedTxs will return the white lister verifier
func (p *processComponentsHolder) WhiteListerVerifiedTxs() process.WhiteListHandler {
	return p.whiteListerVerifiedTxs
}

// HistoryRepository will return the history repository
func (p *processComponentsHolder) HistoryRepository() dblookupext.HistoryRepository {
	return p.historyRepository
}

// ImportStartHandler will return the import start handler
func (p *processComponentsHolder) ImportStartHandler() update.ImportStartHandler {
	return p.importStartHandler
}

// RequestedItemsHandler will return the requested item handler
func (p *processComponentsHolder) RequestedItemsHandler() dataRetriever.RequestedItemsHandler {
	return p.requestedItemsHandler
}

// NodeRedundancyHandler will return the node redundancy handler
func (p *processComponentsHolder) NodeRedundancyHandler() consensus.NodeRedundancyHandler {
	return p.nodeRedundancyHandler
}

// CurrentEpochProvider will return the current epoch provider
func (p *processComponentsHolder) CurrentEpochProvider() process.CurrentNetworkEpochProviderHandler {
	return p.currentEpochProvider
}

// ScheduledTxsExecutionHandler will return the scheduled transactions execution handler
func (p *processComponentsHolder) ScheduledTxsExecutionHandler() process.ScheduledTxsExecutionHandler {
	return p.scheduledTxsExecutionHandler
}

// TxsSenderHandler will return the transactions sender handler
func (p *processComponentsHolder) TxsSenderHandler() process.TxsSenderHandler {
	return p.txsSenderHandler
}

// HardforkTrigger will return the hardfork trigger
func (p *processComponentsHolder) HardforkTrigger() factory.HardforkTrigger {
	return p.hardforkTrigger
}

// ProcessedMiniBlocksTracker will return the processed miniblocks tracker
func (p *processComponentsHolder) ProcessedMiniBlocksTracker() process.ProcessedMiniBlocksTracker {
	return p.processedMiniBlocksTracker
}

// ESDTDataStorageHandlerForAPI will return the esdt data storage handler for api
func (p *processComponentsHolder) ESDTDataStorageHandlerForAPI() vmcommon.ESDTNFTStorageHandler {
	return p.esdtDataStorageHandlerForAPI
}

// AccountsParser will return the accounts parser
func (p *processComponentsHolder) AccountsParser() genesis.AccountsParser {
	return p.accountsParser
}

// ReceiptsRepository returns the receipts repository
func (p *processComponentsHolder) ReceiptsRepository() factory.ReceiptsRepository {
	return p.receiptsRepository
}

func (p *processComponentsHolder) collectClosableComponents() {
	p.closeHandler.AddComponent(p.interceptorsContainer)
	p.closeHandler.AddComponent(p.fullArchiveInterceptorsContainer)
	p.closeHandler.AddComponent(p.resolversContainer)
	p.closeHandler.AddComponent(p.epochStartTrigger)
	p.closeHandler.AddComponent(p.blockProcessor)
	p.closeHandler.AddComponent(p.validatorsProvider)
	p.closeHandler.AddComponent(p.txsSenderHandler)
}

// Close will call the Close methods on all inner components
func (p *processComponentsHolder) Close() error {
	return p.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *processComponentsHolder) IsInterfaceNil() bool {
	return p == nil
}

// Create will do nothing
func (p *processComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (p *processComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (p *processComponentsHolder) String() string {
	return ""
}
