package factory

import (
	"context"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/genesis"
	processComp "github.com/multiversx/mx-chain-go/genesis/process"
	heartbeatData "github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NotifyEpochChangeConfirmed(epoch uint32)
	IsInterfaceNil() bool
}

// EpochStartNotifierWithConfirm defines which actions should be done for handling new epoch's events and confirmation
type EpochStartNotifierWithConfirm interface {
	EpochStartNotifier
	RegisterForEpochChangeConfirmed(handler func(epoch uint32))
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	SetDebugger(debugger process.AntifloodDebugger) error
	SetPeerValidatorMapper(validatorMapper process.PeerValidatorMapper) error
	SetTopicsForAll(topics ...string)
	ApplyConsensusSize(size int)
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsOriginatorEligibleForTopic(pid core.PeerID, topic string) error
	Close() error
	IsInterfaceNil() bool
}

// PreferredPeersHolderHandler defines the behavior of a component able to handle preferred peers operations
type PreferredPeersHolderHandler interface {
	PutConnectionAddress(peerID core.PeerID, address string)
	PutShardID(peerID core.PeerID, shardID uint32)
	Get() map[uint32][]core.PeerID
	Contains(peerID core.PeerID) bool
	Remove(peerID core.PeerID)
	Clear()
	IsInterfaceNil() bool
}

// Closer defines the Close behavior
type Closer interface {
	Close() error
}

// ComponentHandler defines the actions common to all component handlers
type ComponentHandler interface {
	Create() error
	Close() error
	CheckSubcomponents() error
	String() string
}

// CoreComponentsHolder holds the core components
type CoreComponentsHolder interface {
	InternalMarshalizer() marshal.Marshalizer
	SetInternalMarshalizer(marshalizer marshal.Marshalizer) error
	TxMarshalizer() marshal.Marshalizer
	VmMarshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	TxSignHasher() hashing.Hasher
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverter() core.PubkeyConverter
	ValidatorPubKeyConverter() core.PubkeyConverter
	PathHandler() storage.PathManagerHandler
	Watchdog() core.WatchdogTimer
	AlarmScheduler() core.TimersScheduler
	SyncTimer() ntp.SyncTimer
	RoundHandler() consensus.RoundHandler
	EconomicsData() process.EconomicsDataHandler
	APIEconomicsData() process.EconomicsDataHandler
	RatingsData() process.RatingsInfoHandler
	Rater() sharding.PeerAccountListAndRatingHandler
	GenesisNodesSetup() sharding.GenesisNodesSetupHandler
	NodesShuffler() nodesCoordinator.NodesShuffler
	EpochNotifier() process.EpochNotifier
	EnableRoundsHandler() process.EnableRoundsHandler
	RoundNotifier() process.RoundNotifier
	EpochStartNotifierWithConfirm() EpochStartNotifierWithConfirm
	ChanStopNodeProcess() chan endProcess.ArgEndProcess
	GenesisTime() time.Time
	ChainID() string
	MinTransactionVersion() uint32
	TxVersionChecker() process.TxVersionCheckerHandler
	EncodedAddressLen() uint32
	NodeTypeProvider() core.NodeTypeProviderHandler
	WasmVMChangeLocker() common.Locker
	ProcessStatusHandler() common.ProcessStatusHandler
	HardforkTriggerPubKey() []byte
	EnableEpochsHandler() common.EnableEpochsHandler
	IsInterfaceNil() bool
}

// CoreComponentsHandler defines the core components handler actions
type CoreComponentsHandler interface {
	ComponentHandler
	CoreComponentsHolder
}

// StatusCoreComponentsHolder holds the status core components
type StatusCoreComponentsHolder interface {
	ResourceMonitor() ResourceMonitor
	NetworkStatistics() NetworkStatisticsProvider
	TrieSyncStatistics() TrieSyncStatisticsProvider
	AppStatusHandler() core.AppStatusHandler
	StatusMetrics() external.StatusMetricsHandler
	PersistentStatusHandler() PersistentStatusHandler
	StateStatsHandler() common.StateStatisticsHandler
	IsInterfaceNil() bool
}

// StatusCoreComponentsHandler defines the status core components handler actions
type StatusCoreComponentsHandler interface {
	ComponentHandler
	StatusCoreComponentsHolder
}

// CryptoParamsHolder permits access to crypto parameters such as the private and public keys
type CryptoParamsHolder interface {
	PublicKey() crypto.PublicKey
	PrivateKey() crypto.PrivateKey
	PublicKeyString() string
	PublicKeyBytes() []byte
}

