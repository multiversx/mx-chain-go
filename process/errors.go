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

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilGasSchedule signals that an operation has been attempted with a nil gas schedule
var ErrNilGasSchedule = errors.New("nil GasSchedule")

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

// ErrInsufficientFunds signals the funds are insufficient for the move balance operation but the
// transaction fee is covered by the current balance
var ErrInsufficientFunds = errors.New("insufficient funds")

// ErrInsufficientFee signals that the current balance doesn't have the required transaction fee
var ErrInsufficientFee = errors.New("insufficient balance for fees")

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

// ErrNilBootStorer signals that the provided boot storer is bil
var ErrNilBootStorer = errors.New("nil boot storer")

// ErrNilBlockHeader signals that an operation has been attempted to or with a nil block header
var ErrNilBlockHeader = errors.New("nil block header")

// ErrNilBlockBody signals that an operation has been attempted to or with a nil block body
var ErrNilBlockBody = errors.New("nil block body")

// ErrNilTxHash signals that an operation has been attempted with a nil hash
var ErrNilTxHash = errors.New("nil transaction hash")

// ErrNilPubKeysBitmap signals that a operation has been attempted with a nil public keys bitmap
var ErrNilPubKeysBitmap = errors.New("nil public keys bitmap")

// ErrNilPreviousBlockHash signals that a operation has been attempted with a nil previous block header hash
var ErrNilPreviousBlockHash = errors.New("nil previous block header hash")

// ErrNilSignature signals that a operation has been attempted with a nil signature
var ErrNilSignature = errors.New("nil signature")

// ErrNilMiniBlocks signals that an operation has been attempted with a nil mini-block
var ErrNilMiniBlocks = errors.New("nil mini blocks")

// ErrNilMiniBlock signals that an operation has been attempted with a nil miniblock
var ErrNilMiniBlock = errors.New("nil mini block")

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

// ErrValidatorStatsRootHashDoesNotMatch signals that the root hash for the validator statistics does not match
var ErrValidatorStatsRootHashDoesNotMatch = errors.New("root hash for validator statistics does not match")

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

// ErrNilBlockProcessor signals that an operation has been attempted to or with a nil BlockProcessor implementation
var ErrNilBlockProcessor = errors.New("nil block processor")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNilNodesConfigProvider signals that an operation has been attempted to or with a nil nodes config provider
var ErrNilNodesConfigProvider = errors.New("nil nodes config provider")

// ErrNilSystemSCConfig signals that nil system sc config was provided
var ErrNilSystemSCConfig = errors.New("nil system sc config")

// ErrNilRounder signals that an operation has been attempted to or with a nil Rounder implementation
var ErrNilRounder = errors.New("nil Rounder")

// ErrNilMessenger signals that a nil Messenger object was provided
var ErrNilMessenger = errors.New("nil Messenger")

// ErrNilTxDataPool signals that a nil transaction pool has been provided
var ErrNilTxDataPool = errors.New("nil transaction data pool")

// ErrNilHeadersDataPool signals that a nil headers pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

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

// ErrNilNodesCoordinator signals that an operation has been attempted to or with a nil nodes coordinator
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrNilKeyGen signals that an operation has been attempted to or with a nil single sign key generator
var ErrNilKeyGen = errors.New("nil key generator")

// ErrNilSingleSigner signals that a nil single signer is used
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrBlockProposerSignatureMissing signals that block proposer signature is missing from the block aggregated sig
var ErrBlockProposerSignatureMissing = errors.New("block proposer signature is missing")

// ErrNilMultiSigVerifier signals that a nil multi-signature verifier is used
var ErrNilMultiSigVerifier = errors.New("nil multi-signature verifier")

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

// ErrNilTransactionPool signals that a nil transaction pool was used
var ErrNilTransactionPool = errors.New("nil transaction pool")

// ErrNilMiniBlockPool signals that a nil mini blocks pool was used
var ErrNilMiniBlockPool = errors.New("nil mini block pool")

// ErrNilMetaBlocksPool signals that a nil meta blocks pool was used
var ErrNilMetaBlocksPool = errors.New("nil meta blocks pool")

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

// ErrNilArgumentStruct signals that a function has received nil instead of an instantiated Arg... structure
var ErrNilArgumentStruct = errors.New("nil argument struct")

// ErrInvalidContainerKey signals that an element does not exist in the container's map
var ErrInvalidContainerKey = errors.New("element does not exist in container")

