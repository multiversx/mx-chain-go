package factory

import (
	"context"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/statistics"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	heartbeatData "github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/outport"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/vm"
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
	StatusHandlerUtils() factory.StatusHandlersUtils
	StatusHandler() core.AppStatusHandler
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
	RoundNotifier() process.RoundNotifier
	EpochStartNotifierWithConfirm() EpochStartNotifierWithConfirm
	ChanStopNodeProcess() chan endProcess.ArgEndProcess
	GenesisTime() time.Time
	ChainID() string
	MinTransactionVersion() uint32
	TxVersionChecker() process.TxVersionCheckerHandler
	EncodedAddressLen() uint32
	NodeTypeProvider() core.NodeTypeProviderHandler
	ArwenChangeLocker() common.Locker
	ProcessStatusHandler() common.ProcessStatusHandler
	HardforkTriggerPubKey() []byte
	IsInterfaceNil() bool
}

// CoreComponentsHandler defines the core components handler actions
type CoreComponentsHandler interface {
	ComponentHandler
	CoreComponentsHolder
}

// CryptoParamsHolder permits access to crypto parameters such as the private and public keys
type CryptoParamsHolder interface {
	PublicKey() crypto.PublicKey
	PrivateKey() crypto.PrivateKey
	PublicKeyString() string
	PublicKeyBytes() []byte
	PrivateKeyBytes() []byte
}

// CryptoComponentsHolder holds the crypto components
type CryptoComponentsHolder interface {
	CryptoParamsHolder
	TxSingleSigner() crypto.SingleSigner
	BlockSigner() crypto.SingleSigner
	MultiSigner() crypto.MultiSigner
	PeerSignatureHandler() crypto.PeerSignatureHandler
	SetMultiSigner(ms crypto.MultiSigner) error
	BlockSignKeyGen() crypto.KeyGenerator
	TxSignKeyGen() crypto.KeyGenerator
	MessageSignVerifier() vm.MessageSignVerifier
	Clone() interface{}
	IsInterfaceNil() bool
}

// KeyLoaderHandler defines the loading of a key from a pem file and index
type KeyLoaderHandler interface {
	LoadKey(string, int) ([]byte, string, error)
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
	SetBlockchain(chain data.ChainHandler)
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
	IsInterfaceNil() bool
}

// NetworkComponentsHandler defines the network components handler actions
type NetworkComponentsHandler interface {
	ComponentHandler
	NetworkComponentsHolder
}

// TransactionSimulatorProcessor defines the actions which a transaction simulator processor has to implement
type TransactionSimulatorProcessor interface {
	ProcessTx(tx *transaction.Transaction) (*txSimData.SimulationResults, error)
	IsInterfaceNil() bool
}

// ProcessComponentsHolder holds the process components
type ProcessComponentsHolder interface {
	NodesCoordinator() nodesCoordinator.NodesCoordinator
	ShardCoordinator() sharding.Coordinator
	InterceptorsContainer() process.InterceptorsContainer
	ResolversFinder() dataRetriever.ResolversFinder
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
	FallbackHeaderValidator() process.FallbackHeaderValidator
	TransactionSimulatorProcessor() TransactionSimulatorProcessor
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
	TriesContainer() common.TriesHolder
	TrieStorageManagers() map[string]common.StorageManager
	IsInterfaceNil() bool
}

// StatusComponentsHolder holds the status components
type StatusComponentsHolder interface {
	OutportHandler() outport.OutportHandler
	SoftwareVersionChecker() statistics.SoftwareVersionChecker
	IsInterfaceNil() bool
}

// StatusComponentsHandler defines the status components handler actions
type StatusComponentsHandler interface {
	ComponentHandler
	StatusComponentsHolder
	// SetForkDetector should be set before starting Polling for updates
	SetForkDetector(forkDetector process.ForkDetector)
	StartPolling() error
}

// HeartbeatSender sends heartbeat messages
type HeartbeatSender interface {
	SendHeartbeat() error
	IsInterfaceNil() bool
}

// HeartbeatMonitor monitors the received heartbeat messages
type HeartbeatMonitor interface {
	ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	GetHeartbeats() []heartbeatData.PubKeyHeartbeat
	IsInterfaceNil() bool
	Cleanup()
	Close() error
}

// HeartbeatStorer provides storage functionality for the heartbeat component
type HeartbeatStorer interface {
	UpdateGenesisTime(genesisTime time.Time) error
	LoadGenesisTime() (time.Time, error)
	SaveKeys(peersSlice [][]byte) error
	LoadKeys() ([][]byte, error)
	IsInterfaceNil() bool
}

// HeartbeatComponentsHolder holds the heartbeat components
type HeartbeatComponentsHolder interface {
	MessageHandler() heartbeat.MessageHandler
	Monitor() HeartbeatMonitor
	Sender() HeartbeatSender
	Storer() HeartbeatStorer
	IsInterfaceNil() bool
}

// HeartbeatComponentsHandler defines the heartbeat components handler actions
type HeartbeatComponentsHandler interface {
	ComponentHandler
	HeartbeatComponentsHolder
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
	ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
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
	NotifyTriggerReceived() <-chan struct{}
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
	NodesConfig() *nodesCoordinator.NodesCoordinatorRegistry
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
	RoundActivationHandler() process.RoundActivationHandler
	EpochStartBootstrapper() EpochStartBootstrapper
	EpochBootstrapParams() BootstrapParamsHolder
	NodeType() core.NodeType
	ShardCoordinator() sharding.Coordinator
	VersionedHeaderFactory() factory.VersionedHeaderFactory
	HeaderVersionHandler() factory.HeaderVersionHandler
	HeaderIntegrityVerifier() factory.HeaderIntegrityVerifierHandler
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
