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

// ErrNilProtocolSustainabilityAddress signals that a nil protocol sustainability address was provided
var ErrNilProtocolSustainabilityAddress = errors.New("nil protocol sustainability address")

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

// ErrNilCoreComponentsHolder signals that a nil core components holder was provided
var ErrNilCoreComponentsHolder = errors.New("nil core components holder")

// ErrNilCryptoComponentsHolder signals that a nil crypto components holder was provided
var ErrNilCryptoComponentsHolder = errors.New("nil crypto components holder")

// ErrRewardMiniBlocksNumDoesNotMatch signals that number of created and received rewards miniblocks is not equal
var ErrRewardMiniBlocksNumDoesNotMatch = errors.New("number of created and received rewards miniblocks missmatch")

// ErrNilRewardsHandler signals that rewards handler is nil
var ErrNilRewardsHandler = errors.New("rewards handler is nil")

// ErrNilTotalAccumulatedFeesInEpoch signals that total accumulated fees in epoch is nil
var ErrNilTotalAccumulatedFeesInEpoch = errors.New("total accumulated fees in epoch is nil")

// ErrNilTotalDevFeesInEpoch signals that total developer fees in epoch is nil
var ErrNilTotalDevFeesInEpoch = errors.New("total developer fees in epoch is nil")

// ErrEndOfEpochEconomicsDataDoesNotMatch signals that end of epoch data does not match
var ErrEndOfEpochEconomicsDataDoesNotMatch = errors.New("end of epoch economics data does not match")

// ErrNilRoundHandler signals that an operation has been attempted to or with a nil RoundHandler implementation
var ErrNilRoundHandler = errors.New("nil RoundHandler")

// ErrNilNodesConfigProvider signals that an operation has been attempted to or with a nil nodes config provider
var ErrNilNodesConfigProvider = errors.New("nil nodes config provider")

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

// ErrInvalidWorkingDir signals that an invalid working directory has been provided
var ErrInvalidWorkingDir = errors.New("invalid working directory")

// ErrTimeoutWaitingForMetaBlock signals that a timeout event was raised while waiting for the epoch start meta block
var ErrTimeoutWaitingForMetaBlock = errors.New("timeout while waiting for epoch start meta block")

// ErrNilStorageUnitOpener signals that a nil storage unit opener was provided
var ErrNilStorageUnitOpener = errors.New("nil storage unit opener")

// ErrNilLatestStorageDataProvider signals that a nil latest storage data provider was provided
var ErrNilLatestStorageDataProvider = errors.New("nil latest storage data provider")

// ErrNilShuffler signals that a nil shuffler was provided
var ErrNilShuffler = errors.New("nil nodes shuffler provided")

// ErrNotEnoughNumConnectedPeers signals that config is invalid for num of connected peers
var ErrNotEnoughNumConnectedPeers = errors.New("not enough min num of connected peers from config")

// ErrNotEnoughNumOfPeersToConsiderBlockValid signals that config is invalid for num of peer to consider block valid
var ErrNotEnoughNumOfPeersToConsiderBlockValid = errors.New("not enough num of peers to consider block valid from config")

// ErrNilArgumentsParser signals that nil arguments parser was provided
var ErrNilArgumentsParser = errors.New("nil arguments parsers")

// ErrProtocolSustainabilityAddressInMetachain signals that protocol sustainability address is in metachain which is not allowed
var ErrProtocolSustainabilityAddressInMetachain = errors.New("protocol sustainability address in metachain")

// ErrNilGenesisTotalSupply signals that nil genesis total supply has been provided
var ErrNilGenesisTotalSupply = errors.New("nil genesis total supply")

// ErrNilAccountsDB signals that nil accounts DB has been provided
var ErrNilAccountsDB = errors.New("nil accounts DB")

// ErrNilSystemVM signals that nil system VM has been provided
var ErrNilSystemVM = errors.New("nil system VM")

// ErrStakingSCOutputAccountNotFound signals that staking sc output account was not found
var ErrStakingSCOutputAccountNotFound = errors.New("staking sc output was not found")

// ErrNilEndOfEpochCallerAddress signals that nil end of epoch caller address was provided
var ErrNilEndOfEpochCallerAddress = errors.New("nil end of epoch caller address")

// ErrNilStakingSCAddress signals that nil staking system smart contract address was provided
var ErrNilStakingSCAddress = errors.New("nil staking system smart contract address")

// ErrNilHeaderIntegrityVerifier signals that a nil header integrity verifier was provided
var ErrNilHeaderIntegrityVerifier = errors.New("nil header integrity verifier handler")

// ErrNilChanceComputer signals that nil chance computer has been provided
var ErrNilChanceComputer = errors.New("nil chance computer")