// ErrContainerKeyAlreadyExists signals that an element was already set in the container's map
var ErrContainerKeyAlreadyExists = errors.New("provided key already exists in container")

// ErrNilRequestHandler signals that a nil request handler interface was provided
var ErrNilRequestHandler = errors.New("nil request handler")

// ErrNilHaveTimeHandler signals that a nil have time handler func was provided
var ErrNilHaveTimeHandler = errors.New("nil have time handler")

// ErrWrongTypeInContainer signals that a wrong type of object was found in container
var ErrWrongTypeInContainer = errors.New("wrong type of object inside container")

// ErrLenMismatch signals that 2 or more slices have different lengths
var ErrLenMismatch = errors.New("lengths mismatch")

// ErrWrongTypeAssertion signals that an type assertion failed
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrHeaderShardDataMismatch signals that shard header does not match created shard info
var ErrHeaderShardDataMismatch = errors.New("shard header does not match shard info")

// ErrNoDataInMessage signals that no data was found after parsing received p2p message
var ErrNoDataInMessage = errors.New("no data found in received message")

// ErrNilBuffer signals that a provided byte buffer is nil
var ErrNilBuffer = errors.New("provided byte buffer is nil")

// ErrNilRandSeed signals that a nil rand seed has been provided
var ErrNilRandSeed = errors.New("provided rand seed is nil")

// ErrNilPrevRandSeed signals that a nil previous rand seed has been provided
var ErrNilPrevRandSeed = errors.New("provided previous rand seed is nil")

// ErrLowerRoundInBlock signals that a header round is too low for processing it
var ErrLowerRoundInBlock = errors.New("header round is lower than last committed")

// ErrHigherRoundInBlock signals that a block with higher round than permitted has been provided
var ErrHigherRoundInBlock = errors.New("higher round in block")

// ErrLowerNonceInBlock signals that a block with lower nonce than permitted has been provided
var ErrLowerNonceInBlock = errors.New("lower nonce in block")

// ErrHigherNonceInBlock signals that a block with higher nonce than permitted has been provided
var ErrHigherNonceInBlock = errors.New("higher nonce in block")

// ErrRandSeedDoesNotMatch signals that random seed does not match with the previous one
var ErrRandSeedDoesNotMatch = errors.New("random seed do not match")

// ErrHeaderNotFinal signals that header is not final and it should be
var ErrHeaderNotFinal = errors.New("header in metablock is not final")

// ErrShardIdMissmatch signals shard ID does not match expectations
var ErrShardIdMissmatch = errors.New("shard ID missmatch")

// ErrNotarizedHeadersSliceIsNil signals that the slice holding notarized headers is nil
var ErrNotarizedHeadersSliceIsNil = errors.New("notarized headers slice is nil")

// ErrNotarizedHeadersSliceForShardIsNil signals that the slice holding notarized headers for shard is nil
var ErrNotarizedHeadersSliceForShardIsNil = errors.New("notarized headers slice for shard is nil")

// ErrCrossShardMBWithoutConfirmationFromMeta signals that miniblock was not yet notarized by metachain
var ErrCrossShardMBWithoutConfirmationFromMeta = errors.New("cross shard miniblock with destination current shard is not confirmed by metachain")

// ErrHeaderBodyMismatch signals that the header does not attest all data from the block
var ErrHeaderBodyMismatch = errors.New("body cannot be validated from header data")

// ErrNilSmartContractProcessor signals that smart contract call executor is nil
var ErrNilSmartContractProcessor = errors.New("smart contract processor is nil")

// ErrNilArgumentParser signals that the argument parser is nil
var ErrNilArgumentParser = errors.New("argument parser is nil")

// ErrNilSCDestAccount signals that destination account is nil
var ErrNilSCDestAccount = errors.New("nil destination SC account")

// ErrWrongNonceInVMOutput signals that nonce in vm output is wrong
var ErrWrongNonceInVMOutput = errors.New("nonce invalid from SC run")

// ErrNilVMOutput signals that vmoutput is nil
var ErrNilVMOutput = errors.New("nil vm output")

// ErrNilValueFromRewardTransaction signals that the transfered value is nil
var ErrNilValueFromRewardTransaction = errors.New("transferred value is nil in reward transaction")

// ErrNilTemporaryAccountsHandler signals that temporary accounts handler is nil
var ErrNilTemporaryAccountsHandler = errors.New("temporary accounts handler is nil")

