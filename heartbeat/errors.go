package heartbeat

import "errors"

// ErrNilPublicKeysMap signals that a nil public keys map has been provided
var ErrNilPublicKeysMap = errors.New("nil public keys map")

// ErrNilMessenger signals that a nil p2p messenger has been provided
var ErrNilMessenger = errors.New("nil P2P Messenger")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("nil private key")

// ErrNilMarshaller signals that a nil marshaller has been provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")

// ErrInvalidMaxDurationPeerUnresponsive signals that the duration provided is invalid
var ErrInvalidMaxDurationPeerUnresponsive = errors.New("invalid max duration to declare the peer unresponsive")

// ErrNilAppStatusHandler defines the error for setting a nil AppStatusHandler
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")

// ErrNilShardCoordinator signals that an operation has been attempted to or with a nil shard coordinator
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilTimer signals that a nil time getter handler has been provided
var ErrNilTimer = errors.New("nil time getter handler")

// ErrNilPeerTypeProvider signals that a nil peer type provider has been given
var ErrNilPeerTypeProvider = errors.New("nil peer type provider")

// ErrNilMonitorDb signals that a nil monitor db was provided
var ErrNilMonitorDb = errors.New("nil monitor db")

// ErrNilMessageHandler signals that the provided message handler is nil
var ErrNilMessageHandler = errors.New("nil message handler")

// ErrNilHeartbeatStorer signals that the provided heartbeat storer is nil
var ErrNilHeartbeatStorer = errors.New("nil heartbeat storer")

// ErrFetchGenesisTimeFromDb signals that the genesis time cannot be fetched from db
var ErrFetchGenesisTimeFromDb = errors.New("monitor: can't get genesis time from db")

// ErrStoreGenesisTimeToDb signals that the genesis time cannot be store to db
var ErrStoreGenesisTimeToDb = errors.New("monitor: can't store genesis time")

// ErrUnmarshalGenesisTime signals that the unmarshaling of the genesis time didn't work
var ErrUnmarshalGenesisTime = errors.New("monitor: can't unmarshal genesis time")

// ErrMarshalGenesisTime signals that the marshaling of the genesis time didn't work
var ErrMarshalGenesisTime = errors.New("monitor: can't marshal genesis time")

// ErrPropertyTooLong signals that one of the properties is too long
var ErrPropertyTooLong = errors.New("property too long in Heartbeat")

// ErrNilNetworkShardingCollector defines the error for setting a nil network sharding collector
var ErrNilNetworkShardingCollector = errors.New("nil network sharding collector")

// ErrNilAntifloodHandler signals that a nil antiflood handler has been provided
var ErrNilAntifloodHandler = errors.New("nil antiflood handler")

// ErrNilHardforkTrigger signals that a nil hardfork trigger has been provided
var ErrNilHardforkTrigger = errors.New("nil hardfork trigger")

// ErrHeartbeatPidMismatch signals that a received hearbeat did not come from the correct originator
var ErrHeartbeatPidMismatch = errors.New("heartbeat peer id mismatch")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to use a nil pubkey converter")

// ErrZeroHeartbeatRefreshIntervalInSec signals that a zero value was provided for the HeartbeatRefreshIntervalInSec
var ErrZeroHeartbeatRefreshIntervalInSec = errors.New("zero heartbeatRefreshInterval")

// ErrZeroHideInactiveValidatorIntervalInSec signals that a zero value
// was provided for the ErrZeroHideInactiveValidatorIntervalInSec
var ErrZeroHideInactiveValidatorIntervalInSec = errors.New("zero hideInactiveValidatorIntervalInSec")

// ErrInvalidDurationToConsiderUnresponsiveInSec is raised when a value less than 1 has been provided
var ErrInvalidDurationToConsiderUnresponsiveInSec = errors.New("value DurationToConsiderUnresponsiveInSec is less than 1")

// ErrNegativeMaxTimeToWaitBetweenBroadcastsInSec is raised when a value less than 1 has been provided
var ErrNegativeMaxTimeToWaitBetweenBroadcastsInSec = errors.New("value MaxTimeToWaitBetweenBroadcastsInSec is less than 1")

// ErrNegativeMinTimeToWaitBetweenBroadcastsInSec is raised when a value less than 1 has been provided
var ErrNegativeMinTimeToWaitBetweenBroadcastsInSec = errors.New("value MinTimeToWaitBetweenBroadcastsInSec is less than 1")

// ErrWrongValues signals that wrong values were provided
var ErrWrongValues = errors.New("wrong values for heartbeat parameters")

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

// ErrNilEnableEpochsHandler signals that a nil enable epochs handler has been provided
var ErrNilEnableEpochsHandler = errors.New("nil enable epochs handler")

// ErrShouldSkipValidator signals that the validator should be skipped
var ErrShouldSkipValidator = errors.New("validator should be skipped")

// ErrNilHeartbeatMonitor signals that a nil heartbeat monitor was provided
var ErrNilHeartbeatMonitor = errors.New("nil heartbeat monitor")

// ErrNilHeartbeatSenderInfoProvider signals that a nil heartbeat sender info provider was provided
var ErrNilHeartbeatSenderInfoProvider = errors.New("nil heartbeat sender info provider")

// ErrNilTrieSyncStatisticsProvider signals that a nil trie sync statistics provider was provided
var ErrNilTrieSyncStatisticsProvider = errors.New("nil trie sync statistics provider")
