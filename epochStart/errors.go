package epochStart

import "errors"

// ErrNilArgsNewMetaEpochStartTrigger signals that nil arguments were provided
var ErrNilArgsNewMetaEpochStartTrigger = errors.New("nil arguments for meta start of epoch trigger")

// ErrNilEpochStartSettings signals that nil start of epoch settings has been provided
var ErrNilEpochStartSettings = errors.New("nil start of epoch settings")

// ErrInvalidSettingsForEpochStartTrigger signals that settings for start of epoch trigger are invalid
var ErrInvalidSettingsForEpochStartTrigger = errors.New("invalid start of epoch trigger settings")

// ErrNilArgsNewShardEpochStartTrigger signals that nil arguments for shard epoch trigger has been provided
var ErrNilArgsNewShardEpochStartTrigger = errors.New("nil arguments for shard start of epoch trigger")

// ErrNilEpochStartNotifier signals that nil epoch start notifier has been provided
var ErrNilEpochStartNotifier = errors.New("nil epoch start notifier")

// ErrNotEnoughRoundsBetweenEpochs signals that not enough rounds has passed since last epoch start
var ErrNotEnoughRoundsBetweenEpochs = errors.New("tried to force start of epoch before passing of enough rounds")

// ErrForceEpochStartCanBeCalledOnlyOnNewRound signals that force start of epoch was called on wrong round
var ErrForceEpochStartCanBeCalledOnlyOnNewRound = errors.New("invalid time to call force start of epoch, possible only on new round")

// ErrSavedRoundIsHigherThanInputRound signals that input round was wrong
var ErrSavedRoundIsHigherThanInputRound = errors.New("saved round is higher than input round")

// ErrSavedRoundIsHigherThanInput signals that input round was wrong
var ErrSavedRoundIsHigherThanInput = errors.New("saved round is higher than input round")

// ErrWrongTypeAssertion signals wrong type assertion
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilMarshalizer signals that nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilTxSignMarshalizer signals that nil tx sign marshalizer has been provided
var ErrNilTxSignMarshalizer = errors.New("nil tx sign marshalizer")

// ErrNilStorage signals that nil storage has been provided
var ErrNilStorage = errors.New("nil storage")

// ErrNilHeaderHandler signals that a nil header handler has been provided
var ErrNilHeaderHandler = errors.New("nil header handler")

// ErrNilMiniblocks signals that nil argument was passed
var ErrNilMiniblocks = errors.New("nil arguments for miniblocks object")

// ErrNilMiniblock signals that nil miniblock has been provided
var ErrNilMiniblock = errors.New("nil miniblock")

// ErrMetaHdrNotFound signals that metaheader was not found
var ErrMetaHdrNotFound = errors.New("meta header not found")

// ErrNilHasher signals that nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher")

// ErrInvalidConsensusThreshold signals that an invalid consensus threshold has been provided
var ErrInvalidConsensusThreshold = errors.New("invalid consensus threshold")

// ErrNilHeaderValidator signals that nil header validator has been provided
var ErrNilHeaderValidator = errors.New("nil header validator")

// ErrNilDataPoolsHolder signals that nil data pools holder has been provided
var ErrNilDataPoolsHolder = errors.New("nil data pools holder")

// ErrNilStorageService signals that nil storage service has been provided
var ErrNilStorageService = errors.New("nil storage service")

// ErrNilRequestHandler signals that nil request handler has been provided
var ErrNilRequestHandler = errors.New("nil request handler")

// ErrNilMetaBlockStorage signals that nil metablocks storage has been provided
var ErrNilMetaBlockStorage = errors.New("nil metablocks storage")

// ErrNilMetaBlocksPool signals that nil metablock pools holder has been provided
var ErrNilMetaBlocksPool = errors.New("nil metablocks pool")

// ErrNilValidatorInfoProcessor signals that a nil validator info processor has been provided
var ErrNilValidatorInfoProcessor = errors.New("nil validator info processor")

// ErrNilUint64Converter signals that nil uint64 converter has been provided
var ErrNilUint64Converter = errors.New("nil uint64 converter")

// ErrNilTriggerStorage signals that nil meta header storage has been provided
var ErrNilTriggerStorage = errors.New("nil trigger storage")

// ErrNilMetaNonceHashStorage signals that nil meta header nonce hash storage has been provided
var ErrNilMetaNonceHashStorage = errors.New("nil meta nonce hash storage")

// ErrValidatorMiniBlockHashDoesNotMatch signals that created and received validatorInfo miniblock hash does not match
var ErrValidatorMiniBlockHashDoesNotMatch = errors.New("validatorInfo miniblock hash does not match")

// ErrRewardMiniBlockHashDoesNotMatch signals that created and received rewards miniblock hash does not match
var ErrRewardMiniBlockHashDoesNotMatch = errors.New("reward miniblock hash does not match")

