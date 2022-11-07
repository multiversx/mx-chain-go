package vm

import "errors"

// ErrUnknownSystemSmartContract signals that there is no system smart contract on the provided address
var ErrUnknownSystemSmartContract = errors.New("missing system smart contract on selected address")

// ErrNilSystemEnvironmentInterface signals that a nil system environment interface was provided
var ErrNilSystemEnvironmentInterface = errors.New("system environment interface is nil")

// ErrNilSystemContractsContainer signals that the provided system contract container is nil
var ErrNilSystemContractsContainer = errors.New("system contract container is nil")

// ErrNilVMType signals that the provided vm type is nil
var ErrNilVMType = errors.New("vm type is nil")

// ErrInputArgsIsNil signals that input arguments are nil for system smart contract
var ErrInputArgsIsNil = errors.New("input system smart contract arguments are nil")

// ErrInputCallValueIsNil signals that input call value is nil for system smart contract
var ErrInputCallValueIsNil = errors.New("input value for system smart contract is nil")

// ErrInputFunctionIsNil signals that input function is nil for system smart contract
var ErrInputFunctionIsNil = errors.New("input function for system smart contract is nil")

// ErrInputCallerAddrIsNil signals that input caller address is nil for system smart contract
var ErrInputCallerAddrIsNil = errors.New("input called address for system smart contract is nil")

// ErrInputRecipientAddrIsNil signals that input recipient address for system smart contract is nil
var ErrInputRecipientAddrIsNil = errors.New("input recipient address for system smart contract is nil")

// ErrNilBlockchainHook signals that blockchain hook is nil
var ErrNilBlockchainHook = errors.New("blockchain hook is nil")

// ErrNilCryptoHook signals that crypto hook is nil
var ErrNilCryptoHook = errors.New("crypto hook is nil")

// ErrNilOrEmptyKey signals that key is nil or empty
var ErrNilOrEmptyKey = errors.New("nil or empty key")

// ErrNilEconomicsData signals that nil economics data has been provided
var ErrNilEconomicsData = errors.New("nil economics data")

// ErrNegativeInitialStakeValue signals that a negative initial stake value was provided
var ErrNegativeInitialStakeValue = errors.New("initial stake value is negative")

// ErrInvalidMinStakeValue signals that an invalid min stake value was provided
var ErrInvalidMinStakeValue = errors.New("invalid min stake value")

// ErrInvalidNodePrice signals that an invalid node price was provided
var ErrInvalidNodePrice = errors.New("invalid node price")

// ErrInvalidMinStepValue signals that an invalid min step value was provided
var ErrInvalidMinStepValue = errors.New("invalid min step value")

// ErrInvalidMinUnstakeTokensValue signals that an invalid min unstake tokens value was provided
var ErrInvalidMinUnstakeTokensValue = errors.New("invalid min unstake tokens value")

// ErrBLSPublicKeyMismatch signals that public keys do not match
var ErrBLSPublicKeyMismatch = errors.New("public key mismatch")

// ErrBLSPublicKeyMissmatch signals that public keys do not match
var ErrBLSPublicKeyMissmatch = errors.New("public key missmatch")

// ErrKeyAlreadyRegistered signals that bls key is already registered
var ErrKeyAlreadyRegistered = errors.New("bls key already registered")

// ErrNotEnoughArgumentsToStake signals that the arguments provided are not enough
var ErrNotEnoughArgumentsToStake = errors.New("not enough arguments to stake")

// ErrNilKeyGenerator signals that key generator is nil
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrNilSingleSigner signals that the single signer is nil
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrIncorrectConfig signals that the config is incorrect
var ErrIncorrectConfig = errors.New("config incorrect")

// ErrNilMessageSignVerifier signals that message sign verifier is nil
var ErrNilMessageSignVerifier = errors.New("nil message sign verifier")

// ErrNilStakingSmartContractAddress signals that staking smart contract address is nil
var ErrNilStakingSmartContractAddress = errors.New("nil staking smart contract address")

