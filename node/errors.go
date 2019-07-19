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

// ErrNilBlockTracker signals that a nil block tracker has been provided
var ErrNilBlockTracker = errors.New("trying to set nil block tracker")

// ErrNilDataPool signals that a nil data pool has been provided
var ErrNilDataPool = errors.New("trying to set nil data pool")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

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

// ErrNilBlockSizeThrottler signals that a nil blockSizeThrottler object has been provided
var ErrNilBlockSizeThrottler = errors.New("nil block size throttler")

// ErrValidatorAlreadySet signals that a topic validator has already been set
var ErrValidatorAlreadySet = errors.New("topic validator has already been set")

// ErrNilInterceptorsContainer signals that a nil interceptors container has been provided
var ErrNilInterceptorsContainer = errors.New("nil interceptors container")

// ErrNilResolversFinder signals that a nil resolvers finder has been provided
var ErrNilResolversFinder = errors.New("nil resolvers finder")

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

// ErrNilTransactionPool signals that a nil transaction pool was used
var ErrNilTransactionPool = errors.New("nil transaction pool")

// ErrTooManyTransactionsInPool signals that are too many transactions in pool
var ErrTooManyTransactionsInPool = errors.New("too many transactions in pool")

// ErrSystemBusyGeneratingTransactions signals that to many transactions are trying to get generated
var ErrSystemBusyGeneratingTransactions = errors.New("system busy while generating bulk transactions")
