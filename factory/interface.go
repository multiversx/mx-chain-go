package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/process"
)

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	IsInterfaceNil() bool
}

// NodesSetupHandler defines which actions should be done for handling initial nodes setup
type NodesSetupHandler interface {
	InitialNodesPubKeys() map[uint32][]string
	InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error)
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	SetDebugger(debugger process.AntifloodDebugger) error
	ApplyConsensusSize(size int)
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsInterfaceNil() bool
}

// HeaderIntegrityVerifierHandler is the interface needed to check that a header's integrity is correct
type HeaderIntegrityVerifierHandler interface {
	Verify(header data.HeaderHandler) error
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
}

// CoreComponentsHolder holds the core components
type CoreComponentsHolder interface {
	InternalMarshalizer() marshal.Marshalizer
	SetInternalMarshalizer(marshalizer marshal.Marshalizer) error
	TxMarshalizer() marshal.Marshalizer
	VmMarshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverter() state.PubkeyConverter
	ValidatorPubKeyConverter() state.PubkeyConverter
	StatusHandler() core.AppStatusHandler
	SetStatusHandler(statusHandler core.AppStatusHandler) error
	PathHandler() storage.PathManagerHandler
	ChainID() string
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
	SetMultiSigner(ms crypto.MultiSigner) error
	BlockSignKeyGen() crypto.KeyGenerator
	TxSignKeyGen() crypto.KeyGenerator
	MessageSignVerifier() vm.MessageSignVerifier
	IsInterfaceNil() bool
}

// CryptoComponentsHandler defines the crypto components handler actions
type CryptoComponentsHandler interface {
	ComponentHandler
	CryptoComponentsHolder
}

// DataComponentsHolder holds the data components
type DataComponentsHolder interface {
	Blockchain() data.ChainHandler
	SetBlockchain(chain data.ChainHandler)
	StorageService() dataRetriever.StorageService
	Datapool() dataRetriever.PoolsHolder
	IsInterfaceNil() bool
}

// DataComponentsHandler defines the data components handler actions
type DataComponentsHandler interface {
	ComponentHandler
	DataComponentsHolder
}

// NetworkComponentsHolder holds the network components
type NetworkComponentsHolder interface {
	NetworkMessenger() p2p.Messenger
	InputAntiFloodHandler() P2PAntifloodHandler
	OutputAntiFloodHandler() P2PAntifloodHandler
	PeerBlackListHandler() process.BlackListHandler
	IsInterfaceNil() bool
}

// NetworkComponentsHandler defines the network components handler actions
type NetworkComponentsHandler interface {
	ComponentHandler
	NetworkComponentsHolder
}

// ProcessComponentsHolder holds the process components
type ProcessComponentsHolder interface {
	InterceptorsContainer() process.InterceptorsContainer
	ResolversFinder() dataRetriever.ResolversFinder
	Rounder() consensus.Rounder
	EpochStartTrigger() epochStart.TriggerHandler
	ForkDetector() process.ForkDetector
	BlockProcessor() process.BlockProcessor
	BlackListHandler() process.BlackListHandler
	BootStorer() process.BootStorer
	HeaderSigVerifier() process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier() process.HeaderIntegrityVerifier
	ValidatorsStatistics() process.ValidatorStatisticsProcessor
	ValidatorsProvider() process.ValidatorsProvider
	BlockTracker() process.BlockTracker
	PendingMiniBlocksHandler() process.PendingMiniBlocksHandler
	RequestHandler() process.RequestHandler
	TxLogsProcessor() process.TransactionLogProcessorDatabase
	IsInterfaceNil() bool
}

// ProcessComponentsHandler defines the process components handler actions
type ProcessComponentsHandler interface {
	ComponentHandler
	ProcessComponentsHolder
}

// StatusComponentsHolder holds the status components
type StatusComponentsHolder interface {
	TpsBenchmark() statistics.TPSBenchmark
	ElasticIndexer() indexer.Indexer
	SoftwareVersionChecker() statistics.SoftwareVersionChecker
	StatusHandler() core.AppStatusHandler
	IsInterfaceNil() bool
}

// StateComponentsHandler defines the status components handler actions
type StateComponentsHandler interface {
	ComponentHandler
	StatusComponentsHolder
}
