package process

import (
	"errors"
)

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

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

// ErrWrongTransaction signals that transaction is invalid
var ErrWrongTransaction = errors.New("invalid transaction")

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

// ErrNilMetaBlockHeader signals that an operation has been attempted to or with a nil metablock
var ErrNilMetaBlockHeader = errors.New("nil metablock header")

// ErrNilTxBlockBody signals that an operation has been attempted to or with a nil tx block body
var ErrNilTxBlockBody = errors.New("nil tx block body")

// ErrNilStore signals that the provided storage service is nil
var ErrNilStore = errors.New("nil data storage service")

// ErrNilPeerBlockBody signals that an operation has been attempted to or with a nil block body
var ErrNilPeerBlockBody = errors.New("nil block body")

// ErrNilBlockHeader signals that an operation has been attempted to or with a nil block header
var ErrNilBlockHeader = errors.New("nil block header")

// ErrNilBlockBody signals that an operation has been attempted to or with a nil block body
var ErrNilBlockBody = errors.New("nil block body")

// ErrNilTxHash signals that an operation has been attempted with a nil hash
var ErrNilTxHash = errors.New("nil transaction hash")

// ErrNilPublicKey signals that a operation has been attempted with a nil public key
var ErrNilPublicKey = errors.New("nil public key")

// ErrNilPubKeysBitmap signals that a operation has been attempted with a nil public keys bitmap
var ErrNilPubKeysBitmap = errors.New("nil public keys bitmap")

// ErrNilPreviousBlockHash signals that a operation has been attempted with a nil previous block header hash
var ErrNilPreviousBlockHash = errors.New("nil previous block header hash")

// ErrNilSignature signals that a operation has been attempted with a nil signature
var ErrNilSignature = errors.New("nil signature")

// ErrNilMiniBlocks signals that an operation has been attempted with a nil mini-block
var ErrNilMiniBlocks = errors.New("nil mini blocks")

// ErrNilMiniBlockHeaders signals that an operation has been attempted with a nil mini-block
var ErrNilMiniBlockHeaders = errors.New("nil mini block headers")

// ErrNilTxHashes signals that an operation has been atempted with nil transaction hashes
var ErrNilTxHashes = errors.New("nil transaction hashes")

// ErrNilRootHash signals that an operation has been attempted with a nil root hash
var ErrNilRootHash = errors.New("root hash is nil")

// ErrWrongNonceInBlock signals the nonce in block is different than expected nonce
var ErrWrongNonceInBlock = errors.New("wrong nonce in block")

// ErrInvalidBlockHash signals the hash of the block is not matching with the previous one
var ErrInvalidBlockHash = errors.New("invalid block hash")

// ErrMissingTransaction signals that one transaction is missing
var ErrMissingTransaction = errors.New("missing transaction")

// ErrMarshalWithoutSuccess signals that marshal some data was not done with success
var ErrMarshalWithoutSuccess = errors.New("marshal without success")

// ErrUnmarshalWithoutSuccess signals that unmarshal some data was not done with success
var ErrUnmarshalWithoutSuccess = errors.New("unmarshal without success")

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

// ErrNilRounder signals that an operation has been attempted to or with a nil Rounder implementation
var ErrNilRounder = errors.New("nil Rounder")

// ErrNilMessenger signals that a nil Messenger object was provided
var ErrNilMessenger = errors.New("nil Messenger")

// ErrNilTxDataPool signals that a nil transaction pool has been provided
var ErrNilTxDataPool = errors.New("nil transaction data pool")

// ErrNilHeadersDataPool signals that a nil header pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

// ErrNilMetachainHeadersDataPool signals that a nil metachain header pool has been provided
var ErrNilMetachainHeadersDataPool = errors.New("nil metachain headers data pool")

// ErrNilHeadersNoncesDataPool signals that a nil header - nonce cache
var ErrNilHeadersNoncesDataPool = errors.New("nil headers nonces cache")

//ErrNilMetachainHeadersNoncesDataPool signals a nil metachain header - nonce cache
var ErrNilMetachainHeadersNoncesDataPool = errors.New("nil metachain headers nonces cache")

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

