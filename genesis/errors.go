package genesis

import "errors"

// ErrNilEntireSupply signals that the provided entire supply is nil
var ErrNilEntireSupply = errors.New("nil entire supply")

// ErrInvalidEntireSupply signals that the provided entire supply is invalid
var ErrInvalidEntireSupply = errors.New("invalid entire supply")

// ErrEntireSupplyMismatch signals that the provided entire supply mismatches the computed one
var ErrEntireSupplyMismatch = errors.New("entire supply mismatch")

// ErrEmptyAddress signals that an empty address was found in genesis file
var ErrEmptyAddress = errors.New("empty address")

// ErrInvalidAddress signals that an invalid address was found
var ErrInvalidAddress = errors.New("invalid address")

// ErrInvalidPubKey signals that an invalid public key has been provided
var ErrInvalidPubKey = errors.New("invalid public key")

// ErrInvalidSupplyString signals that the supply string is not a valid number
var ErrInvalidSupplyString = errors.New("invalid supply string")

// ErrInvalidBalanceString signals that the balance string is not a valid number
var ErrInvalidBalanceString = errors.New("invalid balance string")

// ErrInvalidStakingBalanceString signals that the staking balance string is not a valid number
var ErrInvalidStakingBalanceString = errors.New("invalid staking balance string")

// ErrInvalidDelegationValueString signals that the delegation balance string is not a valid number
var ErrInvalidDelegationValueString = errors.New("invalid delegation value string")

// ErrEmptyDelegationAddress signals that the delegation address is empty
var ErrEmptyDelegationAddress = errors.New("empty delegation address")

// ErrInvalidDelegationAddress signals that the delegation address is invalid
var ErrInvalidDelegationAddress = errors.New("invalid delegation address")

// ErrInvalidSupply signals that the supply field is invalid
var ErrInvalidSupply = errors.New("invalid supply")

// ErrInvalidBalance signals that the balance field is invalid
var ErrInvalidBalance = errors.New("invalid balance")

// ErrInvalidStakingBalance signals that the staking balance field is invalid
var ErrInvalidStakingBalance = errors.New("invalid staking balance")

// ErrInvalidDelegationValue signals that the delegation value field is invalid
var ErrInvalidDelegationValue = errors.New("invalid delegation value")

// ErrSupplyMismatch signals that the supply value provided is not valid when summing the other fields
var ErrSupplyMismatch = errors.New("supply value mismatch")

// ErrDuplicateAddress signals that the same address was found more than one time
var ErrDuplicateAddress = errors.New("duplicate address")

// ErrAddressIsSmartContract signals that provided address is of type smart contract
var ErrAddressIsSmartContract = errors.New("address is a smart contract")

// ErrNilShardCoordinator signals that the provided shard coordinator is nil
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilPubkeyConverter signals that the provided public key converter is nil
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilAccountsParser signals that the provided accounts parser is nil
var ErrNilAccountsParser = errors.New("nil accounts parser")

// ErrStakingValueIsNotEnough signals that the staking value provided is not enough for provided node(s)
var ErrStakingValueIsNotEnough = errors.New("staking value is not enough")

// ErrDelegationValueIsNotEnough signals that the delegation value provided is not enough for provided node(s)
var ErrDelegationValueIsNotEnough = errors.New("delegation value is not enough")

// ErrNodeNotStaked signals that no one staked for the provided node
var ErrNodeNotStaked = errors.New("for the provided node, no one staked")

// ErrNilInitialNodePrice signals that the provided initial node price is nil
var ErrNilInitialNodePrice = errors.New("nil initial node price")

// ErrInvalidInitialNodePrice signals that the provided initial node price is invalid
var ErrInvalidInitialNodePrice = errors.New("invalid initial node price")

// ErrNilDelegationHandler signals that a nil delegation handler has been used
var ErrNilDelegationHandler = errors.New("nil delegation handler")

// ErrEmptyOwnerAddress signals that an empty owner address was found in genesis file
var ErrEmptyOwnerAddress = errors.New("empty owner address")

// ErrInvalidOwnerAddress signals that an invalid owner address was found
var ErrInvalidOwnerAddress = errors.New("invalid owner address")

// ErrFilenameIsDirectory signals that the provided filename is a directory
var ErrFilenameIsDirectory = errors.New("provided name is actually a directory")

// ErrNilSmartContractParser signals that the smart contract parser is nil
var ErrNilSmartContractParser = errors.New("nil smart contract parser")

// ErrInvalidVmType signals that the provided VM type is invalid
var ErrInvalidVmType = errors.New("invalid vm type")

// ErrEmptyVmType signals that the provided VM type is empty
var ErrEmptyVmType = errors.New("empty vm type")

// ErrWrongTypeAssertion signals that a wrong type assertion occurred
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilTxExecutionProcessor signals that a nil tx execution processor has been provided
var ErrNilTxExecutionProcessor = errors.New("nil tx execution processor")

// ErrNilNodesListSplitter signals that a nil nodes handler has been provided
var ErrNilNodesListSplitter = errors.New("nil nodes list splitter")

// ErrNilNodesSetup signals that a nil nodes setup handler has been provided
var ErrNilNodesSetup = errors.New("nil nodes setup")

// ErrAccountAlreadyExists signals that an account already exists
var ErrAccountAlreadyExists = errors.New("account already exists")

// ErrAccountNotCreated signals that an account could not have been created
var ErrAccountNotCreated = errors.New("account not created")

// ErrNilTrieStorageManager signals that a nil trie storage manager has been provided
var ErrNilTrieStorageManager = errors.New("nil trie storage manager")

// ErrWhileVerifyingDelegation signals that a verification error occurred
var ErrWhileVerifyingDelegation = errors.New("error occurred while verifying delegation SC")

// ErrNilQueryService signals that a nil query service has been provided
var ErrNilQueryService = errors.New("nil query service")

// ErrEmptyReturnData signals an empty return data from vmOutput was received
var ErrEmptyReturnData = errors.New("empty return data")

// ErrSignatureMismatch signals a signature mismatch occurred
var ErrSignatureMismatch = errors.New("signature mismatch")

// ErrGetVersionFromSC signals that a call to "version" function on a contract resulted in an unexpected result
var ErrGetVersionFromSC = errors.New("get version from contract returned an invalid response")

// ErrEmptyPubKey signals that empty public key has been provided
var ErrEmptyPubKey = errors.New("empty public key")

// ErrNilKeyGenerator signals that nil key generator has been provided
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrChangeOwnerAddressFailed signals that change owner address failed
var ErrChangeOwnerAddressFailed = errors.New("change owner address failed")

// ErrTooManyDNSContracts signals that too many DNS contracts are in genesis contracts json file
var ErrTooManyDNSContracts = errors.New("too many DNS contracts")

// ErrSmartContractWasNotDeployed signals that smart contract was not deployed
var ErrSmartContractWasNotDeployed = errors.New("smart contract was not deployed")

// ErrBLSKeyNotStaked signals that bls staking was not successful
var ErrBLSKeyNotStaked = errors.New("bls key not staked")
