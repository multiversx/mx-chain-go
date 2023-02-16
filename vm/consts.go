package vm

const (
	// InsufficientGasLimit defined constant for return message
	InsufficientGasLimit = "insufficient gas limit"
	// StakeNotEnabled defined constant for return message
	StakeNotEnabled = "stake is not enabled"
	// UnBondNotEnabled defined constant for return message
	UnBondNotEnabled = "unBond is not enabled"
	// UnStakeNotEnabled defined constant for return message
	UnStakeNotEnabled = "unStake is not enabled"
	// TransactionValueMustBeZero defined constant for return message
	TransactionValueMustBeZero = "transaction value must be zero"
	// CannotGetOrCreateRegistrationData defined constant for return message
	CannotGetOrCreateRegistrationData = "cannot get or create registration data: error - "
	// CannotGetAllBlsKeysFromRegistrationData defined constant for return message
	CannotGetAllBlsKeysFromRegistrationData = "could not get all blsKeys from registration data: error - "

	// CreateNewDelegationContract defined constant for delegationManager Execute function
	CreateNewDelegationContract = "createNewDelegationContract"
	// GetAllContractAddresses defined constant for delegationManager Execute function
	GetAllContractAddresses = "getAllContractAddresses"
	// GetContractConfig defined constant for delegationManager Execute function
	GetContractConfig = "getContractConfig"
	// ChangeMinDeposit defined constant for delegationManager Execute function
	ChangeMinDeposit = "changeMinDeposit"
	// ChangeMinDelegationAmount defined constant for delegationManager Execute function
	ChangeMinDelegationAmount = "changeMinDelegationAmount"
	// MakeNewContractFromValidatorData defined constant for delegationManager Execute function
	MakeNewContractFromValidatorData = "makeNewContractFromValidatorData"
	// MergeValidatorToDelegationSameOwner defined constant for delegationManager Execute function
	MergeValidatorToDelegationSameOwner = "mergeValidatorToDelegationSameOwner"
	// MergeValidatorToDelegationWithWhitelist defined constant for delegationManager Execute function
	MergeValidatorToDelegationWithWhitelist = "mergeValidatorToDelegationWithWhitelist"
)