// ErrNilKeyGen signals that an operation has been attempted to or with a nil single sign key generator
var ErrNilKeyGen = errors.New("nil key generator")

// ErrNilSingleSigner signals that a nil single signer is used
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrNilMultiSigVerifier signals that a nil multi-signature verifier is used
var ErrNilMultiSigVerifier = errors.New("nil multi-signature verifier")

// ErrInvalidBlockBodyType signals that an operation has been attempted with an invalid block body type
var ErrInvalidBlockBodyType = errors.New("invalid block body type")

// ErrNotImplementedBlockProcessingType signals that a not supported block body type was found in header
var ErrNotImplementedBlockProcessingType = errors.New("not implemented block processing type")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")

// ErrNilPoolsHolder signals that an operation has been attempted to or with a nil pools holder object
var ErrNilPoolsHolder = errors.New("nil pools holder")

// ErrNilTxStorage signals that a nil transaction storage has been provided
var ErrNilTxStorage = errors.New("nil transaction storage")

// ErrNilStorage signals that a nil storage has been provided
var ErrNilStorage = errors.New("nil storage")

// ErrNilBlocksTracker signals that a nil blocks tracker has been provided
var ErrNilBlocksTracker = errors.New("nil blocks tracker")

// ErrInvalidTxInPool signals an invalid transaction in the transactions pool
var ErrInvalidTxInPool = errors.New("invalid transaction in the transactions pool")

// ErrNilHeadersStorage signals that a nil header storage has been provided
var ErrNilHeadersStorage = errors.New("nil headers storage")

// ErrNilMetachainHeadersStorage signals that a nil metachain header storage has been provided
var ErrNilMetachainHeadersStorage = errors.New("nil metachain headers storage")

// ErrNilResolverSender signals that a nil resolver sender object has been provided
var ErrNilResolverSender = errors.New("nil resolver sender")

// ErrNilBlockBodyStorage signals that a nil block body storage has been provided
var ErrNilBlockBodyStorage = errors.New("nil block body storage")

// ErrNilTransactionPool signals that a nil transaction pool was used
var ErrNilTransactionPool = errors.New("nil transaction pool")

// ErrNilMiniBlockPool signals that a nil mini blocks pool was used
var ErrNilMiniBlockPool = errors.New("nil mini block pool")

// ErrNilMetaBlockPool signals that a nil meta blocks pool was used
var ErrNilMetaBlockPool = errors.New("nil meta block pool")

// ErrNilTxProcessor signals that a nil transactions processor was used
var ErrNilTxProcessor = errors.New("nil transactions processor")

// ErrNilDataPoolHolder signals that the data pool holder is nil
var ErrNilDataPoolHolder = errors.New("nil data pool holder")

// ErrTimeIsOut signals that time is out
var ErrTimeIsOut = errors.New("time is out")

// ErrNilForkDetector signals that the fork detector is nil
var ErrNilForkDetector = errors.New("nil fork detector")

// ErrNilContainerElement signals when trying to add a nil element in the container
var ErrNilContainerElement = errors.New("element cannot be nil")

// ErrInvalidContainerKey signals that an element does not exist in the container's map
var ErrInvalidContainerKey = errors.New("element does not exist in container")

// ErrContainerKeyAlreadyExists signals that an element was already set in the container's map
var ErrContainerKeyAlreadyExists = errors.New("provided key already exists in container")

// ErrNilResolverContainer signals that a nil resolver container was provided
var ErrNilResolverContainer = errors.New("nil resolver container")

// ErrNilRequestHandler signals that a nil request handler interface was provided
var ErrNilRequestHandler = errors.New("nil request handler")

// ErrNilHaveTimeHandler signals that a nil have time handler func was provided
var ErrNilHaveTimeHandler = errors.New("nil have time handler")

// ErrCouldNotDecodeUnderlyingBody signals that an InterceptedBlockBody could not be decoded to a block.Body using type assertion
var ErrCouldNotDecodeUnderlyingBody = errors.New("could not decode InterceptedBlockBody to block.Body")