// ErrNotEnoughValidBlocksInStorage signals that bootstrap from storage failed due to not enough valid blocks stored
var ErrNotEnoughValidBlocksInStorage = errors.New("not enough valid blocks to start from storage")

// ErrNilSmartContractResult signals that the smart contract result is nil
var ErrNilSmartContractResult = errors.New("smart contract result is nil")

// ErrNilRewardTransaction signals that the reward transaction is nil
var ErrNilRewardTransaction = errors.New("reward transaction is nil")

// ErrNilUTxDataPool signals that unsigned transaction pool is nil
var ErrNilUTxDataPool = errors.New("unsigned transactions pool is nil")

// ErrNilRewardTxDataPool signals that the reward transactions pool is nil
var ErrNilRewardTxDataPool = errors.New("reward transactions pool is nil")

// ErrNilUnsignedTxDataPool signals that the unsigned transactions pool is nil
var ErrNilUnsignedTxDataPool = errors.New("unsigned transactions pool is nil")

// ErrNilUTxStorage signals that unsigned transaction storage is nil
var ErrNilUTxStorage = errors.New("unsigned transactions storage is nil")

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

// ErrNilPreProcessor signals that preprocessors is nil
var ErrNilPreProcessor = errors.New("preprocessor is nil")

// ErrNilGasHandler signals that gas handler is nil
var ErrNilGasHandler = errors.New("nil gas handler")

// ErrUnknownBlockType signals that block type is not correct
var ErrUnknownBlockType = errors.New("block type is unknown")

// ErrMissingPreProcessor signals that required pre processor is missing
var ErrMissingPreProcessor = errors.New("pre processor is missing")

// ErrNilAppStatusHandler defines the error for setting a nil AppStatusHandler
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")

// ErrNilInterceptedDataFactory signals that a nil intercepted data factory was provided
var ErrNilInterceptedDataFactory = errors.New("nil intercepted data factory")

// ErrNilInterceptedDataProcessor signals that a nil intercepted data processor was provided
var ErrNilInterceptedDataProcessor = errors.New("nil intercepted data processor")

// ErrNilInterceptorThrottler signals that a nil interceptor throttler was provided
var ErrNilInterceptorThrottler = errors.New("nil interceptor throttler")

// ErrNilUnsignedTxHandler signals that the unsigned tx handler is nil
var ErrNilUnsignedTxHandler = errors.New("nil unsigned tx handler")

// ErrNilTxTypeHandler signals that tx type handler is nil
var ErrNilTxTypeHandler = errors.New("nil tx type handler")

// ErrNilPeerAccountsAdapter signals that a nil peer accounts database was provided
var ErrNilPeerAccountsAdapter = errors.New("nil peer accounts database")

// ErrInvalidPeerAccount signals that a peer account is invalid
var ErrInvalidPeerAccount = errors.New("invalid peer account")

// ErrInvalidMetaHeader signals that a wrong implementation of HeaderHandler was provided
var ErrInvalidMetaHeader = errors.New("invalid header provided, expected MetaBlock")

// ErrNilEpochStartTrigger signals that a nil start of epoch trigger was provided
var ErrNilEpochStartTrigger = errors.New("nil start of epoch trigger")

// ErrNilEpochHandler signals that a nil epoch handler was provided
var ErrNilEpochHandler = errors.New("nil epoch handler")

// ErrNilEpochStartNotifier signals that the ErrNilEpochStartNotifier is nil
var ErrNilEpochStartNotifier = errors.New("nil epochStartNotifier")

// ErrInvalidCacheRefreshIntervalInSec signals that the cacheRefreshIntervalInSec is invalid - zero or less
var ErrInvalidCacheRefreshIntervalInSec = errors.New("invalid cacheRefreshIntervalInSec")

// ErrEpochDoesNotMatch signals that epoch does not match between headers
var ErrEpochDoesNotMatch = errors.New("epoch does not match")

// ErrOverallBalanceChangeFromSC signals that all sumed balance changes are not zero
var ErrOverallBalanceChangeFromSC = errors.New("SC output balance updates are wrong")

// ErrOverflow signals that an overflow occured
var ErrOverflow = errors.New("type overflow occured")

// ErrNilTxValidator signals that a nil tx validator has been provided
var ErrNilTxValidator = errors.New("nil transaction validator")

// ErrNilHdrValidator signals that a nil header validator has been provided
var ErrNilHdrValidator = errors.New("nil header validator")

