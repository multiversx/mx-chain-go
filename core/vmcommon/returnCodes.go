package vmcommon

import "fmt"

// ReturnCode is an enum with the possible error codes returned by the VM
type ReturnCode int

func (rc ReturnCode) String() string {
	switch rc {
	case Ok:
		return "ok"
	case FunctionNotFound:
		return "function not found"
	case FunctionWrongSignature:
		return "wrong signature for function"
	case ContractNotFound:
		return "contract not found"
	case UserError:
		return "user error"
	case OutOfGas:
		return "out of gas"
	case AccountCollision:
		return "account collision"
	case OutOfFunds:
		return "out of funds"
	case CallStackOverFlow:
		return "call stack overflow"
	case ContractInvalid:
		return "contract invalid"
	case ExecutionFailed:
		return "execution failed"
	default:
		return fmt.Sprintf("unknown error, code: %d", rc)
	}
}

const (
	// Ok is returned when execution was completed normally.
	Ok ReturnCode = 0

	// FunctionNotFound is returned when the input specifies a function name that does not exist or is not public.
	FunctionNotFound ReturnCode = 1

	// FunctionWrongSignature is returned when the wrong number of arguments is provided.
	FunctionWrongSignature ReturnCode = 2

	// ContractNotFound is returned when the called contract does not exist.
	ContractNotFound ReturnCode = 3

	// UserError is returned for various execution errors.
	UserError ReturnCode = 4

	// OutOfGas is returned when VM execution runs out of gas.
	OutOfGas ReturnCode = 5

	// AccountCollision is returned when created account already exists.
	AccountCollision ReturnCode = 6

	// OutOfFunds is returned when the caller (sender) runs out of funds.
	OutOfFunds ReturnCode = 7

	// CallStackOverFlow is returned when stack overflow occurs.
	CallStackOverFlow ReturnCode = 8

	// ContractInvalid is returned when the contract is invalid.
	ContractInvalid ReturnCode = 9

	// ExecutionFailed is returned when the execution of the specified function has failed.
	ExecutionFailed ReturnCode = 10

	// UpgradeFailed is returned when the upgrade of the contract has failed
	UpgradeFailed ReturnCode = 11
)