// ErrNilShardCoordinator is raised when a valid shard coordinator is expected but nil used
var ErrNilShardCoordinator = errors.New("shard coordinator is nil")

// ErrNilPubkeyConverter signals that nil address converter was provided
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrRewardMiniBlocksNumDoesNotMatch signals that number of created and received rewards miniblocks is not equal
var ErrRewardMiniBlocksNumDoesNotMatch = errors.New("number of created and received rewards miniblocks missmatch")

// ErrNilRewardsHandler signals that rewards handler is nil
var ErrNilRewardsHandler = errors.New("rewards handler is nil")

// ErrNilTotalAccumulatedFeesInEpoch signals that total accumulated fees in epoch is nil
var ErrNilTotalAccumulatedFeesInEpoch = errors.New("total accumulated fees in epoch is nil")

// ErrEndOfEpochEconomicsDataDoesNotMatch signals that end of epoch data does not match
var ErrEndOfEpochEconomicsDataDoesNotMatch = errors.New("end of epoch economics data does not match")

// ErrNilRounder signals that an operation has been attempted to or with a nil Rounder implementation
var ErrNilRounder = errors.New("nil Rounder")

// ErrNilNodesCoordinator signals that an operation has been attempted to or with a nil nodes coordinator
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrNotEpochStartBlock signals that block is not of type epoch start
var ErrNotEpochStartBlock = errors.New("not epoch start block")

// ErrNilShardHeaderStorage signals that shard header storage is nil
var ErrNilShardHeaderStorage = errors.New("nil shard header storage")

// ErrValidatorInfoMiniBlocksNumDoesNotMatch signals that number of created and received validatorInfo miniblocks is not equal
var ErrValidatorInfoMiniBlocksNumDoesNotMatch = errors.New("number of created and received validatorInfo miniblocks missmatch")

// ErrNilValidatorInfo signals that a nil value for the validatorInfo has been provided
var ErrNilValidatorInfo = errors.New("validator info is nil")

// ErrNilMetaBlock signals that a nil metablock has been provided
var ErrNilMetaBlock = errors.New("nil metablock")

// ErrNilMiniBlockPool signals that a nil mini blocks pool was used
var ErrNilMiniBlockPool = errors.New("nil mini block pool")

// ErrNilStatusHandler signals that a nil status handler has been provided
var ErrNilStatusHandler = errors.New("nil app status handler")

// ErrEpochStartDataForShardNotFound signals that epoch start shard data was not found for current shard id
var ErrEpochStartDataForShardNotFound = errors.New("epoch start data for current shard not found")

// ErrMissingHeader signals that searched header is missing
var ErrMissingHeader = errors.New("missing header")

// ErrNilPathManager signals that a nil path manager has been provided
var ErrNilPathManager = errors.New("nil path manager")

// ErrNilMessenger signals that a nil messenger has been provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilEconomicsData signals that a nil economics data handler has been provided
var ErrNilEconomicsData = errors.New("nil economics data")

// ErrNilPubKey signals that a nil public key has been provided
var ErrNilPubKey = errors.New("nil public key")

// ErrNilBlockKeyGen signals that a nil block key generator has been provided
var ErrNilBlockKeyGen = errors.New("nil block key generator")

// ErrNilKeyGen signals that a nil key generator has been provided
var ErrNilKeyGen = errors.New("nil key generator")

// ErrNilSingleSigner signals that a nil single signer has been provided
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrNilBlockSingleSigner signals that a nil block single signer has been provided
var ErrNilBlockSingleSigner = errors.New("nil block single signer")

// ErrNilGenesisNodesConfig signals that a nil genesis nodes config has been provided
var ErrNilGenesisNodesConfig = errors.New("nil genesis nodes config")

// ErrNilRater signals that a nil rater has been provided
var ErrNilRater = errors.New("nil rater")

// ErrNilTriesContainer signals that a nil tries container has been provided
var ErrNilTriesContainer = errors.New("nil tries container")

// ErrNilTrieStorageManager signals that a nil trie storage managers map has been provided
var ErrNilTrieStorageManager = errors.New("nil trie storage managers map")

// ErrInvalidDefaultDBPath signals that an invalid default database path has been provided
var ErrInvalidDefaultDBPath = errors.New("invalid default db path")

// ErrInvalidDefaultEpochString signals that an invalid default epoch string has been provided
var ErrInvalidDefaultEpochString = errors.New("invalid default epoch string")

// ErrInvalidDefaultShardString signals that an invalid default shard string has been provided
var ErrInvalidDefaultShardString = errors.New("invalid default shard string")

// ErrInvalidWorkingDir signals that an invalid working directory has been provided
var ErrInvalidWorkingDir = errors.New("invalid working directory")

// ErrTimeoutWaitingForMetaBlock signals that a timeout event was raised while waiting for the epoch start meta block
var ErrTimeoutWaitingForMetaBlock = errors.New("timeout while waiting for epoch start meta block")