// ErrNilPendingMiniBlocksHandler signals that a nil pending miniblocks handler has been provided
var ErrNilPendingMiniBlocksHandler = errors.New("nil pending miniblocks handler")

// ErrNilEconomicsFeeHandler signals that fee handler is nil
var ErrNilEconomicsFeeHandler = errors.New("nil economics fee handler")

// ErrSystemBusy signals that the system is busy
var ErrSystemBusy = errors.New("system busy")

// ErrInsufficientGasPriceInTx signals that a lower gas price than required was provided
var ErrInsufficientGasPriceInTx = errors.New("insufficient gas price in tx")

// ErrInsufficientGasLimitInTx signals that a lower gas limit than required was provided
var ErrInsufficientGasLimitInTx = errors.New("insufficient gas limit in tx")

// ErrHigherGasLimitRequiredInTx signals that a higher gas limit was required in tx
var ErrHigherGasLimitRequiredInTx = errors.New("higher gas limit required in tx")

// ErrInvalidMaxGasLimitPerBlock signals that an invalid max gas limit per block has been read from config file
var ErrInvalidMaxGasLimitPerBlock = errors.New("invalid max gas limit per block")

// ErrInvalidGasPerDataByte signals that an invalid gas per data byte has been read from config file
var ErrInvalidGasPerDataByte = errors.New("invalid gas per data byte")

// ErrMaxGasLimitPerMiniBlockInSenderShardIsReached signals that max gas limit per mini block in sender shard has been reached
var ErrMaxGasLimitPerMiniBlockInSenderShardIsReached = errors.New("max gas limit per mini block in sender shard is reached")

// ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached signals that max gas limit per mini block in receiver shard has been reached
var ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached = errors.New("max gas limit per mini block in receiver shard is reached")

// ErrMaxGasLimitPerBlockInSelfShardIsReached signals that max gas limit per block in self shard has been reached
var ErrMaxGasLimitPerBlockInSelfShardIsReached = errors.New("max gas limit per block in self shard is reached")

// ErrInvalidMinimumGasPrice signals that an invalid gas price has been read from config file
var ErrInvalidMinimumGasPrice = errors.New("invalid minimum gas price")

// ErrInvalidMinimumGasLimitForTx signals that an invalid minimum gas limit for transactions has been read from config file
var ErrInvalidMinimumGasLimitForTx = errors.New("invalid minimum gas limit for transactions")

// ErrInvalidRewardsValue signals that an invalid rewards value has been read from config file
var ErrInvalidRewardsValue = errors.New("invalid rewards value")

// ErrInvalidUnBondPeriod signals that an invalid unbond period has been read from config file
var ErrInvalidUnBondPeriod = errors.New("invalid unbond period")

// ErrInvalidRewardsPercentages signals that rewards percentages are not correct
var ErrInvalidRewardsPercentages = errors.New("invalid rewards percentages")

// ErrInvalidInflationPercentages signals that inflation percentages are not correct
var ErrInvalidInflationPercentages = errors.New("invalid inflation percentages")

// ErrInvalidNonceRequest signals that invalid nonce was requested
var ErrInvalidNonceRequest = errors.New("invalid nonce request")

// ErrNilBlockChainHook signals that nil blockchain hook has been provided
var ErrNilBlockChainHook = errors.New("nil blockchain hook")

// ErrNilSCDataGetter signals that a nil sc data getter has been provided
var ErrNilSCDataGetter = errors.New("nil sc data getter")

// ErrNilTxForCurrentBlockHandler signals that nil tx for current block handler has been provided
var ErrNilTxForCurrentBlockHandler = errors.New("nil tx for current block handler")

// ErrNilSCToProtocol signals that nil smart contract to protocol handler has been provided
var ErrNilSCToProtocol = errors.New("nil sc to protocol")

// ErrNilNodesSetup signals that nil nodes setup has been provided
var ErrNilNodesSetup = errors.New("nil nodes setup")

// ErrNilBlackListCacher signals that a nil black list cacher was provided
var ErrNilBlackListCacher = errors.New("nil black list cacher")

// ErrNilPeerShardMapper signals that a nil peer shard mapper has been provided
var ErrNilPeerShardMapper = errors.New("nil peer shard mapper")

// ErrNilBlockTracker signals that a nil block tracker was provided
var ErrNilBlockTracker = errors.New("nil block tracker")

// ErrHeaderIsBlackListed signals that the header provided is black listed
var ErrHeaderIsBlackListed = errors.New("header is black listed")

