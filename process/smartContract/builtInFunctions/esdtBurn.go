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
	"github.com/ElrondNetwork/elrond-go/vm"
)

var _ process.BuiltinFunction = (*esdtBurn)(nil)

type esdtBurn struct {
	funcGasCost  uint64
	marshalizer  marshal.Marshalizer
	keyPrefix    []byte
	pauseHandler process.ESDTPauseHandler
	mutExecution sync.RWMutex
}

// NewESDTBurnFunc returns the esdt burn built-in function component
func NewESDTBurnFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
	pauseHandler process.ESDTPauseHandler,
) (*esdtBurn, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(pauseHandler) {
		return nil, process.ErrNilPauseHandler
	}

	e := &esdtBurn{
		funcGasCost:  funcGasCost,
		marshalizer:  marshalizer,
		keyPrefix:    []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier),
		pauseHandler: pauseHandler,
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtBurn) SetNewGasConfig(gasCost *process.GasCost) {
	e.mutExecution.Lock()
	e.funcGasCost = gasCost.BuiltInCost.ESDTBurn
	e.mutExecution.Unlock()
}

// ProcessBuiltinFunction resolves ESDT burn function call
func (e *esdtBurn) ProcessBuiltinFunction(
	acntSnd, _ state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	e.mutExecution.RLock()
	defer e.mutExecution.RUnlock()

	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return nil, process.ErrBuiltInFunctionCalledWithValue
	}
	if len(vmInput.Arguments) != 2 {
		return nil, process.ErrInvalidArguments
	}
	value := big.NewInt(0).SetBytes(vmInput.Arguments[1])
	if value.Cmp(zero) <= 0 {
		return nil, process.ErrNegativeValue
	}
	if !bytes.Equal(vmInput.RecipientAddr, vm.ESDTSCAddress) {
		return nil, process.ErrAddressIsNotESDTSystemSC
	}
	if check.IfNil(acntSnd) {
		return nil, process.ErrNilUserAccount
	}

	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	log.Trace("esdtBurn", "sender", vmInput.CallerAddr, "receiver", vmInput.RecipientAddr, "value", value, "token", esdtTokenKey)

	if vmInput.GasProvided < e.funcGasCost {
		return nil, process.ErrNotEnoughGas
	}

	err := addToESDTBalance(vmInput.CallerAddr, acntSnd, esdtTokenKey, big.NewInt(0).Neg(value), e.marshalizer, e.pauseHandler)
	if err != nil {
		return nil, err
	}

	gasRemaining := computeGasRemaining(acntSnd, vmInput.GasProvided, e.funcGasCost)
	vmOutput := &vmcommon.VMOutput{GasRemaining: gasRemaining, ReturnCode: vmcommon.Ok}
	if core.IsSmartContractAddress(vmInput.CallerAddr) {
		addOutPutTransferToVMOutput(
			core.BuiltInFunctionESDTBurn,
			vmInput.Arguments,
			vmInput.RecipientAddr,
			vmInput.GasLocked,
			vmOutput)
	}

	return vmOutput, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtBurn) IsInterfaceNil() bool {
	return e == nil
}
