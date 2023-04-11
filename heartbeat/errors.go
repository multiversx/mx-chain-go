package heartbeat

import "errors"

// ErrNilMessenger signals that a nil p2p messenger has been provided
var ErrNilMessenger = errors.New("nil P2P Messenger")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("nil private key")

// ErrNilMarshaller signals that a nil marshaller has been provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilAppStatusHandler defines the error for setting a nil AppStatusHandler
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")

// ErrNilPeerTypeProvider signals that a nil peer type provider has been given
var ErrNilPeerTypeProvider = errors.New("nil peer type provider")

// ErrPropertyTooLong signals that one of the properties is too long
var ErrPropertyTooLong = errors.New("property too long in Heartbeat")

// ErrNilHardforkTrigger signals that a nil hardfork trigger has been provided
var ErrNilHardforkTrigger = errors.New("nil hardfork trigger")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to use a nil pubkey converter")

// ErrNilPeerSignatureHandler signals that a nil peerSignatureHandler object has been provided
var ErrNilPeerSignatureHandler = errors.New("trying to set nil peerSignatureHandler")

// ErrNilCurrentBlockProvider signals that a nil current block provider
var ErrNilCurrentBlockProvider = errors.New("nil current block provider")

// ErrNilRedundancyHandler signals that a nil redundancy handler was provided
var ErrNilRedundancyHandler = errors.New("nil redundancy handler")

// ErrEmptySendTopic signals that an empty topic string was provided
var ErrEmptySendTopic = errors.New("empty topic for sending messages")

// ErrInvalidTimeDuration signals that an invalid time duration was provided
var ErrInvalidTimeDuration = errors.New("invalid time duration")

// ErrInvalidThreshold signals that an invalid threshold was provided
var ErrInvalidThreshold = errors.New("invalid threshold")

// ErrNilRequestHandler signals that a nil request handler interface was provided
var ErrNilRequestHandler = errors.New("nil request handler")

// ErrNilNodesCoordinator signals that an operation has been attempted to or with a nil nodes coordinator
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrNilPeerAuthenticationPool signals that a nil peer authentication pool has been provided
var ErrNilPeerAuthenticationPool = errors.New("nil peer authentication pool")

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilRandomizer signals that a nil randomizer has been provided
var ErrNilRandomizer = errors.New("nil randomizer")

// ErrNilCacher signals that a nil cache has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilPeerShardMapper signals that a nil peer shard mapper has been provided
var ErrNilPeerShardMapper = errors.New("nil peer shard mapper")

// ErrShouldSkipValidator signals that the validator should be skipped
var ErrShouldSkipValidator = errors.New("validator should be skipped")

// ErrNilHeartbeatMonitor signals that a nil heartbeat monitor was provided
var ErrNilHeartbeatMonitor = errors.New("nil heartbeat monitor")

// ErrNilHeartbeatSenderInfoProvider signals that a nil heartbeat sender info provider was provided
var ErrNilHeartbeatSenderInfoProvider = errors.New("nil heartbeat sender info provider")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilTrieSyncStatisticsProvider signals that a nil trie sync statistics provider was provided
var ErrNilTrieSyncStatisticsProvider = errors.New("nil trie sync statistics provider")

// ErrNilManagedPeersHolder signals that a nil managed peers holder has been provided
var ErrNilManagedPeersHolder = errors.New("nil managed peers holder")

// ErrInvalidConfiguration signals that an invalid configuration has been provided
var ErrInvalidConfiguration = errors.New("invalid configuration")
