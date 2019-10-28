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

// ErrBlockHashDoesNotMatch signals that header hash does not match with the previous one
var ErrBlockHashDoesNotMatch = errors.New("block hash does not match")

// ErrMissingTransaction signals that one transaction is missing
var ErrMissingTransaction = errors.New("missing transaction")

// ErrMarshalWithoutSuccess signals that marshal some data was not done with success
var ErrMarshalWithoutSuccess = errors.New("marshal without success")

// ErrUnmarshalWithoutSuccess signals that unmarshal some data was not done with success
var ErrUnmarshalWithoutSuccess = errors.New("unmarshal without success")

// ErrRootStateDoesNotMatch signals that root state does not match
var ErrRootStateDoesNotMatch = errors.New("root state does not match")

// ErrAccountStateDirty signals that the accounts were modified before starting the current modification
var ErrAccountStateDirty = errors.New("accountState was dirty before starting to change")

// ErrInvalidShardId signals that the shard id is invalid
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrMissingHeader signals that header of the block is missing
var ErrMissingHeader = errors.New("missing header")

// ErrMissingHashForHeaderNonce signals that hash of the block is missing
var ErrMissingHashForHeaderNonce = errors.New("missing hash for header nonce")

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

// ErrEmptyTxDataPool signals that a empty transaction pool has been provided
var ErrEmptyTxDataPool = errors.New("empty transaction data pool")

// ErrNilHeadersDataPool signals that a nil headers pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

// ErrNilMetaHeadersDataPool signals that a nil metachain header pool has been provided
var ErrNilMetaHeadersDataPool = errors.New("nil meta headers data pool")

// ErrNilHeadersNoncesDataPool signals that a nil header - nonce cache
var ErrNilHeadersNoncesDataPool = errors.New("nil headers nonces cache")

//ErrNilMetaHeadersNoncesDataPool signals a nil metachain header - nonce cache
var ErrNilMetaHeadersNoncesDataPool = errors.New("nil meta headers nonces cache")

// ErrNilCacher signals that a nil cache has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilUint64SyncMapCacher signals that a nil Uint64SyncMapCache has been provided
var ErrNilUint64SyncMapCacher = errors.New("nil Uint64SyncMapCacher")

// ErrNilRcvAddr signals that an operation has been attempted to or with a nil receiver address
var ErrNilRcvAddr = errors.New("nil receiver address")

// ErrNilSndAddr signals that an operation has been attempted to or with a nil sender address
var ErrNilSndAddr = errors.New("nil sender address")

// ErrNegativeValue signals that a negative value has been detected and it is not allowed
var ErrNegativeValue = errors.New("negative value")

// ErrNilShardCoordinator signals that an operation has been attempted to or with a nil shard coordinator
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilNodesCoordinator signals that an operation has been attempted to or with a nil nodes coordinator
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrInvalidRcvAddr signals that an operation has been attempted to or with an invalid receiver address
var ErrInvalidRcvAddr = errors.New("invalid receiver address")

// ErrInvalidSndAddr signals that an operation has been attempted to or with an invalid sender address
var ErrInvalidSndAddr = errors.New("invalid sender address")

// ErrNilKeyGen signals that an operation has been attempted to or with a nil single sign key generator
var ErrNilKeyGen = errors.New("nil key generator")

// ErrNilSingleSigner signals that a nil single signer is used
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrBlockProposerSignatureMissing signals that block proposer signature is missing from the block aggregated sig
var ErrBlockProposerSignatureMissing = errors.New("block proposer signature is missing")

// ErrNilMultiSigVerifier signals that a nil multi-signature verifier is used
var ErrNilMultiSigVerifier = errors.New("nil multi-signature verifier")

// ErrInvalidBlockBodyType signals that an operation has been attempted with an invalid block body type
var ErrInvalidBlockBodyType = errors.New("invalid block body type")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")

// ErrNilPoolsHolder signals that an operation has been attempted to or with a nil pools holder object
var ErrNilPoolsHolder = errors.New("nil pools holder")

// ErrNilTxStorage signals that a nil transaction storage has been provided
var ErrNilTxStorage = errors.New("nil transaction storage")

// ErrNilStorage signals that a nil storage has been provided
var ErrNilStorage = errors.New("nil storage")

// ErrNilShardedDataCacherNotifier signals that a nil sharded data cacher notifier has been provided
var ErrNilShardedDataCacherNotifier = errors.New("nil sharded data cacher notifier")