// ErrNilEconomicsData signals that nil economics data has been provided
var ErrNilEconomicsData = errors.New("nil economics data")

// ErrZeroMaxComputableRounds signals that a value of zero was provided on the maxComputableRounds
var ErrZeroMaxComputableRounds = errors.New("max computable rounds is zero")

// ErrNilRater signals that nil rater has been provided
var ErrNilRater = errors.New("nil rater")

// ErrNilNetworkWatcher signals that a nil network watcher has been provided
var ErrNilNetworkWatcher = errors.New("nil network watcher")

// ErrNilHeaderValidator signals that nil header validator has been provided
var ErrNilHeaderValidator = errors.New("nil header validator")

// ErrMaxRatingIsSmallerThanMinRating signals that the max rating is smaller than the min rating value
var ErrMaxRatingIsSmallerThanMinRating = errors.New("max rating is smaller than min rating")

// ErrMinRatingSmallerThanOne signals that the min rating is smaller than the min value of 1
var ErrMinRatingSmallerThanOne = errors.New("min rating is smaller than one")

// ErrStartRatingNotBetweenMinAndMax signals that the start rating is not between min and max rating
var ErrStartRatingNotBetweenMinAndMax = errors.New("start rating is not between min and max rating")

// ErrSignedBlocksThresholdNotBetweenZeroAndOne signals that the signed blocks threshold is not between 0 and 1
var ErrSignedBlocksThresholdNotBetweenZeroAndOne = errors.New("signed blocks threshold is not between 0 and 1")

// ErrConsecutiveMissedBlocksPenaltyLowerThanOne signals that the ConsecutiveMissedBlocksPenalty is lower than 1
var ErrConsecutiveMissedBlocksPenaltyLowerThanOne = errors.New("consecutive missed blocks penalty lower than 1")

// ErrDecreaseRatingsStepMoreThanMinusOne signals that the decrease rating step has a vale greater than -1
var ErrDecreaseRatingsStepMoreThanMinusOne = errors.New("decrease rating step has a value greater than -1")

// ErrHoursToMaxRatingFromStartRatingZero signals that the number of hours to reach max rating step is zero
var ErrHoursToMaxRatingFromStartRatingZero = errors.New("hours to reach max rating is zero")

// ErrSCDeployFromSCRIsNotPermitted signals that operation is not permitted
var ErrSCDeployFromSCRIsNotPermitted = errors.New("it is not permitted to deploy a smart contract from another smart contract cross shard")

// ErrNotEnoughGas signals that not enough gas has been provided
var ErrNotEnoughGas = errors.New("not enough gas was sent in the transaction")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value provided")

// ErrNilQuotaStatusHandler signals that a nil quota status handler has been provided
var ErrNilQuotaStatusHandler = errors.New("nil quota status handler")

// ErrNilAntifloodHandler signals that a nil antiflood handler has been provided
var ErrNilAntifloodHandler = errors.New("nil antiflood handler")

// ErrNilHeaderSigVerifier signals that a nil header sig verifier has been provided
var ErrNilHeaderSigVerifier = errors.New("nil header sig verifier")

// ErrNilHeaderIntegrityVerifier signals that a nil header integrity verifier has been provided
var ErrNilHeaderIntegrityVerifier = errors.New("nil header integrity verifier")

// ErrFailedTransaction signals that transaction is of type failed.
var ErrFailedTransaction = errors.New("failed transaction, gas consumed")

// ErrNilBadTxHandler signals that bad tx handler is nil
var ErrNilBadTxHandler = errors.New("nil bad tx handler")

// ErrNilReceiptHandler signals that receipt handler is nil
var ErrNilReceiptHandler = errors.New("nil receipt handler")

// ErrTooManyReceiptsMiniBlocks signals that there were too many receipts miniblocks created
var ErrTooManyReceiptsMiniBlocks = errors.New("too many receipts miniblocks")

// ErrReceiptsHashMissmatch signals that overall receipts has does not match
var ErrReceiptsHashMissmatch = errors.New("receipts hash missmatch")

// ErrMiniBlockNumMissMatch signals that number of miniblocks does not match
var ErrMiniBlockNumMissMatch = errors.New("num miniblocks does not match")

// ErrEpochStartDataDoesNotMatch signals that EpochStartData is not the same as the leader created
var ErrEpochStartDataDoesNotMatch = errors.New("epoch start data does not match")

// ErrInvalidMinStepValue signals the min step value is invalid
var ErrInvalidMinStepValue = errors.New("invalid min step value")

