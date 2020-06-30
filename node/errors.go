package node

import (
	"errors"
)

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("trying to set nil marshalizer")

// ErrNilMessenger signals that a nil messenger has been provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("trying to set nil hasher")

// ErrNilAccountsAdapter signals that a nil accounts adapter has been provided
var ErrNilAccountsAdapter = errors.New("trying to set nil accounts adapter")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to use a nil pubkey converter")

// ErrNilBlockchain signals that a nil blockchain structure has been provided
var ErrNilBlockchain = errors.New("nil blockchain")

// ErrNilStore signals that a nil store has been provided
var ErrNilStore = errors.New("nil data store")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("trying to set nil private key")

// ErrNilSingleSignKeyGen signals that a nil single key generator has been provided
var ErrNilSingleSignKeyGen = errors.New("trying to set nil single sign key generator")

// ErrNilKeyGenForBalances signals that a nil keygen for balances has been provided
var ErrNilKeyGenForBalances = errors.New("trying to set a nil key gen for signing")

// ErrNilTxFeeHandler signals that a nil tx fee handler was provided
var ErrNilTxFeeHandler = errors.New("trying to set a nil tx fee handler")

// ErrNilPublicKey signals that a nil public key has been provided
var ErrNilPublicKey = errors.New("trying to set nil public key")

// ErrZeroRoundDurationNotSupported signals that 0 seconds round duration is not supported
var ErrZeroRoundDurationNotSupported = errors.New("0 round duration time is not supported")

// ErrNegativeOrZeroConsensusGroupSize signals that 0 elements consensus group is not supported
var ErrNegativeOrZeroConsensusGroupSize = errors.New("group size should be a strict positive number")

// ErrNilSyncTimer signals that a nil sync timer has been provided
var ErrNilSyncTimer = errors.New("trying to set nil sync timer")

// ErrNilRounder signals that a nil rounder has been provided
var ErrNilRounder = errors.New("trying to set nil rounder")

// ErrNilBlockProcessor signals that a nil block processor has been provided
var ErrNilBlockProcessor = errors.New("trying to set nil block processor")

// ErrNilDataPool signals that a nil data pool has been provided
var ErrNilDataPool = errors.New("trying to set nil data pool")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

// ErrNilNodesCoordinator signals that a nil nodes coordinator has been provided
var ErrNilNodesCoordinator = errors.New("trying to set nil nodes coordinator")

// ErrNilUint64ByteSliceConverter signals that a nil uint64 <-> byte slice converter has been provided
var ErrNilUint64ByteSliceConverter = errors.New("trying to set nil uint64 - byte slice converter")

// ErrNilSingleSig signals that a nil singlesig object has been provided
var ErrNilSingleSig = errors.New("trying to set nil singlesig")

// ErrNilMultiSig signals that a nil multiSigner object has been provided
var ErrNilMultiSig = errors.New("trying to set nil multiSigner")

// ErrNilForkDetector signals that a nil forkdetector object has been provided
var ErrNilForkDetector = errors.New("nil fork detector")

// ErrValidatorAlreadySet signals that a topic validator has already been set
var ErrValidatorAlreadySet = errors.New("topic validator has already been set")

// ErrNilInterceptorsContainer signals that a nil interceptors container has been provided
var ErrNilInterceptorsContainer = errors.New("nil interceptors container")

// ErrNilResolversFinder signals that a nil resolvers finder has been provided
var ErrNilResolversFinder = errors.New("nil resolvers finder")

// ErrNilEpochStartTrigger signals that a nil start of epoch trigger has been provided
var ErrNilEpochStartTrigger = errors.New("nil start of epoch trigger")

// ErrGenesisBlockNotInitialized signals that genesis block is not initialized
var ErrGenesisBlockNotInitialized = errors.New("genesis block is not initialized")

// ErrNilPeerDenialEvaluator signals that a nil peer denial evaluator was provided
var ErrNilPeerDenialEvaluator = errors.New("nil peer denial evaluator")

// ErrNilTimeCache signals that a nil time cache was provided
var ErrNilTimeCache = errors.New("nil time cache")

// ErrNilRequestedItemsHandler signals that a nil requested items handler was provided
var ErrNilRequestedItemsHandler = errors.New("nil requested items handler")

