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

// ErrNilAddressConverter signals that a nil address converter has been provided
var ErrNilAddressConverter = errors.New("trying to set nil address converter")

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

// ErrNilBalances signals that a nil list of initial balances has been provided
var ErrNilBalances = errors.New("trying to set nil balances")

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

// ErrNilBlockHeader is raised when a valid block header is expected but nil was used
var ErrNilBlockHeader = errors.New("block header is nil")

// ErrNilTxBlockBody is raised when a valid tx block body is expected but nil was used
var ErrNilTxBlockBody = errors.New("tx block body is nil")

// ErrWrongTypeAssertion is raised when a type assertion occurs
var ErrWrongTypeAssertion = errors.New("wrong type assertion: expected *block.Header")

// ErrNegativeDurationInSecToConsiderUnresponsive is raised when a value less than 1 has been provided
var ErrNegativeDurationInSecToConsiderUnresponsive = errors.New("value DurationInSecToConsiderUnresponsive is less" +
	" than 1")

// ErrNegativeMaxTimeToWaitBetweenBroadcastsInSec is raised when a value less than 1 has been provided
var ErrNegativeMaxTimeToWaitBetweenBroadcastsInSec = errors.New("value MaxTimeToWaitBetweenBroadcastsInSec is less " +
	"than 1")

// ErrNegativeMinTimeToWaitBetweenBroadcastsInSec is raised when a value less than 1 has been provided
var ErrNegativeMinTimeToWaitBetweenBroadcastsInSec = errors.New("value MinTimeToWaitBetweenBroadcastsInSec is less " +
	"than 1")

// ErrWrongValues signals that wrong values were provided
var ErrWrongValues = errors.New("wrong values for heartbeat parameters")

// ErrGenesisBlockNotInitialized signals that genesis block is not initialized
var ErrGenesisBlockNotInitialized = errors.New("genesis block is not initialized")

// ErrNilBlackListHandler signals that a nil black list handler was provided
var ErrNilBlackListHandler = errors.New("nil black list handler")

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

// ErrNilBootStorer signals that a nil boot storer was provided
var ErrNilBootStorer = errors.New("nil boot storer")

// ErrNilHeaderSigVerifier signals that a nil header sig verifier has been provided
var ErrNilHeaderSigVerifier = errors.New("nil header sig verifier")

// ErrNilValidatorStatistics signals that a nil validator statistics has been provided
var ErrNilValidatorStatistics = errors.New("nil validator statistics")

// ErrCannotConvertToPeerAccount signals that the given account cannot be converted to a peer account
var ErrCannotConvertToPeerAccount = errors.New("cannot convert to peer account")

// ErrInvalidChainID signals that an invalid chain ID has been provided
var ErrInvalidChainID = errors.New("invalid chain ID in Node")