// ErrInvalidTxInPool signals an invalid transaction in the transactions pool
var ErrInvalidTxInPool = errors.New("invalid transaction in the transactions pool")

// ErrTxNotFound signals that a transaction has not found
var ErrTxNotFound = errors.New("transaction not found")

// ErrNilHeadersStorage signals that a nil header storage has been provided
var ErrNilHeadersStorage = errors.New("nil headers storage")

// ErrNilHeadersNonceHashStorage signals that a nil header nonce hash storage has been provided
var ErrNilHeadersNonceHashStorage = errors.New("nil headers nonce hash storage")

// ErrNilBlockBodyStorage signals that a nil block body storage has been provided
var ErrNilBlockBodyStorage = errors.New("nil block body storage")

// ErrNilTransactionPool signals that a nil transaction pool was used
var ErrNilTransactionPool = errors.New("nil transaction pool")

// ErrNilMiniBlockPool signals that a nil mini blocks pool was used
var ErrNilMiniBlockPool = errors.New("nil mini block pool")

// ErrNilMetaBlockPool signals that a nil meta blocks pool was used
var ErrNilMetaBlockPool = errors.New("nil meta block pool")

// ErrNilShardBlockPool signals that a nil shard blocks pool was used
var ErrNilShardBlockPool = errors.New("nil shard block pool")

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

// ErrNilInternalTransactionProducer signals that a nil system transactions producer was provided
var ErrNilInternalTransactionProducer = errors.New("nil internal transaction producere")

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

// ErrNilRandSeed signals that a nil rand seed has been provided
var ErrNilRandSeed = errors.New("provided rand seed is nil")

// ErrNilPrevRandSeed signals that a nil previous rand seed has been provided
var ErrNilPrevRandSeed = errors.New("provided previous rand seed is nil")

// ErrNilRequestHeaderHandlerByNonce signals that a nil header request handler by nonce func was provided
var ErrNilRequestHeaderHandlerByNonce = errors.New("nil request header handler by nonce")

// ErrLowerRoundInBlock signals that a header round is too low for processing it
var ErrLowerRoundInBlock = errors.New("header round is lower than last committed")

// ErrRandSeedDoesNotMatch signals that random seed does not match with the previous one
var ErrRandSeedDoesNotMatch = errors.New("random seed do not match")

// ErrHeaderNotFinal signals that header is not final and it should be
var ErrHeaderNotFinal = errors.New("header in metablock is not final")

// ErrShardIdMissmatch signals shard ID does not match expectations
var ErrShardIdMissmatch = errors.New("shard ID missmatch")

// ErrMintAddressNotInThisShard signals that the mint address does not belong to current shard
var ErrMintAddressNotInThisShard = errors.New("mint address does not belong to current shard")

// ErrNotarizedHdrsSliceIsNil signals that the slice holding last notarized headers is nil
var ErrNotarizedHdrsSliceIsNil = errors.New("notarized shard headers slice is nil")

// ErrCrossShardMBWithoutConfirmationFromMeta signals that miniblock was not yet notarized by metachain
var ErrCrossShardMBWithoutConfirmationFromMeta = errors.New("cross shard miniblock with destination current shard is not confirmed by metachain")

// ErrHeaderBodyMismatch signals that the header does not attest all data from the block
var ErrHeaderBodyMismatch = errors.New("body cannot be validated from header data")

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

// ErrWrongNonceInVMOutput signals that nonce in vm output is wrong
var ErrWrongNonceInVMOutput = errors.New("nonce invalid from SC run")

// ErrNilVMOutput signals that vmoutput is nil
var ErrNilVMOutput = errors.New("nil vm output")

// ErrNilBalanceFromSC signals that balance is nil
var ErrNilBalanceFromSC = errors.New("output balance from VM is nil")

// ErrNilValueFromRewardTransaction signals that the transfered value is nil
var ErrNilValueFromRewardTransaction = errors.New("transferred value is nil in reward transaction")

// ErrNilTemporaryAccountsHandler signals that temporary accounts handler is nil
var ErrNilTemporaryAccountsHandler = errors.New("temporary accounts handler is nil")

// ErrNotEnoughValidBlocksInStorage signals that bootstrap from storage failed due to not enough valid blocks stored
var ErrNotEnoughValidBlocksInStorage = errors.New("not enough valid blocks in storage")

