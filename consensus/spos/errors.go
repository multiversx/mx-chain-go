package spos

import (
	"errors"
)

// ErrNilConsensusGroup is raised when an operation is attempted with a nil consensus group
var ErrNilConsensusGroup = errors.New("consensusGroup is null")

// ErrEmptyConsensusGroup is raised when an operation is attempted with an empty consensus group
var ErrEmptyConsensusGroup = errors.New("consensusGroup is empty")

// ErrNotFoundInConsensus is raised when self expected in consensus group but not found
var ErrNotFoundInConsensus = errors.New("self not found in consensus group")

// ErrNilSignature is raised when a valid signature was expected but nil was used
var ErrNilSignature = errors.New("signature is nil")

// ErrNilMultiSigner is raised when a valid multiSigner is expected but nil used
var ErrNilMultiSigner = errors.New("multiSigner is nil")

// ErrNilMultiSignerContainer is raised when a valid multiSigner container is expected, but nil used
var ErrNilMultiSignerContainer = errors.New("multiSigner container is nil")

// ErrNilConsensusState is raised when a valid consensus is expected but nil used
var ErrNilConsensusState = errors.New("consensus state is nil")

// ErrNilConsensusCore is raised when a valid ConsensusCore is expected but nil used
var ErrNilConsensusCore = errors.New("consensus core is nil")

// ErrNilConsensusService is raised when a valid ConsensusService is expected but nil used
var ErrNilConsensusService = errors.New("consensus service is nil")

// ErrNilBlockChain is raised when a valid blockchain is expected but nil used
var ErrNilBlockChain = errors.New("blockchain is nil")

// ErrNilHasher is raised when a valid hasher is expected but nil used
var ErrNilHasher = errors.New("hasher is nil")

// ErrNilMarshalizer is raised when a valid marshalizer is expected but nil used
var ErrNilMarshalizer = errors.New("marshalizer is nil")

// ErrNilMessenger is raised when a valid messenger is expected but nil used
var ErrNilMessenger = errors.New("messenger is nil")

// ErrNilBlockProcessor is raised when a valid block processor is expected but nil used
var ErrNilBlockProcessor = errors.New("block processor is nil")

// ErrNilBootstrapper is raised when a valid block processor is expected but nil used
var ErrNilBootstrapper = errors.New("bootstrapper is nil")

// ErrNilBroadcastMessenger is raised when a valid broadcast messenger is expected but nil used
var ErrNilBroadcastMessenger = errors.New("broadcast messenger is nil")

// ErrNilHeadersSubscriber is raised when a valid headers subscriber is expected but nil is provided
var ErrNilHeadersSubscriber = errors.New("headers subscriber is nil")

// ErrNilAlarmScheduler is raised when a valid alarm scheduler is expected but nil is provided
var ErrNilAlarmScheduler = errors.New("alarm scheduler is nil")

// ErrInvalidKey is raised when an invalid key is used with a map
var ErrInvalidKey = errors.New("map key is invalid")

// ErrNilRoundState is raised when a valid round state is expected but nil used
var ErrNilRoundState = errors.New("round state is nil")

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")

// ErrNilWorker is raised when a valid Worker is expected but nil used
var ErrNilWorker = errors.New("worker is nil")

// ErrNilWorkerArgs signals that nil a workerArgs has been provided
var ErrNilWorkerArgs = errors.New("worker args is nil")

// ErrNilShardCoordinator is raised when a valid shard coordinator is expected but nil used
var ErrNilShardCoordinator = errors.New("shard coordinator is nil")

// ErrNilNodesCoordinator is raised when a valid validator group selector is expected but nil used
var ErrNilNodesCoordinator = errors.New("validator group selector is nil")

// ErrNilInterceptorsContainer is raised when a nil interceptor container is provided
var ErrNilInterceptorsContainer = errors.New("interceptor container is nil")