// ErrNilEndOfEpochSmartContractAddress signals that the end of epoch smart contract address is nil
var ErrNilEndOfEpochSmartContractAddress = errors.New("nil end of epoch smart contract address")

// ErrNilArgumentsParser signals that arguments parses is nil
var ErrNilArgumentsParser = errors.New("nil arguments parser")

// ErrNilValidatorSmartContractAddress signals that validator smart contract address is nil
var ErrNilValidatorSmartContractAddress = errors.New("nil validator smart contract address")

// ErrInvalidStakingAccessAddress signals that invalid staking access address was provided
var ErrInvalidStakingAccessAddress = errors.New("invalid staking access address")

// ErrInvalidJailAccessAddress signals that invalid jailing access address was provided
var ErrInvalidJailAccessAddress = errors.New("invalid jailing access address")

// ErrNotEnoughGas signals that there is not enough gas for execution
var ErrNotEnoughGas = errors.New("not enough gas")

// ErrNilNodesConfigProvider signals that an operation has been attempted to or with a nil nodes config provider
var ErrNilNodesConfigProvider = errors.New("nil nodes config provider")

// ErrInvalidBaseIssuingCost signals that invalid base issuing cost has been provided
var ErrInvalidBaseIssuingCost = errors.New("invalid base issuing cost")

// ErrInvalidMinCreationDeposit signals that invalid min creation deposit has been provided
var ErrInvalidMinCreationDeposit = errors.New("invalid min creation deposit")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil Hasher")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNegativeOrZeroInitialSupply signals that negative initial supply has been provided
var ErrNegativeOrZeroInitialSupply = errors.New("negative initial supply was provided")

// ErrInvalidNumberOfDecimals signals that an invalid number of decimals has been provided
var ErrInvalidNumberOfDecimals = errors.New("invalid number of decimals")

// ErrNilSystemSCConfig signals that nil system sc config was provided
var ErrNilSystemSCConfig = errors.New("nil system sc config")

// ErrNilValidatorAccountsDB signals that nil validator accounts DB was provided
var ErrNilValidatorAccountsDB = errors.New("nil validator accounts DB")

// ErrInvalidStartEndVoteNonce signals that invalid arguments where passed for start or end vote nonce
var ErrInvalidStartEndVoteNonce = errors.New("invalid start/end vote nonce")

// ErrEmptyStorage signals that the storage is empty for given key
var ErrEmptyStorage = errors.New("storage is nil for given key")

// ErrVotedForAnExpiredProposal signals that voting was done for an expired proposal
var ErrVotedForAnExpiredProposal = errors.New("voting period is over for this proposal")

// ErrVotingNotStartedForProposal signals that voting was done for a proposal that not begins yet
var ErrVotingNotStartedForProposal = errors.New("voting has not yet started for this proposal")

// ErrNilPublicKey signals that nil public key has been provided
var ErrNilPublicKey = errors.New("nil public key")

// ErrInvalidWaitingList signals that waiting list is invalid
var ErrInvalidWaitingList = errors.New("invalid waiting list")

// ErrElementNotFound signals that element was not found
var ErrElementNotFound = errors.New("element was not found")

// ErrInvalidUnJailCost signals that provided unjail cost is invalid
var ErrInvalidUnJailCost = errors.New("invalid unjail cost")

// ErrInvalidGenesisTotalSupply signals that provided genesis total supply is invalid
var ErrInvalidGenesisTotalSupply = errors.New("invalid genesis total supply cost")

// ErrNegativeBleedPercentagePerRound signals that negative bleed percentage per round has been provided
var ErrNegativeBleedPercentagePerRound = errors.New("negative bleed percentage per round")

// ErrNegativeMaximumPercentageToBleed signals that negative maximum percentage to bleed has been provided
var ErrNegativeMaximumPercentageToBleed = errors.New("negative maximum percentage to bleed")

// ErrInvalidMaxNumberOfNodes signals that invalid number of max number of nodes has been provided
var ErrInvalidMaxNumberOfNodes = errors.New("invalid number of max number of nodes")

// ErrTokenNameNotHumanReadable signals that token name is not human-readable
var ErrTokenNameNotHumanReadable = errors.New("token name is not human readable")

