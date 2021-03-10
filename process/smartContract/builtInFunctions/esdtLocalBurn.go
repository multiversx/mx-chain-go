package builtInFunctions

import (
	"bytes"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.BuiltinFunction = (*esdtLocalBurn)(nil)

type esdtLocalBurn struct {
	keyPrefix    []byte
	marshalizer  marshal.Marshalizer
	pauseHandler process.ESDTPauseHandler
	rolesHandler process.ESDTRoleHandler
	funcGasCost  uint64
	mutExecution sync.RWMutex
}

// NewESDTLocalBurnFunc returns the esdt local burn built-in function component
func NewESDTLocalBurnFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
	pauseHandler process.ESDTPauseHandler,
	rolesHandler process.ESDTRoleHandler,
) (*esdtLocalBurn, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(pauseHandler) {
		return nil, process.ErrNilPauseHandler
	}
	if check.IfNil(rolesHandler) {
		return nil, process.ErrNilRolesHandler
	}

	e := &esdtLocalBurn{
		keyPrefix:    []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier),
		marshalizer:  marshalizer,
		pauseHandler: pauseHandler,
		rolesHandler: rolesHandler,
		funcGasCost:  funcGasCost,
		mutExecution: sync.RWMutex{},
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtLocalBurn) SetNewGasConfig(gasCost *process.GasCost) {
	if gasCost == nil {
		return
	}

	e.mutExecution.Lock()
	e.funcGasCost = gasCost.BuiltInCost.ESDTLocalBurn
	e.mutExecution.Unlock()
}

// ProcessBuiltinFunction resolves ESDT change roles function call
func (e *esdtLocalBurn) ProcessBuiltinFunction(
	acntSnd, _ state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	e.mutExecution.RLock()
	defer e.mutExecution.RUnlock()

	err := checkInputArgumentsForLocalAction(acntSnd, vmInput, e.funcGasCost)
	if err != nil {
		return nil, err
	}

	tokenID := vmInput.Arguments[0]
	err = e.rolesHandler.CheckAllowedToExecute(acntSnd, tokenID, []byte(core.ESDTRoleLocalBurn))
	if err != nil {
		return nil, err
	}

	value := big.NewInt(0).SetBytes(vmInput.Arguments[1])
	esdtTokenKey := append(e.keyPrefix, tokenID...)
	err = addToESDTBalance(vmInput.CallerAddr, acntSnd, esdtTokenKey, big.NewInt(0).Neg(value), e.marshalizer, e.pauseHandler)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{ReturnCode: vmcommon.Ok, GasRemaining: vmInput.GasProvided - e.funcGasCost}
	return vmOutput, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtLocalBurn) IsInterfaceNil() bool {
	return e == nil
}

func checkBasicESDTArguments(vmInput *vmcommon.ContractCallInput) error {
	if vmInput == nil {
		return process.ErrNilVmInput
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return process.ErrBuiltInFunctionCalledWithValue
	}
	if len(vmInput.Arguments) < 2 {
		return process.ErrInvalidArguments
	}
	return nil
}

func checkInputArgumentsForLocalAction(
	acntSnd state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
	funcGasCost uint64,
) error {
	err := checkBasicESDTArguments(vmInput)
	if err != nil {
		return err
	}
	if !bytes.Equal(vmInput.CallerAddr, vmInput.RecipientAddr) {
		return process.ErrInvalidRcvAddr
	}
	if check.IfNil(acntSnd) {
		return process.ErrNilUserAccount
	}
	value := big.NewInt(0).SetBytes(vmInput.Arguments[1])
	if value.Cmp(zero) <= 0 {
		return process.ErrNegativeValue
	}
	if vmInput.GasProvided < funcGasCost {
		return process.ErrNotEnoughGas
	}

	return nil
}