// ErrNotEpochStartBlock signals that block is not of type epoch start
var ErrNotEpochStartBlock = errors.New("not epoch start block")

// ErrGettingShardDataFromEpochStartData signals that could not get shard data from previous epoch start block
var ErrGettingShardDataFromEpochStartData = errors.New("could not find shard data from previous epoch start metablock")

// ErrNilValidityAttester signals that a nil validity attester has been provided
var ErrNilValidityAttester = errors.New("nil validity attester")

// ErrNilHeaderHandler signals that a nil header handler has been provided
var ErrNilHeaderHandler = errors.New("nil header handler")

// ErrNilMiniBlocksProvider signals that a nil miniblocks data provider has been passed over
var ErrNilMiniBlocksProvider = errors.New("nil miniblocks provider")

// ErrNilWhiteListHandler signals that white list handler is nil
var ErrNilWhiteListHandler = errors.New("nil whitelist handler")

// ErrMiniBlocksInWrongOrder signals the miniblocks are in wrong order
var ErrMiniBlocksInWrongOrder = errors.New("miniblocks in wrong order, should have been only from me")

// ErrEmptyTopic signals that an empty topic has been provided
var ErrEmptyTopic = errors.New("empty topic")

// ErrInvalidArguments signals that invalid arguments were given to process built-in function
var ErrInvalidArguments = errors.New("invalid arguments to process built-in function")

// ErrNilBuiltInFunction signals that built in function is nil
var ErrNilBuiltInFunction = errors.New("built in function is nil")

// ErrRewardMiniBlockNotFromMeta signals that miniblock has a different sender shard than meta
var ErrRewardMiniBlockNotFromMeta = errors.New("rewards miniblocks should come only from meta")

// ErrValidatorInfoMiniBlockNotFromMeta signals that miniblock has a different sender shard than meta
var ErrValidatorInfoMiniBlockNotFromMeta = errors.New("validatorInfo miniblocks should come only from meta")

// ErrAccumulatedFeesDoNotMatch signals that accumulated fees do not match
var ErrAccumulatedFeesDoNotMatch = errors.New("accumulated fees do not match")

// ErrDeveloperFeesDoNotMatch signals that developer fees do not match
var ErrDeveloperFeesDoNotMatch = errors.New("developer fees do not match")

// ErrAccumulatedFeesInEpochDoNotMatch signals that accumulated fees in epoch do not match
var ErrAccumulatedFeesInEpochDoNotMatch = errors.New("accumulated fees in epoch do not match")

// ErrDevFeesInEpochDoNotMatch signals that developer fees in epoch do not match
var ErrDevFeesInEpochDoNotMatch = errors.New("developer fees in epoch do not match")

// ErrNilRewardsHandler signals that rewards handler is nil
var ErrNilRewardsHandler = errors.New("rewards handler is nil")

// ErrNilEpochEconomics signals that nil end of epoch econimics was provided
var ErrNilEpochEconomics = errors.New("nil epoch economics")

// ErrNilEpochStartDataCreator signals that nil epoch start data creator was provided
var ErrNilEpochStartDataCreator = errors.New("nil epoch start data creator")

// ErrNilEpochStartRewardsCreator signals that nil epoch start rewards creator was provided
var ErrNilEpochStartRewardsCreator = errors.New("nil epoch start rewards creator")

// ErrNilEpochStartValidatorInfoCreator signals that nil epoch start validator info creator was provided
var ErrNilEpochStartValidatorInfoCreator = errors.New("nil epoch start validator info creator")

// ErrInvalidGenesisTotalSupply signals that invalid genesis total supply was provided
var ErrInvalidGenesisTotalSupply = errors.New("invalid genesis total supply")

// ErrInvalidAuctionEnableNonce signals that auction enable nonce is invalid
var ErrInvalidAuctionEnableNonce = errors.New("invalid auction enable nonce")

// ErrInvalidStakingEnableNonce signals that the staking enable nonce is invalid
var ErrInvalidStakingEnableNonce = errors.New("invalid staking enable nonce")

// ErrInvalidUnJailPrice signals that invalid unjail price was provided
var ErrInvalidUnJailPrice = errors.New("invalid unjail price")

// ErrOperationNotPermitted signals that operation is not permitted
var ErrOperationNotPermitted = errors.New("operation in account not permitted")

// ErrInvalidAddressLength signals that address length is invalid
var ErrInvalidAddressLength = errors.New("invalid address length")