// ErrWrongTypeInContainer signals that a wrong type of object was found in container
var ErrWrongTypeInContainer = errors.New("wrong type of object inside container")

// ErrLenMismatch signals that 2 or more slices have different lengths
var ErrLenMismatch = errors.New("lengths mismatch")

// ErrWrongTypeAssertion signals that an type assertion failed
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrRollbackFromGenesis signals that a rollback from genesis is called
var ErrRollbackFromGenesis = errors.New("roll back from genesis is not supported")

// ErrNoTransactionInMessage signals that no transaction was found after parsing received p2p message
var ErrNoTransactionInMessage = errors.New("no transaction found in received message")

// ErrNilBuffer signals that a provided byte buffer is nil
var ErrNilBuffer = errors.New("provided byte buffer is nil")

// ErrNilChronologyValidator signals that a nil chronology validator has been provided
var ErrNilChronologyValidator = errors.New("provided chronology validator object is nil")

// ErrNilRandSeed signals that a nil rand seed has been provided
var ErrNilRandSeed = errors.New("provided rand seed is nil")

// ErrNilPrevRandSeed signals that a nil previous rand seed has been provided
var ErrNilPrevRandSeed = errors.New("provided previous rand seed is nil")

// ErrNilRequestHeaderHandlerByNonce signals that a nil header request handler by nonce func was provided
var ErrNilRequestHeaderHandlerByNonce = errors.New("nil request header handler by nonce")

// ErrLowShardHeaderRound signals that shard header round is too low for processing
var ErrLowShardHeaderRound = errors.New("shard header round is lower than last committed for this shard")

// ErrRandSeedMismatch signals that random seeds are not equal
var ErrRandSeedMismatch = errors.New("random seeds do not match")

// ErrHeaderNotFinal signals that header is not final and it should be
var ErrHeaderNotFinal = errors.New("header in metablock is not final")

// ErrShardIdMissmatch signals shard ID does not match expectations
var ErrShardIdMissmatch = errors.New("shard ID missmatch")

// ErrMintAddressNotInThisShard signals that the mint address does not belong to current shard
var ErrMintAddressNotInThisShard = errors.New("mint address does not belong to current shard")

// ErrNotarizedHdrsSliceIsNil signals that the slice holding last notarized headers is nil
var ErrNotarizedHdrsSliceIsNil = errors.New("notarized shard headers slice is nil")

// ErrNoNewMetablocks signals that no new metablocks are in the pool
var ErrNoNewMetablocks = errors.New("there is no new metablocks")

// ErrNoSortedHdrsForShard signals that there are no sorted hdrs in pool
var ErrNoSortedHdrsForShard = errors.New("no sorted headers in pool")

// ErrCrossShardMBWithoutConfirmationFromMeta signals that miniblock was not yet notarized by metachain
var ErrCrossShardMBWithoutConfirmationFromMeta = errors.New("cross shard miniblock with destination current shard is not confirmed by metachain")

// ErrHeaderBodyMismatch signals that the header does not attest all data from the block
var ErrHeaderBodyMismatch = errors.New("body cannot be validated from header data")

// ErrMetaBlockNotFinal signals that metablock is not final
var ErrMetaBlockNotFinal = errors.New("cannot attest meta blocks finality")

// ErrNilSmartContractProcessor signals that smart contract call executor is nil
var ErrNilSmartContractProcessor = errors.New("smart contract processor is nil")

// ErrNilArguments signals that arguments from transactions data is nil
var ErrNilArguments = errors.New("smart contract arguments are nil")

// ErrNilCode signals that code from transaction data is nil
var ErrNilCode = errors.New("smart contract code is nil")

// ErrNilFunction signals that function from transaction data is nil
var ErrNilFunction = errors.New("smart contract function is nil")

// ErrStringSplitFailed signals that data splitting into arguments and code failed
var ErrStringSplitFailed = errors.New("data splitting into arguments and code/function failed")

// ErrNilArgumentParser signals that the argument parser is nil
var ErrNilArgumentParser = errors.New("argument parser is nil")

// ErrNilSCDestAccount signals that destination account is nil
var ErrNilSCDestAccount = errors.New("nil destination SC account")