// CryptoComponentsHolder holds the crypto components
type CryptoComponentsHolder interface {
	CryptoParamsHolder
	P2pPublicKey() crypto.PublicKey
	P2pPrivateKey() crypto.PrivateKey
	P2pSingleSigner() crypto.SingleSigner
	TxSingleSigner() crypto.SingleSigner
	BlockSigner() crypto.SingleSigner
	SetMultiSignerContainer(container cryptoCommon.MultiSignerContainer) error
	MultiSignerContainer() cryptoCommon.MultiSignerContainer
	GetMultiSigner(epoch uint32) (crypto.MultiSigner, error)
	PeerSignatureHandler() crypto.PeerSignatureHandler
	BlockSignKeyGen() crypto.KeyGenerator
	TxSignKeyGen() crypto.KeyGenerator
	P2pKeyGen() crypto.KeyGenerator
	MessageSignVerifier() vm.MessageSignVerifier
	ConsensusSigningHandler() consensus.SigningHandler
	ManagedPeersHolder() common.ManagedPeersHolder
	KeysHandler() consensus.KeysHandler
	Clone() interface{}
	IsInterfaceNil() bool
}

// KeyLoaderHandler defines the loading of a key from a pem file and index
type KeyLoaderHandler interface {
	LoadKey(string, int) ([]byte, string, error)
	LoadAllKeys(path string) ([][]byte, []string, error)
	IsInterfaceNil() bool
}

// CryptoComponentsHandler defines the crypto components handler actions
type CryptoComponentsHandler interface {
	ComponentHandler
	CryptoComponentsHolder
}