// ErrDuplicateThreshold signals that two thresholds are the same
var ErrDuplicateThreshold = errors.New("two thresholds are the same")

// ErrNoChancesForMaxThreshold signals that the max threshold has no chance defined
var ErrNoChancesForMaxThreshold = errors.New("max threshold has no chances")

// ErrNoChancesProvided signals that there were no chances provided
var ErrNoChancesProvided = errors.New("no chances are provided")

// ErrNilMinChanceIfZero signals that there was no min chance provided if a chance is still needed
var ErrNilMinChanceIfZero = errors.New("no min chance ")

// ErrInvalidShardCacherIdentifier signals an invalid identifier
var ErrInvalidShardCacherIdentifier = errors.New("invalid identifier for shard cacher")

// ErrMaxBlockSizeReached signals that max block size has been reached
var ErrMaxBlockSizeReached = errors.New("max block size has been reached")

// ErrBlockBodyHashMismatch signals that block body hashes does not match
var ErrBlockBodyHashMismatch = errors.New("block bodies does not match")

// ErrInvalidMiniBlockType signals that an invalid miniblock type has been provided
var ErrInvalidMiniBlockType = errors.New("invalid miniblock type")

// ErrInvalidBody signals that an invalid body has been provided
var ErrInvalidBody = errors.New("invalid body")

// ErrNilBlockSizeComputationHandler signals that a nil block size computation handler has been provided
var ErrNilBlockSizeComputationHandler = errors.New("nil block size computation handler")

// ErrNilValidatorStatistics signals that a nil validator statistics has been provided
var ErrNilValidatorStatistics = errors.New("nil validator statistics")

// ErrAccountNotFound signals that the account was not found for the provided address
var ErrAccountNotFound = errors.New("account not found")

// ErrMaxRatingZero signals that maxrating with a value of zero has been provided
var ErrMaxRatingZero = errors.New("max rating is zero")

// ErrNilValidatorInfos signals that a nil validator infos has been provided
var ErrNilValidatorInfos = errors.New("nil validator infos")

// ErrNilBlockSizeThrottler signals that block size throttler si nil
var ErrNilBlockSizeThrottler = errors.New("block size throttler is nil")

// ErrInvalidMetaTransaction signals that meta transaction is invalid
var ErrInvalidMetaTransaction = errors.New("meta transaction is invalid")

// ErrLogNotFound is the error returned when a transaction has no logs
var ErrLogNotFound = errors.New("no logs for queried transaction")

// ErrNilTxLogsProcessor is the error returned when a transaction has no logs
var ErrNilTxLogsProcessor = errors.New("nil transaction logs processor")

// ErrIncreaseStepLowerThanOne signals that an increase step lower than one has been provided
var ErrIncreaseStepLowerThanOne = errors.New("increase step is lower than one")

// ErrNilVmInput signals that provided vm input is nil
var ErrNilVmInput = errors.New("nil vm input")

// ErrNilDnsAddresses signals that nil dns addresses map was provided
var ErrNilDnsAddresses = errors.New("nil dns addresses map")

// ErrNilCommunityAddress signals that a nil community address was provided
var ErrNilCommunityAddress = errors.New("nil community address")

// ErrCallerIsNotTheDNSAddress signals that called address is not the DNS address
var ErrCallerIsNotTheDNSAddress = errors.New("not a dns address")

// ErrUserNameChangeIsDisabled signals the user name change is not allowed
var ErrUserNameChangeIsDisabled = errors.New("user name change is disabled")

// ErrDestinationNotInSelfShard signals that user is not in self shard
var ErrDestinationNotInSelfShard = errors.New("destination is not in self shard")

// ErrUserNameDoesNotMatch signals that user name does not match
var ErrUserNameDoesNotMatch = errors.New("user name does not match")

// ErrUserNameDoesNotMatchInCrossShardTx signals that user name does not match in case of cross shard tx
var ErrUserNameDoesNotMatchInCrossShardTx = errors.New("user name does not match in destination shard")

// ErrNilBalanceComputationHandler signals that a nil balance computation handler has been provided
var ErrNilBalanceComputationHandler = errors.New("nil balance computation handler")

// ErrNilRatingsInfoHandler signals that nil ratings info handler has been provided
var ErrNilRatingsInfoHandler = errors.New("nil ratings info handler")

// ErrNilDebugger signals that a nil debug handler has been provided
var ErrNilDebugger = errors.New("nil debug handler")