// ErrNilParameter is raised when a nil parameter is provided
var ErrNilParameter = errors.New("parameter is nil")

// ErrNilChronologyHandler is raised when a valid chronology handler is expected but nil used
var ErrNilChronologyHandler = errors.New("chronology handler is nil")

// ErrNilRoundHandler is raised when a valid roundHandler is expected but nil used
var ErrNilRoundHandler = errors.New("roundHandler is nil")

// ErrNilSyncTimer is raised when a valid sync timer is expected but nil used
var ErrNilSyncTimer = errors.New("sync timer is nil")

// ErrNilSubround is raised when a valid subround is expected but nil used
var ErrNilSubround = errors.New("subround is nil")

// ErrNilChannel is raised when a valid channel is expected but nil used
var ErrNilChannel = errors.New("channel is nil")

// ErrRoundCanceled is raised when round is canceled
var ErrRoundCanceled = errors.New("round is canceled")

// ErrNodeIsNotInEligibleList is raised when a node is not in eligible list
var ErrNodeIsNotInEligibleList = errors.New("node is not in eligible list")

// ErrMessageForPastRound is raised when message is for past round
var ErrMessageForPastRound = errors.New("message is for past round")

// ErrMessageForFutureRound is raised when message is for future round
var ErrMessageForFutureRound = errors.New("message is for future round")

// ErrInvalidSignature is raised when signature is invalid
var ErrInvalidSignature = errors.New("signature is invalid")

// ErrInvalidHeader is raised when header is invalid
var ErrInvalidHeader = errors.New("header is invalid")

// ErrMessageFromItself is raised when a message from itself is received
var ErrMessageFromItself = errors.New("message is from itself")

// ErrNilHeader is raised when an expected header is nil
var ErrNilHeader = errors.New("header is nil")

// ErrNilHeaderHash is raised when a nil header hash is provided
var ErrNilHeaderHash = errors.New("header hash is nil")

// ErrNilBody is raised when an expected body is nil
var ErrNilBody = errors.New("body is nil")

// ErrNilMetaHeader is raised when an expected meta header is nil
var ErrNilMetaHeader = errors.New("meta header is nil")

// ErrInvalidMetaHeader is raised when an invalid meta header was provided
var ErrInvalidMetaHeader = errors.New("meta header is invalid")

// ErrNilForkDetector is raised when a valid fork detector is expected but nil used
var ErrNilForkDetector = errors.New("fork detector is nil")

// ErrNilExecuteStoredMessages is raised when a valid executeStoredMessages function is expected but nil used
var ErrNilExecuteStoredMessages = errors.New("executeStoredMessages is nil")

// ErrNilAppStatusHandler defines the error for setting a nil AppStatusHandler
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")

// ErrNilAntifloodHandler signals that a nil antiflood handler has been provided
var ErrNilAntifloodHandler = errors.New("nil antiflood handler")

// ErrNilPoolAdder signals that a nil pool adder has been provided
var ErrNilPoolAdder = errors.New("nil pool adder")

// ErrNilHeaderSigVerifier signals that a nil header sig verifier has been provided
var ErrNilHeaderSigVerifier = errors.New("nil header sig verifier")

// ErrNilHeaderIntegrityVerifier signals that a nil header integrity verifier has been provided
var ErrNilHeaderIntegrityVerifier = errors.New("nil header integrity verifier")

// ErrInvalidChainID signals that an invalid chain ID has been provided
var ErrInvalidChainID = errors.New("invalid chain ID in consensus")

// ErrNilNetworkShardingCollector defines the error for setting a nil network sharding collector
var ErrNilNetworkShardingCollector = errors.New("nil network sharding collector")

// ErrInvalidMessageType signals that an invalid message type has been received from consensus topic
var ErrInvalidMessageType = errors.New("invalid message type")

// ErrInvalidHeaderHashSize signals that an invalid header hash size has been received from consensus topic
var ErrInvalidHeaderHashSize = errors.New("invalid header hash size")