// ErrSystemBusyGeneratingTransactions signals that to many transactions are trying to get generated
var ErrSystemBusyGeneratingTransactions = errors.New("system busy while generating bulk transactions")

// ErrNilStatusHandler is returned when the status handler is nil
var ErrNilStatusHandler = errors.New("nil AppStatusHandler")

// ErrNoTxToProcess signals that no transaction were sent for processing
var ErrNoTxToProcess = errors.New("no transaction to process")

// ErrInvalidValue signals that an invalid value has been provided such as NaN to an integer field
var ErrInvalidValue = errors.New("invalid value")

// ErrNilNetworkShardingCollector defines the error for setting a nil network sharding collector
var ErrNilNetworkShardingCollector = errors.New("nil network sharding collector")

// ErrNilBootStorer signals that a nil boot storer was provided
var ErrNilBootStorer = errors.New("nil boot storer")

// ErrNilHeaderSigVerifier signals that a nil header sig verifier has been provided
var ErrNilHeaderSigVerifier = errors.New("nil header sig verifier")

// ErrNilHeaderIntegrityVerifier signals that a nil header integrity verifier has been provided
var ErrNilHeaderIntegrityVerifier = errors.New("nil header integrity verifier")

// ErrNilValidatorStatistics signals that a nil validator statistics has been provided
var ErrNilValidatorStatistics = errors.New("nil validator statistics")

// ErrInvalidChainID signals that an invalid chain ID has been provided
var ErrInvalidChainID = errors.New("invalid chain ID in Node")

// ErrNilBlockTracker signals that a nil block tracker has been provided
var ErrNilBlockTracker = errors.New("trying to set nil block tracker")

// ErrNilPendingMiniBlocksHandler signals that a nil pending miniblocks handler has been provided
var ErrNilPendingMiniBlocksHandler = errors.New("trying to set nil pending miniblocks handler")

// ErrNilRequestHandler signals that a nil request handler has been provided
var ErrNilRequestHandler = errors.New("trying to set nil request handler")

// ErrNilAntifloodHandler signals that a nil antiflood handler has been provided
var ErrNilAntifloodHandler = errors.New("nil antiflood handler")

// ErrNilTxAccumulator signals that a nil Accumulator instance has been provided
var ErrNilTxAccumulator = errors.New("nil tx accumulator")

// ErrNilHardforkTrigger signals that a nil hardfork trigger has been provided
var ErrNilHardforkTrigger = errors.New("nil hardfork trigger")

// ErrNilWhiteListHandler signals that white list handler is nil
var ErrNilWhiteListHandler = errors.New("nil whitelist handler")

// ErrNilNodeStopChannel signals that a nil channel for node process stop has been provided
var ErrNilNodeStopChannel = errors.New("nil node stop channel")

// ErrNilQueryHandler signals that a nil query handler has been provided
var ErrNilQueryHandler = errors.New("nil query handler")

// ErrQueryHandlerAlreadyExists signals that the query handler is already registered
var ErrQueryHandlerAlreadyExists = errors.New("query handler already exists")

// ErrEmptyQueryHandlerName signals that an empty string can not be used to be used in the query handler container
var ErrEmptyQueryHandlerName = errors.New("empty query handler name")

// ErrNilApiTransactionByHashThrottler signals that a nil API transaction by hash throttler has been provided
var ErrNilApiTransactionByHashThrottler = errors.New("nil api transaction by hash throttler")

// ErrSystemBusyTxHash signals that too many requests occur in the same time on the transaction by hash provider
var ErrSystemBusyTxHash = errors.New("system busy. try again later")

// ErrUnknownPeerID signals that the provided peer is unknown by the current node
var ErrUnknownPeerID = errors.New("unknown peer ID")

// ErrNilPeerHonestyHandler signals that a nil peer honesty handler has been provided
var ErrNilPeerHonestyHandler = errors.New("nil peer honesty handler")

// ErrNilWatchdog signals that a nil watchdog has been provided
var ErrNilWatchdog = errors.New("nil watchdog")

// ErrInvalidTransactionVersion signals  that an invalid transaction version has been provided
var ErrInvalidTransactionVersion = errors.New("invalid transaction version")