// ErrBuiltInFunctionCalledWithValue signals that builtin function was called with value that is not allowed
var ErrBuiltInFunctionCalledWithValue = errors.New("built in function called with tx value is not allowed")

// ErrEmptySoftwareVersion signals that empty software version was called
var ErrEmptySoftwareVersion = errors.New("empty software version")

// ErrEmptyFloodPreventerList signals that an empty flood preventer list has been provided
var ErrEmptyFloodPreventerList = errors.New("empty flood preventer provided")

// ErrNilTopicFloodPreventer signals that a nil topic flood preventer has been provided
var ErrNilTopicFloodPreventer = errors.New("nil topic flood preventer")

// ErrOriginatorIsBlacklisted signals that a message originator is blacklisted on the current node
var ErrOriginatorIsBlacklisted = errors.New("originator is blacklisted")

// ErrShardIsStuck signals that a shard is stuck
var ErrShardIsStuck = errors.New("shard is stuck")

// ErrRelayedTxBeneficiaryDoesNotMatchReceiver signals that an invalid address was provided in the relayed tx
var ErrRelayedTxBeneficiaryDoesNotMatchReceiver = errors.New("invalid address in relayed tx")

// ErrInvalidVMType signals that invalid vm type was provided
var ErrInvalidVMType = errors.New("invalid VM type")

// ErrRecursiveRelayedTxIsNotAllowed signals that recursive relayed tx is not allowed
var ErrRecursiveRelayedTxIsNotAllowed = errors.New("recursive relayed tx is not allowed")

// ErrRelayedTxValueHigherThenUserTxValue signals that relayed tx value is higher then user tx value
var ErrRelayedTxValueHigherThenUserTxValue = errors.New("relayed tx value is higher than user tx value")

// ErrNilInterceptorContainer signals that nil interceptor container has been provided
var ErrNilInterceptorContainer = errors.New("nil interceptor container")

// ErrInvalidChainID signals that an invalid chain ID has been provided
var ErrInvalidChainID = errors.New("invalid chain ID")

// ErrInvalidTransactionVersion signals  that an invalid transaction version has been provided
var ErrInvalidTransactionVersion = errors.New("invalid transaction version")

// ErrTxValueTooBig signals that transaction value is too big
var ErrTxValueTooBig = errors.New("tx value is too big")

// ErrInvalidUserNameLength signals that provided user name length is invalid
var ErrInvalidUserNameLength = errors.New("invalid user name length")

// ErrTxValueOutOfBounds signals that transaction value is out of bounds
var ErrTxValueOutOfBounds = errors.New("tx value is out of bounds")

// ErrNilBlackListedPkCache signals that a nil black listed public key cache has been provided
var ErrNilBlackListedPkCache = errors.New("nil black listed public key cache")

// ErrInvalidDecayCoefficient signals that the provided decay coefficient is invalid
var ErrInvalidDecayCoefficient = errors.New("decay coefficient is invalid")

// ErrInvalidDecayIntervalInSeconds signals that an invalid interval in seconds was provided
var ErrInvalidDecayIntervalInSeconds = errors.New("invalid decay interval in seconds")

// ErrInvalidMinScore signals that an invalid minimum score was provided
var ErrInvalidMinScore = errors.New("invalid minimum score")

// ErrInvalidMaxScore signals that an invalid maximum score was provided
var ErrInvalidMaxScore = errors.New("invalid maximum score")

// ErrInvalidUnitValue signals that an invalid unit value was provided
var ErrInvalidUnitValue = errors.New("invalid unit value")

// ErrInvalidBadPeerThreshold signals that an invalid bad peer threshold has been provided
var ErrInvalidBadPeerThreshold = errors.New("invalid bad peer threshold")

// ErrNilPeerValidatorMapper signals that nil peer validator mapper has been provided
var ErrNilPeerValidatorMapper = errors.New("nil peer validator mapper")

// ErrOnlyValidatorsCanUseThisTopic signals that topic can be used by validator only
var ErrOnlyValidatorsCanUseThisTopic = errors.New("only validators can use this topic")

// ErrTransactionIsNotWhitelisted signals that a transaction is not whitelisted
var ErrTransactionIsNotWhitelisted = errors.New("transaction is not whitelisted")

// ErrInterceptedDataNotForCurrentShard signals that intercepted data is not for current shard
var ErrInterceptedDataNotForCurrentShard = errors.New("intercepted data not for current shard")