// ErrNilSmartContractResult signals that the smart contract result is nil
var ErrNilSmartContractResult = errors.New("smart contract result is nil")

// ErrNilRewardTransaction signals that the reward transaction is nil
var ErrNilRewardTransaction = errors.New("reward transaction is nil")

// ErrRewardTransactionNotFound is raised when reward transaction should be present but was not found
var ErrRewardTransactionNotFound = errors.New("reward transaction not found")

// ErrInvalidDataInput signals that the data input is invalid for parsing
var ErrInvalidDataInput = errors.New("data input is invalid to create key, value storage output")

// ErrNoUnsignedTransactionInMessage signals that message does not contain required data
var ErrNoUnsignedTransactionInMessage = errors.New("no unsigned transactions in message")

// ErrNoRewardTransactionInMessage signals that message does not contain required data
var ErrNoRewardTransactionInMessage = errors.New("no reward transactions in message")

// ErrNilUTxDataPool signals that unsigned transaction pool is nil
var ErrNilUTxDataPool = errors.New("unsigned transactions pool is nil")

// ErrNilRewardTxDataPool signals that the reward transactions pool is nil
var ErrNilRewardTxDataPool = errors.New("reward transactions pool is nil")

// ErrNilUTxStorage signals that unsigned transaction storage is nil
var ErrNilUTxStorage = errors.New("unsigned transactions storage is nil")

// ErrNilRewardsTxStorage signals that rewards transaction storage is nil
var ErrNilRewardsTxStorage = errors.New("reward transactions storage is nil")

// ErrNilScAddress signals that a nil smart contract address has been provided
var ErrNilScAddress = errors.New("nil SC address")

// ErrEmptyFunctionName signals that an empty function name has been provided
var ErrEmptyFunctionName = errors.New("empty function name")

// ErrMiniBlockHashMismatch signals that miniblock hashes does not match
var ErrMiniBlockHashMismatch = errors.New("miniblocks does not match")

// ErrNilIntermediateTransactionHandler signals that nil intermediate transaction handler was provided
var ErrNilIntermediateTransactionHandler = errors.New("intermediate transaction handler is nil")

// ErrWrongTypeInMiniBlock signals that type is not correct for processing
var ErrWrongTypeInMiniBlock = errors.New("type in miniblock is not correct for processing")

// ErrNilTransactionCoordinator signals that transaction coordinator is nil
var ErrNilTransactionCoordinator = errors.New("transaction coordinator is nil")

// ErrNilUint64Converter signals that uint64converter is nil
var ErrNilUint64Converter = errors.New("unit64converter is nil")

// ErrNilSmartContractResultProcessor signals that smart contract result processor is nil
var ErrNilSmartContractResultProcessor = errors.New("nil smart contract result processor")

// ErrNilRewardsTxProcessor signals that the rewards transaction processor is nil
var ErrNilRewardsTxProcessor = errors.New("nil rewards transaction processor")

// ErrNilIntermediateProcessorContainer signals that intermediate processors container is nil
var ErrNilIntermediateProcessorContainer = errors.New("intermediate processor container is nil")

// ErrNilPreProcessorsContainer signals that preprocessors container is nil
var ErrNilPreProcessorsContainer = errors.New("preprocessors container is nil")

// ErrUnknownBlockType signals that block type is not correct
var ErrUnknownBlockType = errors.New("block type is unknown")

// ErrMissingPreProcessor signals that required pre processor is missing
var ErrMissingPreProcessor = errors.New("pre processor is missing")

// ErrNilTxHandlerValidator signals that a nil tx handler validator has been provided
var ErrNilTxHandlerValidator = errors.New("nil tx handler validator provided")

// ErrNilHeaderHandlerValidator signals that a nil header handler validator has been provided
var ErrNilHeaderHandlerValidator = errors.New("nil header handler validator provided")

// ErrNilAppStatusHandler defines the error for setting a nil AppStatusHandler
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")

// ErrNilUnsignedTxHandler signals that the unsigned tx handler is nil
var ErrNilUnsignedTxHandler = errors.New("nil unsigned tx handler")

// ErrRewardTxsDoNotMatch signals that reward txs do not match
var ErrRewardTxsDoNotMatch = errors.New("calculated reward tx with block reward tx does not match")

// ErrRewardTxNotFound signals that the reward transaction was not found
var ErrRewardTxNotFound = errors.New("reward transaction not found")