// ErrInvalidMinNumberOfNodes signals that the minimum number of nodes is invalid
var ErrInvalidMinNumberOfNodes = errors.New("minimum number of nodes invalid")

// ErrNilEpochNotifier signals that the provided EpochNotifier is nil
var ErrNilEpochNotifier = errors.New("nil EpochNotifier")

// ErrNilMetablockProcessor signals that a nil metablock processor was provided
var ErrNilMetablockProcessor = errors.New("nil metablock processor")

// ErrCouldNotInitDelegationSystemSC signals that delegation system sc init failed
var ErrCouldNotInitDelegationSystemSC = errors.New("could not init delegation system sc")

// ErrNilLocalTxCache signals that nil local tx cache has been provided
var ErrNilLocalTxCache = errors.New("nil local tx cache")

// ErrSystemDelegationCall signals that delegation system sc returned with error
var ErrSystemDelegationCall = errors.New("delegation system sc call returned with error")

// ErrExecutingSystemScCode signals that an error occurred while executing system SC code
var ErrExecutingSystemScCode = errors.New("error executing system SC code")

// ErrNilSystemVmInstance signals that a nil system VM instance was provided
var ErrNilSystemVmInstance = errors.New("nil system VM instance")

// ErrNilStakingDataProvider signals that a nil staking data provider was used
var ErrNilStakingDataProvider = errors.New("nil staking data provider")

// ErrNilDataTrie signals that a nil data trie was provided
var ErrNilDataTrie = errors.New("nil data trie")

// ErrInvalidMinNodePrice signals that the minimum node price is invalid (e.g negative, not a number, etc)
var ErrInvalidMinNodePrice = errors.New("minimum node price is invalid")

// ErrInvalidRewardsTopUpGradientPoint signals that the given point controlling the top-up gradient is invalid
var ErrInvalidRewardsTopUpGradientPoint = errors.New("top-up gradient point invalid")

// ErrInvalidRewardsTopUpFactor signals that the factor for computing the top-up rewards out of the full rewards is invalid
var ErrInvalidRewardsTopUpFactor = errors.New("top-up factor invalid")

// ErrNilEconomicsDataProvider signals that the economics data provider is nil
var ErrNilEconomicsDataProvider = errors.New("end of epoch economics data provider is nil")

// ErrInvalidMaxNumberOfNodes signals that the maximum number of nodes is invalid
var ErrInvalidMaxNumberOfNodes = errors.New("maximum number of nodes invalid")

// ErrInvalidSystemSCReturn signals that system sc returned with something unexpected
var ErrInvalidSystemSCReturn = errors.New("invalid system sc return")

// ErrUnStakeExecuteError signals that unstaked returned with error
var ErrUnStakeExecuteError = errors.New("unstake execution error")

// ErrSystemValidatorSCCall signals that system validator sc call failed
var ErrSystemValidatorSCCall = errors.New("system validator sc call failed")

// ErrOwnerDoesntHaveEligibleNodesInEpoch signals that the owner doesn't have any eligible nodes in epoch
var ErrOwnerDoesntHaveEligibleNodesInEpoch = errors.New("owner has no eligible nodes in epoch")

// ErrInvalidMaxHardCapForMissingNodes signals that the maximum hardcap value for missing nodes is invalid
var ErrInvalidMaxHardCapForMissingNodes = errors.New("invalid max hardcap for missing nodes")

// ErrInvalidNumConcurrentTrieSyncers signals that the number of concurrent trie syncers is invalid
var ErrInvalidNumConcurrentTrieSyncers = errors.New("invalid num concurrent trie syncers")

// ErrInvalidInflationRate signals usage of an invalid inflation rate
var ErrInvalidInflationRate = errors.New("inflation rate is invalid")

// ErrInvalidAccumulatedFees signals that the accumulated fees are invalid
var ErrInvalidAccumulatedFees = errors.New("accumulated rewards are invalid")

// ErrInvalidEstimatedProtocolSustainabilityRewards signals that the estimated protocol sustainability rewards are invalid
var ErrInvalidEstimatedProtocolSustainabilityRewards = errors.New("estimated protocol sustainability rewards are invalid")

// ErrInvalidAmountMintedTokens signals that the amount of minted tokens is invalid
var ErrInvalidAmountMintedTokens = errors.New("amount of minted tokens is invalid")

// ErrInvalidTotalToDistribute signals that the computed total rewards to be distributed is invalid
var ErrInvalidTotalToDistribute = errors.New("invalid total to distribute")

// ErrInvalidRewardsPerBlock signals that the computed rewards per block are invalid
var ErrInvalidRewardsPerBlock = errors.New("invalid rewards per block")

// ErrResetLastUnJailedFromQueue signals that reset unjailed from queue failed
var ErrResetLastUnJailedFromQueue = errors.New("reset last unjailed from queue failed")
