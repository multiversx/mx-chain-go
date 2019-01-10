package process

import (
	"errors"
)

// ErrNilAccountsAdapter defines the error when trying to use a nil AccountsAddapter
var ErrNilAccountsAdapter = errors.New("nil AccountsAdapter")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil Hasher")

// ErrNilAddressConverter signals that an operation has been attempted to or with a nil AddressConverter implementation
var ErrNilAddressConverter = errors.New("nil AddressConverter")

// ErrNilAddressContainer signals that an operation has been attempted to or with a nil AddressContainer implementation
var ErrNilAddressContainer = errors.New("nil AddressContainer")

// ErrNilTransaction signals that an operation has been attempted to or with a nil transaction
var ErrNilTransaction = errors.New("nil transaction")

// ErrNoVM signals that no SCHandler has been set
var ErrNoVM = errors.New("no VM (hook not set)")

// ErrHigherNonceInTransaction signals the nonce in transaction is higher than the account's nonce
var ErrHigherNonceInTransaction = errors.New("higher nonce in transaction")

// ErrLowerNonceInTransaction signals the nonce in transaction is lower than the account's nonce
var ErrLowerNonceInTransaction = errors.New("lower nonce in transaction")

// ErrInsufficientFunds signals the funds are insufficient
var ErrInsufficientFunds = errors.New("insufficient funds")

// ErrNilValue signals the value is nil
var ErrNilValue = errors.New("nil value")

// ErrNilBlockChain signals that an operation has been attempted to or with a nil blockchain
var ErrNilBlockChain = errors.New("nil block chain")

// ErrNilTxBlockBody signals that an operation has been attempted to or with a nil block body
var ErrNilTxBlockBody = errors.New("nil block body")

// ErrNilStateBlockBody signals that an operation has been attempted to or with a nil block body
var ErrNilStateBlockBody = errors.New("nil block body")

// ErrNilPeerBlockBody signals that an operation has been attempted to or with a nil block body
var ErrNilPeerBlockBody = errors.New("nil block body")

// ErrNilBlockHeader signals that an operation has been attempted to or with a nil block header
var ErrNilBlockHeader = errors.New("nil block header")

// ErrNilBlockBodyHash signals that an operation has been attempted to or with a nil block body hash
var ErrNilBlockBodyHash = errors.New("nil block body hash")

// ErrNilTxHash signals that an operation has been attempted with a nil hash
var ErrNilTxHash = errors.New("nil transaction hash")

// ErrNilPeerChanges signals that an operation has been attempted with nil peer changes
var ErrNilPeerChanges = errors.New("nil peer block changes")

// ErrNilPublicKey signals that a operation has been attempted with a nil public key
var ErrNilPublicKey = errors.New("nil public key")

// ErrNilPubKeysBitmap signals that a operation has been attempted with a nil public keys bitmap
var ErrNilPubKeysBitmap = errors.New("nil public keys bitmap")

// ErrNilPreviousBlockHash signals that a operation has been attempted with a nil previous block header hash
var ErrNilPreviousBlockHash = errors.New("nil previous block header hash")

// ErrNilSignature signals that a operation has been attempted with a nil signature
var ErrNilSignature = errors.New("nil signature")

// ErrNilChallenge signals that a operation has been attempted with a nil challenge
var ErrNilChallenge = errors.New("nil challenge")

// ErrNilCommitment signals that a operation has been attempted with a nil commitment
var ErrNilCommitment = errors.New("nil commitment")

// ErrNilMiniBlocks signals that an operation has been attempted with a nil mini-block
var ErrNilMiniBlocks = errors.New("nil mini blocks")

// ErrNilTxHashes signals that an operation has been atempted with snil transaction hashes
var ErrNilTxHashes = errors.New("nil transaction hashes")

// ErrNilRootHash signals that an operation has been attempted with a nil root hash
var ErrNilRootHash = errors.New("root hash is nil")

// ErrWrongNonceInBlock signals the nonce in block is different than expected nounce
var ErrWrongNonceInBlock = errors.New("wrong nonce in block")

// ErrInvalidBlockHash signals the hash of the block is not matching with the previous one
var ErrInvalidBlockHash = errors.New("invalid block hash")

// ErrInvalidBlockSignature signals the signature of the block is not valid
var ErrInvalidBlockSignature = errors.New("invalid block signature")

// ErrMissingTransaction signals that one transaction is missing
var ErrMissingTransaction = errors.New("missing transaction")

// ErrMarshalWithoutSuccess signals that marshal some data was not done with success
var ErrMarshalWithoutSuccess = errors.New("marshal without success")

// ErrPersistWithoutSuccess signals that persist some data was not done with success
var ErrPersistWithoutSuccess = errors.New("persist without success")

// ErrRootStateMissmatch signals that persist some data was not done with success
var ErrRootStateMissmatch = errors.New("root state does not match")

// ErrAccountStateDirty signals that the accounts were modified before starting the current modification
var ErrAccountStateDirty = errors.New("accountState was dirty before starting to change")

// ErrInvalidShardId signals that the shard id is invalid
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrMissingHeader signals that header of the block is missing
var ErrMissingHeader = errors.New("missing header")