// ErrInvalidBodySize signals that an invalid body size has been received from consensus topic
var ErrInvalidBodySize = errors.New("invalid body size")

// ErrInvalidHeaderSize signals that an invalid header size has been received from consensus topic
var ErrInvalidHeaderSize = errors.New("invalid header size")

// ErrInvalidPublicKeySize signals that an invalid public key size has been received from consensus topic
var ErrInvalidPublicKeySize = errors.New("invalid public key size")

// ErrInvalidSignatureSize signals that an invalid signature size has been received from consensus topic
var ErrInvalidSignatureSize = errors.New("invalid signature size")

// ErrInvalidMessage signals that an invalid message has been received from consensus topic
var ErrInvalidMessage = errors.New("invalid message")

// ErrInvalidPublicKeyBitmapSize signals that an invalid public key bitmap size has been received from consensus topic
var ErrInvalidPublicKeyBitmapSize = errors.New("invalid public key bitmap size")

// ErrInvalidCacheSize signals an invalid size provided for cache
var ErrInvalidCacheSize = errors.New("invalid cache size")

// ErrNilPeerHonestyHandler signals that a nil peer honesty handler has been provided
var ErrNilPeerHonestyHandler = errors.New("nil peer honesty handler")

// ErrOriginatorMismatch signals that an original consensus message has been re-broadcast manually by another peer
var ErrOriginatorMismatch = errors.New("consensus message originator mismatch")

// ErrNilPeerSignatureHandler signals that a nil peerSignatureHandler object has been provided
var ErrNilPeerSignatureHandler = errors.New("trying to set nil peerSignatureHandler")

// ErrMessageTypeLimitReached signals that a consensus message type limit has been reached for a public key
var ErrMessageTypeLimitReached = errors.New("consensus message type limit has been reached")

// ErrNilFallbackHeaderValidator signals that a nil fallback header validator has been provided
var ErrNilFallbackHeaderValidator = errors.New("nil fallback header validator")

// ErrNilNodeRedundancyHandler signals that provided node redundancy handler is nil
var ErrNilNodeRedundancyHandler = errors.New("nil node redundancy handler")

// ErrNilScheduledProcessor signals that the provided scheduled processor is nil
var ErrNilScheduledProcessor = errors.New("nil scheduled processor")

// ErrInvalidNumSigShares signals that an invalid number of signature shares has been provided
var ErrInvalidNumSigShares = errors.New("invalid number of sig shares")

// ErrNilMessageSigningHandler signals that the provided message signing handler is nil
var ErrNilMessageSigningHandler = errors.New("nil message signing handler")

// ErrNilPeerBlacklistHandler signals that the provided peer blacklist handler is nil
var ErrNilPeerBlacklistHandler = errors.New("nil peer blacklist handler")

// ErrNilPeerBlacklistCacher signals that a nil peer blacklist cacher has been provided
var ErrNilPeerBlacklistCacher = errors.New("nil peer blacklist cacher")

// ErrBlacklistedConsensusPeer signals that a consensus message has been received from a blacklisted peer
var ErrBlacklistedConsensusPeer = errors.New("blacklisted consensus peer")

// ErrNilSignatureOnP2PMessage signals that a p2p message without signature was received
var ErrNilSignatureOnP2PMessage = errors.New("nil signature on the p2p message")

// ErrNilSigningHandler signals that provided signing handler is nil
var ErrNilSigningHandler = errors.New("nil signing handler")

// ErrNilKeysHandler signals that a nil keys handler was provided
var ErrNilKeysHandler = errors.New("nil keys handler")

// ErrNilFunctionHandler signals that a nil function handler was provided
var ErrNilFunctionHandler = errors.New("nil function handler")

// ErrWrongHashForHeader signals that the hash of the header is not the expected one
var ErrWrongHashForHeader = errors.New("wrong hash for header")
