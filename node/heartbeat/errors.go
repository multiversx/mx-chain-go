package heartbeat

import "errors"

// ErrEmptyPublicKeysMap signals that a nil or empty public keys map has been provided
var ErrEmptyPublicKeysMap = errors.New("nil or empty public keys map")

// ErrNilMessenger signals that a nil p2p messenger has been provided
var ErrNilMessenger = errors.New("nil P2P Messenger")

// ErrNilSingleSigner signals that a nil single signer has been provided
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("nil private key")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil used
var ErrNilKeyGenerator = errors.New("key generator is nil")

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

// ErrNilNetworkShardingUpdater defines the error for setting a nil network sharding updater
var ErrNilNetworkShardingUpdater = errors.New("nil network sharding updater")