// ErrMissingBody signals that body of the block is missing
var ErrMissingBody = errors.New("missing body")

// ErrNilBlockExecutor signals that an operation has been attempted to or with a nil BlockExecutor implementation
var ErrNilBlockExecutor = errors.New("nil BlockExecutor")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNilRound signals that an operation has been attempted to or with a nil Round
var ErrNilRound = errors.New("nil Round")

// ErrNilMessenger signals that a nil Messenger object was provided
var ErrNilMessenger = errors.New("nil Messenger")

// ErrNilNewer signals that a nil Newer object was provided
var ErrNilNewer = errors.New("nil Newer")

// ErrRegisteringValidator signals that a registration validator occur
var ErrRegisteringValidator = errors.New("error while registering validator")

// ErrNilInterceptor signals that a nil Interceptor has been provided
var ErrNilInterceptor = errors.New("nil Interceptor")

// ErrNilTxDataPool signals that a nil transaction pool has been provided
var ErrNilTxDataPool = errors.New("nil transaction data pool")

// ErrNilHeadersDataPool signals that a nil header pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

// ErrNilHeadersNoncesDataPool signals that a nil header - nonce cache
var ErrNilHeadersNoncesDataPool = errors.New("nil headers nonces cache")

// ErrNilCacher signals that a nil cache has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilRcvAddr signals that an operation has been attempted to or with a nil receiver address
var ErrNilRcvAddr = errors.New("nil receiver address")

// ErrNilSndAddr signals that an operation has been attempted to or with a nil sender address
var ErrNilSndAddr = errors.New("nil sender address")

// ErrNegativeValue signals that a negative value has been detected and it is not allowed
var ErrNegativeValue = errors.New("negative value")

// ErrNilShardCoordinator signals that an operation has been attempted to or with a nil shard coordinator
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrInvalidRcvAddr signals that an operation has been attempted to or with an invalid receiver address
var ErrInvalidRcvAddr = errors.New("invalid receiver address")

// ErrInvalidSndAddr signals that an operation has been attempted to or with an invalid sender address
var ErrInvalidSndAddr = errors.New("invalid sender address")

// ErrNilSingleSignKeyGen signals that an operation has been attempted to or with a nil single sign key generator
var ErrNilSingleSignKeyGen = errors.New("nil single sign key generator")

// ErrInvalidBlockBodyType signals that an operation has been attempted with an invalid block body type
var ErrInvalidBlockBodyType = errors.New("invalid block body type")

// ErrNilTransientDataHolder signals that an operation has been attempted to or with a nil transient data holder
var ErrNilTransientDataHolder = errors.New("nil transient data holder")

// ErrNotImplementedBlockProcessingType signals that a not supported block body type was found in header
var ErrNotImplementedBlockProcessingType = errors.New("not implemented block processing type")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")

// ErrBadInterceptorTopicImplementation signals that a bad interceptor-topic implementation occurred
var ErrBadInterceptorTopicImplementation = errors.New("bad interceptor-topic implementation")

// ErrNilBlockBody signals that a nil block body has been provided
var ErrNilBlockBody = errors.New("nil block body")

// ErrNilTransientPool signals that an operation has been attempted to or with a nil transient pool of data
var ErrNilTransientPool = errors.New("nil transient pool")

// ErrNilTxStorage signals that a nil transaction storage has been provided
var ErrNilTxStorage = errors.New("nil transaction storage")

// ErrNilHeadersStorage signals that a nil header storage has been provided
var ErrNilHeadersStorage = errors.New("nil headers storage")

// ErrNilTopic signals that a nil topic has been provided/fetched
var ErrNilTopic = errors.New("nil topic")

// ErrResolveRequestAlreadyAssigned signals that ResolveRequest is not nil for a particular topic
var ErrResolveRequestAlreadyAssigned = errors.New("resolve request func has already been assigned for this topic")

// ErrTopicNotWiredToMessenger signals that a call to a not-correctly-instantiated topic has been made
var ErrTopicNotWiredToMessenger = errors.New("topic has not been wired to a p2p.Messenger implementation")

// ErrNilResolver signals that a nil resolver object has been provided
var ErrNilResolver = errors.New("nil resolver")

// ErrNilNonceConverter signals that a nil nonce converter has been provided
var ErrNilNonceConverter = errors.New("nil nonce converter")

// ErrInvalidNonceByteSlice signals that an invalid byte slice has been provided
// and an uint64 can not be decoded from that byte slice
var ErrInvalidNonceByteSlice = errors.New("invalid nonce byte slice")

// ErrResolveNotHashType signals that an expected resolve type was other than hash type
var ErrResolveNotHashType = errors.New("expected resolve type was hash type")

// ErrResolveTypeUnknown signals that an unknown resolve type was provided
var ErrResolveTypeUnknown = errors.New("unknown resolve type")

// ErrNilBlockBodyPool signals that a nil block body pool has been provided
var ErrNilBlockBodyPool = errors.New("nil block body pool")

// ErrNilBlockBodyStorage signals that a nil block body storage has been provided
var ErrNilBlockBodyStorage = errors.New("nil block body storage")