// ErrRewardTxsMismatchCreatedReceived signals a mismatch between the nb of created and received reward transactions
var ErrRewardTxsMismatchCreatedReceived = errors.New("mismatch between created and received reward transactions")

// ErrNilTxTypeHandler signals that tx type handler is nil
var ErrNilTxTypeHandler = errors.New("nil tx type handler")

// ErrNilSpecialAddressHandler signals that special address handler is nil
var ErrNilSpecialAddressHandler = errors.New("nil special address handler")

// ErrNotEnoughArgumentsToDeploy signals that there are not enough arguments to deploy the smart contract
var ErrNotEnoughArgumentsToDeploy = errors.New("not enough arguments to deploy the smart contract")

// ErrVMTypeLengthInvalid signals that vm type length is too long
var ErrVMTypeLengthInvalid = errors.New("vm type length is too long")

// ErrOverallBalanceChangeFromSC signals that all sumed balance changes are not zero
var ErrOverallBalanceChangeFromSC = errors.New("SC output balance updates are wrong")

// ErrNilTxsPoolsCleaner signals that a nil transactions pools cleaner has been provided
var ErrNilTxsPoolsCleaner = errors.New("nil transactions pools cleaner")

// ErrZeroMaxCleanTime signals that cleaning time for pools is less or equal with 0
var ErrZeroMaxCleanTime = errors.New("cleaning time is equal or less than zero")

// ErrNilEconomicsRewardsHandler signals that rewards handler is nil
var ErrNilEconomicsRewardsHandler = errors.New("nil economics rewards handler")

// ErrNilEconomicsFeeHandler signals that fee handler is nil
var ErrNilEconomicsFeeHandler = errors.New("nil economics fee handler")

// ErrNilThrottler signals that a nil throttler has been provided
var ErrNilThrottler = errors.New("nil throttler")

// ErrSystemBusy signals that the system is busy
var ErrSystemBusy = errors.New("system busy")

// ErrInsufficientGasPriceInTx signals that a lower gas price than required was provided
var ErrInsufficientGasPriceInTx = errors.New("insufficient gas price in tx")

// ErrInsufficientGasLimitInTx signals that a lower gas limit than required was provided
var ErrInsufficientGasLimitInTx = errors.New("insufficient gas limit in tx")

// ErrInvalidMinimumGasPrice signals that an invalid gas price has been read from config file
var ErrInvalidMinimumGasPrice = errors.New("invalid minimum gas price")

// ErrInvalidMinimumGasLimitForTx signals that an invalid minimum gas limit for transactions has been read from config file
var ErrInvalidMinimumGasLimitForTx = errors.New("invalid minimum gas limit for transactions")

// ErrInvalidRewardsValue signals that an invalid rewards value has been read from config file
var ErrInvalidRewardsValue = errors.New("invalid rewards value")

// ErrInvalidRewardsPercentages signals that rewards percentages are not correct
var ErrInvalidRewardsPercentages = errors.New("invalid rewards percentages")

// ErrInvalidNonceRequest signals that invalid nonce was requested
var ErrInvalidNonceRequest = errors.New("invalid nonce request")

// ErrNilBlockChainHook signals that nil blockchain hook has been provided
var ErrNilBlockChainHook = errors.New("nil blockchain hook")

// ErrNilSCDataGetter signals that a nil sc data getter has been provided
var ErrNilSCDataGetter = errors.New("nil sc data getter")

// ErrPeerChangesHashDoesNotMatch signals that peer changes from header does not match the created ones
var ErrPeerChangesHashDoesNotMatch = errors.New("peer changes hash does not match")

// ErrNilPeerAccountsAdapter signals that nil peer accounts has been provided
var ErrNilPeerAccountsAdapter = errors.New("nil peer accounts adapter")

// ErrNilTxForCurrentBlockHandler signals that nil tx for current block handler has been provided
var ErrNilTxForCurrentBlockHandler = errors.New("nil tx for current block handler")

// ErrNilSCToProtocol signals that nil smart contract to protocol handler has been provided
var ErrNilSCToProtocol = errors.New("nil sc to protocol")

// ErrNilPeerChangesHandler signals that nil peer changes handler has been provided
var ErrNilPeerChangesHandler = errors.New("nil peer changes handler")

// ErrNilSystemContractsContainer signals that the provided system contract container is nil
var ErrNilSystemContractsContainer = errors.New("system contract container is nil")