// ErrTickerNameNotValid signals that ticker name is not valid
var ErrTickerNameNotValid = errors.New("ticker name is not valid")

// ErrCouldNotCreateNewTokenIdentifier signals that token identifier could not be created
var ErrCouldNotCreateNewTokenIdentifier = errors.New("token identifier could not be created")

// ErrBLSPublicKeyAlreadyJailed signals that bls public key was already jailed
var ErrBLSPublicKeyAlreadyJailed = errors.New("bls public key already jailed")

// ErrInvalidEndOfEpochAccessAddress signals that end of epoch access address is invalid
var ErrInvalidEndOfEpochAccessAddress = errors.New("invalid end of epoch access address")

// ErrNilChanceComputer signals that nil chance computer has been provided
var ErrNilChanceComputer = errors.New("nil chance computer")

// ErrNilAddressPubKeyConverter signals that the provided public key converter is nil
var ErrNilAddressPubKeyConverter = errors.New("nil address public key converter")

// ErrNoTickerWithGivenName signals that ticker does not exist with given name
var ErrNoTickerWithGivenName = errors.New("no ticker with given name")

// ErrInvalidAddress signals that invalid address has been provided
var ErrInvalidAddress = errors.New("invalid address")

// ErrDataNotFoundUnderKey signals that data was not found under requested key
var ErrDataNotFoundUnderKey = errors.New("data was not found under requested key")

// ErrInvalidBLSKeys signals that invalid bls keys has been provided
var ErrInvalidBLSKeys = errors.New("invalid bls keys")

// ErrInvalidNumOfArguments signals that invalid number of arguments has been provided
var ErrInvalidNumOfArguments = errors.New("invalid number of arguments")

// ErrInvalidArgument signals that invalid argument has been provided
var ErrInvalidArgument = errors.New("invalid argument")

// ErrNilGasSchedule signals that nil gas schedule has been provided
var ErrNilGasSchedule = errors.New("nil gas schedule")

// ErrDuplicatesFoundInArguments signals that duplicates were found in arguments
var ErrDuplicatesFoundInArguments = errors.New("duplicates found in arguments")

// ErrInvalidCaller signals that the functions was called by a not authorized user
var ErrInvalidCaller = errors.New("the function was called by a not authorized user")

// ErrCallValueMustBeZero signals that call value must be zero
var ErrCallValueMustBeZero = errors.New("call value must be zero")

// ErrInvalidDelegationSCConfig signals that invalid delegation sc config has been provided
var ErrInvalidDelegationSCConfig = errors.New("invalid delegation sc config")

// ErrOwnerCannotUnDelegate signals that owner cannot undelegate as contract is still active
var ErrOwnerCannotUnDelegate = errors.New("owner cannot undelegate, contract still active")

// ErrNotEnoughInitialOwnerFunds signals that not enough initial owner funds has been provided
var ErrNotEnoughInitialOwnerFunds = errors.New("not enough initial owner funds")

// ErrNFTCreateRoleAlreadyExists signals that NFT create role already exists
var ErrNFTCreateRoleAlreadyExists = errors.New("NFT create role already exists")

// ErrRedelegateValueBelowMinimum signals that the re-delegate added to the remaining value will be below the minimum required
var ErrRedelegateValueBelowMinimum = errors.New("can not re-delegate as the remaining value will be below the minimum required")

// ErrWrongRewardAddress signals that a wrong reward address was provided
var ErrWrongRewardAddress = errors.New("wrong reward address")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrProposalNotFound signals that the storage is empty for given key
var ErrProposalNotFound = errors.New("proposal was not found in storage")

// ErrInvalidNumOfInitialWhiteListedAddress signals that 0 initial whiteListed addresses were provided to the governance contract
var ErrInvalidNumOfInitialWhiteListedAddress = errors.New("0 initial whiteListed addresses provided to the governance contract")

// ErrNilEnableEpochsHandler signals that a nil enable epochs handler has been provided
var ErrNilEnableEpochsHandler = errors.New("nil enable epochs handler")