// MiniBlockProvider defines what a miniblock data provider should do
type MiniBlockProvider interface {
	GetMiniBlocks(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	GetMiniBlocksFromPool(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	GetMiniBlocksFromStorer(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	IsInterfaceNil() bool
}

// DataComponentsHolder holds the data components
type DataComponentsHolder interface {
	Blockchain() data.ChainHandler
	SetBlockchain(chain data.ChainHandler) error
	StorageService() dataRetriever.StorageService
	Datapool() dataRetriever.PoolsHolder
	MiniBlocksProvider() MiniBlockProvider
	Clone() interface{}
	IsInterfaceNil() bool
}

// DataComponentsHandler defines the data components handler actions
type DataComponentsHandler interface {
	ComponentHandler
	DataComponentsHolder
}

// PeerHonestyHandler defines the behaivour of a component able to handle/monitor the peer honesty of nodes which are
// participating in consensus
type PeerHonestyHandler interface {
	ChangeScore(pk string, topic string, units int)
	IsInterfaceNil() bool
	Close() error
}

// NetworkComponentsHolder holds the network components
type NetworkComponentsHolder interface {
	NetworkMessenger() p2p.Messenger
	InputAntiFloodHandler() P2PAntifloodHandler
	OutputAntiFloodHandler() P2PAntifloodHandler
	PubKeyCacher() process.TimeCacher
	PeerBlackListHandler() process.PeerBlackListCacher
	PeerHonestyHandler() PeerHonestyHandler
	PreferredPeersHolderHandler() PreferredPeersHolderHandler
	PeersRatingHandler() p2p.PeersRatingHandler
	PeersRatingMonitor() p2p.PeersRatingMonitor
	FullArchiveNetworkMessenger() p2p.Messenger
	FullArchivePreferredPeersHolderHandler() PreferredPeersHolderHandler
	IsInterfaceNil() bool
}

// NetworkComponentsHandler defines the network components handler actions
type NetworkComponentsHandler interface {
	ComponentHandler
	NetworkComponentsHolder
}

// TransactionEvaluator defines the transaction evaluator actions
type TransactionEvaluator interface {
	SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResultsWithVMOutput, error)
	ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error)
	IsInterfaceNil() bool
}

// ProcessComponentsHolder holds the process components
type ProcessComponentsHolder interface {
	NodesCoordinator() nodesCoordinator.NodesCoordinator
	ShardCoordinator() sharding.Coordinator
	InterceptorsContainer() process.InterceptorsContainer
	FullArchiveInterceptorsContainer() process.InterceptorsContainer
	ResolversContainer() dataRetriever.ResolversContainer
	RequestersFinder() dataRetriever.RequestersFinder
	RoundHandler() consensus.RoundHandler
	EpochStartTrigger() epochStart.TriggerHandler
	EpochStartNotifier() EpochStartNotifier
	ForkDetector() process.ForkDetector
	BlockProcessor() process.BlockProcessor
	BlackListHandler() process.TimeCacher
	BootStorer() process.BootStorer
	HeaderSigVerifier() process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier() process.HeaderIntegrityVerifier
	ValidatorsStatistics() process.ValidatorStatisticsProcessor
	ValidatorsProvider() process.ValidatorsProvider
	BlockTracker() process.BlockTracker
	PendingMiniBlocksHandler() process.PendingMiniBlocksHandler
	RequestHandler() process.RequestHandler
	TxLogsProcessor() process.TransactionLogProcessorDatabase
	HeaderConstructionValidator() process.HeaderConstructionValidator
	PeerShardMapper() process.NetworkShardingCollector
	FullArchivePeerShardMapper() process.NetworkShardingCollector
	FallbackHeaderValidator() process.FallbackHeaderValidator
	APITransactionEvaluator() TransactionEvaluator
	WhiteListHandler() process.WhiteListHandler
	WhiteListerVerifiedTxs() process.WhiteListHandler
	HistoryRepository() dblookupext.HistoryRepository
	ImportStartHandler() update.ImportStartHandler
	RequestedItemsHandler() dataRetriever.RequestedItemsHandler
	NodeRedundancyHandler() consensus.NodeRedundancyHandler
	CurrentEpochProvider() process.CurrentNetworkEpochProviderHandler
	ScheduledTxsExecutionHandler() process.ScheduledTxsExecutionHandler
	TxsSenderHandler() process.TxsSenderHandler
	HardforkTrigger() HardforkTrigger
	ProcessedMiniBlocksTracker() process.ProcessedMiniBlocksTracker
	ESDTDataStorageHandlerForAPI() vmcommon.ESDTNFTStorageHandler
	AccountsParser() genesis.AccountsParser
	ReceiptsRepository() ReceiptsRepository
	SentSignaturesTracker() process.SentSignaturesTracker
	EpochSystemSCProcessor() process.EpochStartSystemSCProcessor
	IsInterfaceNil() bool
}

// ProcessComponentsHandler defines the process components handler actions
type ProcessComponentsHandler interface {
	ComponentHandler
	ProcessComponentsHolder
}

// StateComponentsHandler defines the state components handler actions
type StateComponentsHandler interface {
	ComponentHandler
	StateComponentsHolder
}

// StateComponentsHolder holds the
type StateComponentsHolder interface {
	PeerAccounts() state.AccountsAdapter
	AccountsAdapter() state.AccountsAdapter
	AccountsAdapterAPI() state.AccountsAdapter
	AccountsRepository() state.AccountsRepository
	TriesContainer() common.TriesHolder
	TrieStorageManagers() map[string]common.StorageManager
	MissingTrieNodesNotifier() common.MissingTrieNodesNotifier
	Close() error
	IsInterfaceNil() bool
}

// StatusComponentsHolder holds the status components
type StatusComponentsHolder interface {
	OutportHandler() outport.OutportHandler
	SoftwareVersionChecker() statistics.SoftwareVersionChecker
	ManagedPeersMonitor() common.ManagedPeersMonitor
	IsInterfaceNil() bool
}

// StatusComponentsHandler defines the status components handler actions
type StatusComponentsHandler interface {
	ComponentHandler
	StatusComponentsHolder
	// SetForkDetector should be set before starting Polling for updates
	SetForkDetector(forkDetector process.ForkDetector) error
	StartPolling() error
	ManagedPeersMonitor() common.ManagedPeersMonitor
}

// HeartbeatV2Monitor monitors the cache of heartbeatV2 messages
type HeartbeatV2Monitor interface {
	GetHeartbeats() []heartbeatData.PubKeyHeartbeat
	IsInterfaceNil() bool
}

// HeartbeatV2ComponentsHolder holds the heartbeatV2 components
type HeartbeatV2ComponentsHolder interface {
	Monitor() HeartbeatV2Monitor
	IsInterfaceNil() bool
}

// HeartbeatV2ComponentsHandler defines the heartbeatV2 components handler actions
type HeartbeatV2ComponentsHandler interface {
	ComponentHandler
	HeartbeatV2ComponentsHolder
}

// ConsensusWorker is the consensus worker handle for the exported functionality
type ConsensusWorker interface {
	Close() error
	StartWorking()
	// AddReceivedMessageCall adds a new handler function for a received message type
	AddReceivedMessageCall(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool)
	// AddReceivedHeaderHandler adds a new handler function for a received header
	AddReceivedHeaderHandler(handler func(data.HeaderHandler))
	// RemoveAllReceivedMessagesCalls removes all the functions handlers
	RemoveAllReceivedMessagesCalls()
	// ProcessReceivedMessage method redirects the received message to the channel which should handle it
	ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error
	// Extend does an extension for the subround with subroundId
	Extend(subroundId int)
	// GetConsensusStateChangedChannel gets the channel for the consensusStateChanged
	GetConsensusStateChangedChannel() chan bool
	// ExecuteStoredMessages tries to execute all the messages received which are valid for execution
	ExecuteStoredMessages()
	// DisplayStatistics method displays statistics of worker at the end of the round
	DisplayStatistics()
	// ResetConsensusMessages resets at the start of each round all the previous consensus messages received
	ResetConsensusMessages()
	// ReceivedHeader method is a wired method through which worker will receive headers from network
	ReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte)
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// HardforkTrigger defines the hard-fork trigger functionality
type HardforkTrigger interface {
	SetExportFactoryHandler(exportFactoryHandler update.ExportFactoryHandler) error
	TriggerReceived(payload []byte, data []byte, pkBytes []byte) (bool, error)
	RecordedTriggerMessage() ([]byte, bool)
	Trigger(epoch uint32, withEarlyEndOfEpoch bool) error
	CreateData() []byte
	AddCloser(closer update.Closer) error
	NotifyTriggerReceivedV2() <-chan struct{}
	IsSelfTrigger() bool
	IsInterfaceNil() bool
}

// ConsensusComponentsHolder holds the consensus components
type ConsensusComponentsHolder interface {
	Chronology() consensus.ChronologyHandler
	ConsensusWorker() ConsensusWorker
	BroadcastMessenger() consensus.BroadcastMessenger
	ConsensusGroupSize() (int, error)
	Bootstrapper() process.Bootstrapper
	IsInterfaceNil() bool
}

// ConsensusComponentsHandler defines the consensus components handler actions
type ConsensusComponentsHandler interface {
	ComponentHandler
	ConsensusComponentsHolder
}

// BootstrapParamsHolder gives read access to parameters after bootstrap
type BootstrapParamsHolder interface {
	Epoch() uint32
	SelfShardID() uint32
	NumOfShards() uint32
	NodesConfig() nodesCoordinator.NodesCoordinatorRegistryHandler
	IsInterfaceNil() bool
}

// EpochStartBootstrapper defines the epoch start bootstrap functionality
type EpochStartBootstrapper interface {
	Bootstrap() (bootstrap.Parameters, error)
	IsInterfaceNil() bool
	Close() error
}

// BootstrapComponentsHolder holds the bootstrap components
type BootstrapComponentsHolder interface {
	EpochStartBootstrapper() EpochStartBootstrapper
	EpochBootstrapParams() BootstrapParamsHolder
	NodeType() core.NodeType
	ShardCoordinator() sharding.Coordinator
	VersionedHeaderFactory() factory.VersionedHeaderFactory
	HeaderVersionHandler() factory.HeaderVersionHandler
	HeaderIntegrityVerifier() factory.HeaderIntegrityVerifierHandler
	GuardedAccountHandler() process.GuardedAccountHandler
	NodesCoordinatorRegistryFactory() nodesCoordinator.NodesCoordinatorRegistryFactory
	IsInterfaceNil() bool
}

// BootstrapComponentsHandler defines the bootstrap components handler actions
type BootstrapComponentsHandler interface {
	ComponentHandler
	BootstrapComponentsHolder
}

// ShuffleOutCloser defines the action for end of processing
type ShuffleOutCloser interface {
	EndOfProcessingHandler(event endProcess.ArgEndProcess) error
	IsInterfaceNil() bool
	Close() error
}

// EconomicsHandler provides some economics related computation and read access to economics data
type EconomicsHandler interface {
	LeaderPercentage() float64
	ProtocolSustainabilityPercentage() float64
	ProtocolSustainabilityAddress() string
	MinInflationRate() float64
	MaxInflationRate(year uint32) float64
	DeveloperPercentage() float64
	GenesisTotalSupply() *big.Int
	MaxGasLimitPerBlock(shardID uint32) uint64
	ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValues(tx data.TransactionWithFeeHandler) error
	MinGasPrice() uint64
	MinGasLimit() uint64
	GasPerDataByte() uint64
	GasPriceModifier() float64
	ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	IsInterfaceNil() bool
}

// LogsFacade defines the interface of a logs facade
type LogsFacade interface {
	GetLog(logKey []byte, epoch uint32) (*transaction.ApiLogs, error)
	IncludeLogsInTransactions(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error
	IsInterfaceNil() bool
}

// ReceiptsRepository defines the interface of a receiptsRepository
type ReceiptsRepository interface {
	SaveReceipts(holder common.ReceiptsHolder, header data.HeaderHandler, headerHash []byte) error
	LoadReceipts(header data.HeaderHandler, headerHash []byte) (common.ReceiptsHolder, error)
	IsInterfaceNil() bool
}

// ProcessDebuggerSetter allows setting a debugger on the process component
type ProcessDebuggerSetter interface {
	SetProcessDebugger(debugger process.Debugger) error
}

// ResourceMonitor defines the function implemented by a struct that can monitor resources
type ResourceMonitor interface {
	Close() error
	IsInterfaceNil() bool
}

// NetworkStatisticsProvider is able to provide network statistics
type NetworkStatisticsProvider interface {
	BpsSent() uint64
	BpsRecv() uint64
	BpsSentPeak() uint64
	BpsRecvPeak() uint64
	PercentSent() uint64
	PercentRecv() uint64
	TotalBytesSentInCurrentEpoch() uint64
	TotalBytesReceivedInCurrentEpoch() uint64
	TotalSentInCurrentEpoch() string
	TotalReceivedInCurrentEpoch() string
	EpochConfirmed(epoch uint32, timestamp uint64)
	Close() error
	IsInterfaceNil() bool
}

// TrieSyncStatisticsProvider is able to provide trie sync statistics
type TrieSyncStatisticsProvider interface {
	data.SyncStatisticsHandler
	AddNumBytesReceived(bytes uint64)
	NumBytesReceived() uint64
	NumTries() int
	AddProcessingTime(duration time.Duration)
	IncrementIteration()
	ProcessingTime() time.Duration
	NumIterations() int
}

// PersistentStatusHandler defines a persistent status handler
type PersistentStatusHandler interface {
	core.AppStatusHandler
	SetStorage(store storage.Storer) error
}

// RunTypeComponentsHandler defines the run type components handler actions
type RunTypeComponentsHandler interface {
	ComponentHandler
	RunTypeComponentsHolder
}

// RunTypeComponentsHolder holds the run type components
type RunTypeComponentsHolder interface {
	BlockChainHookHandlerCreator() hooks.BlockChainHookHandlerCreator
	BlockProcessorCreator() processBlock.BlockProcessorCreator
	BlockTrackerCreator() track.BlockTrackerCreator
	BootstrapperFromStorageCreator() storageBootstrap.BootstrapperFromStorageCreator
	BootstrapperCreator() storageBootstrap.BootstrapperCreator
	EpochStartBootstrapperCreator() bootstrap.EpochStartBootstrapperCreator
	ForkDetectorCreator() sync.ForkDetectorCreator
	HeaderValidatorCreator() processBlock.HeaderValidatorCreator
	RequestHandlerCreator() requestHandlers.RequestHandlerCreator
	ScheduledTxsExecutionCreator() preprocess.ScheduledTxsExecutionCreator
	TransactionCoordinatorCreator() coordinator.TransactionCoordinatorCreator
	ValidatorStatisticsProcessorCreator() peer.ValidatorStatisticsProcessorCreator
	AdditionalStorageServiceCreator() process.AdditionalStorageServiceCreator
	SCProcessorCreator() scrCommon.SCProcessorCreator
	SCResultsPreProcessorCreator() preprocess.SmartContractResultPreProcessorCreator
	ConsensusModel() consensus.ConsensusModel
	VmContainerMetaFactoryCreator() factoryVm.VmContainerCreator
	VmContainerShardFactoryCreator() factoryVm.VmContainerCreator
	AccountsParser() genesis.AccountsParser
	AccountsCreator() state.AccountFactory
	VMContextCreator() systemSmartContracts.VMContextCreatorHandler
	OutGoingOperationsPoolHandler() sovereignBlock.OutGoingOperationsPool
	DataCodecHandler() sovereign.DataCodecHandler
	TopicsCheckerHandler() sovereign.TopicsCheckerHandler
	ShardCoordinatorCreator() sharding.ShardCoordinatorFactory
	NodesCoordinatorWithRaterCreator() nodesCoordinator.NodesCoordinatorWithRaterFactory
	RequestersContainerFactoryCreator() requesterscontainer.RequesterContainerFactoryCreator
	InterceptorsContainerFactoryCreator() interceptorscontainer.InterceptorsContainerFactoryCreator
	ShardResolversContainerFactoryCreator() resolverscontainer.ShardResolversContainerFactoryCreator
	TxPreProcessorCreator() preprocess.TxPreProcessorCreator
	ExtraHeaderSigVerifierHolder() headerCheck.ExtraHeaderSigVerifierHolder
	GenesisBlockCreatorFactory() processComp.GenesisBlockCreatorFactory
	GenesisMetaBlockCheckerCreator() processComp.GenesisMetaBlockChecker
	Create() error
	Close() error
	CheckSubcomponents() error
	String() string
	IsInterfaceNil() bool
}
