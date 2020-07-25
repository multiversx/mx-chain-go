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

// ErrNotEnoughQualifiedNodes signals that there are insufficient number of qualified nodes
var ErrNotEnoughQualifiedNodes = errors.New("not enough qualified nodes")

// ErrBLSPublicKeyMissmatch signals that public keys do not match
var ErrBLSPublicKeyMissmatch = errors.New("public key missmatch")

// ErrKeyAlreadyRegistered signals that bls key is already registered
var ErrKeyAlreadyRegistered = errors.New("bls key already registered")

// ErrNotEnoughArgumentsToStake signals that the arguments provided are not enough
var ErrNotEnoughArgumentsToStake = errors.New("not enough arguments to stake")

// ErrNilKeyGenerator signals that key generator is nil
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrSingleSigner signals that single signer is nil
var ErrSingleSigner = errors.New("nil single signer")

// ErrIncorrectConfig signals that auction config is incorrect
var ErrIncorrectConfig = errors.New("config incorrect")

// ErrNilMessageSignVerifier signals that message sign verifier is nil
var ErrNilMessageSignVerifier = errors.New("nil message sign verifier")

// ErrNilStakingSmartContractAddress signals that staking smart contract address is nil
var ErrNilStakingSmartContractAddress = errors.New("nil staking smart contract address")

// ErrNilArgumentsParser signals that arguments parses is nil
var ErrNilArgumentsParser = errors.New("nil arguments parser")

// ErrOnExecutionAtStakingSC signals that there was an error at staking sc call
var ErrOnExecutionAtStakingSC = errors.New("execution error at staking sc")

// ErrNilAuctionSmartContractAddress signals that auction smart contract address is nil
var ErrNilAuctionSmartContractAddress = errors.New("nil auction smart contract address")

// ErrInvalidStakingAccessAddress signals that invalid staking access address was provided
var ErrInvalidStakingAccessAddress = errors.New("invalid staking access address")

// ErrInvalidJailAccessAddress signals that invalid jailing access address was provided
var ErrInvalidJailAccessAddress = errors.New("invalid jailing access address")

// ErrNotEnoughGas signals that there is not enough gas for execution
var ErrNotEnoughGas = errors.New("not enough gas")

// ErrNilNodesConfigProvider signals that an operation has been attempted to or with a nil nodes config provider
var ErrNilNodesConfigProvider = errors.New("nil nodes config provider")

// ErrInvalidMinNumberOfNodes signals that provided minimum number of nodes is invalid
var ErrInvalidMinNumberOfNodes = errors.New("invalid min number of nodes")

// ErrInvalidBaseIssuingCost signals that invalid base issuing cost has been provided
var ErrInvalidBaseIssuingCost = errors.New("invalid base issuing cost")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil Hasher")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNegativeOrZeroInitialSupply signals that negative initial supply has been provided
var ErrNegativeOrZeroInitialSupply = errors.New("negative initial supply was provided")

// ErrTokenAlreadyRegistered signals that token was already registered
var ErrTokenAlreadyRegistered = errors.New("token was already registered")

// ErrNilSystemSCConfig signals that nil system sc config was provided
var ErrNilSystemSCConfig = errors.New("nil system sc config")

// ErrNilValidatorAccountsDB signals that nil validator accounts DB was provided
var ErrNilValidatorAccountsDB = errors.New("nil validator accounts DB")

// ErrInvalidStartEndVoteNonce signals that invalid arguments where passed for start or end vote nonce
var ErrInvalidStartEndVoteNonce = errors.New("invalid start/end vote nonce")

// ErrEmptyStorage signals that the storage is empty for given key
var ErrEmptyStorage = errors.New("storage is nil for given key")

// ErrVotedForAnExpiredProposal signals that voting was done for an expired proposal
var ErrVotedForAnExpiredProposal = errors.New("voted for an expired proposal")

// ErrVotedForAProposalThatNotBeginsYet signals that voting was done for a proposal that not begins yet
var ErrVotedForAProposalThatNotBeginsYet = errors.New("voted for a proposal that not begins yet")

// ErrNilPublicKey signals that nil public key has been provided
var ErrNilPublicKey = errors.New("nil public key")

// ErrInvalidWaitingList signals that waiting list is invalid
var ErrInvalidWaitingList = errors.New("invalid waiting list")

// ErrElementNotFound signals that element was not found
var ErrElementNotFound = errors.New("element was not found")

// ErrInvalidUnJailCost signals that provided unjail cost is invalid
var ErrInvalidUnJailCost = errors.New("invalid unjail cost")

// ErrNegativeWaitingNodesPercentage signals that negative waiting nodes percentage was provided
var ErrNegativeWaitingNodesPercentage = errors.New("negative waiting nodes percentage")

// ErrNegativeBleedPercentagePerRound signals that negative bleed percentage per round has been provided
var ErrNegativeBleedPercentagePerRound = errors.New("negative bleed percentage per round")

// ErrNegativeMaximumPercentageToBleed signals that negative maximum percentage to bleed has been provided
var ErrNegativeMaximumPercentageToBleed = errors.New("negative maximum percentage to bleed")

// ErrInvalidMaxNumberOfNodes signals that invalid number of max number of nodes has been provided
var ErrInvalidMaxNumberOfNodes = errors.New("invalid number of max number of nodes")

// ErrTokenNameNotHumanReadable signals that token name is not human readable
var ErrTokenNameNotHumanReadable = errors.New("token name is not human readable")
