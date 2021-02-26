package builtInFunctions

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.BuiltinFunction = (*esdtLocalMint)(nil)

type esdtLocalMint struct {
	keyPrefix    []byte
	marshalizer  marshal.Marshalizer
	pauseHandler process.ESDTPauseHandler
	rolesHandler process.ESDTRoleHandler
	funcGasCost  uint64
	mutExecution sync.RWMutex
}

// NewESDTLocalMintFunc returns the esdt local mint built-in function component
func NewESDTLocalMintFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
	pauseHandler process.ESDTPauseHandler,
	rolesHandler process.ESDTRoleHandler,
) (*esdtLocalMint, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(pauseHandler) {
		return nil, process.ErrNilPauseHandler
	}
	if check.IfNil(rolesHandler) {
		return nil, process.ErrNilRolesHandler
	}

	e := &esdtLocalMint{
		keyPrefix:    []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier),
		marshalizer:  marshalizer,
		pauseHandler: pauseHandler,
		rolesHandler: rolesHandler,
		funcGasCost:  funcGasCost,
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtLocalMint) SetNewGasConfig(gasCost *process.GasCost) {
	e.mutExecution.Lock()
	e.funcGasCost = gasCost.BuiltInCost.ESDTTransfer
	e.mutExecution.Unlock()
}

// ProcessBuiltinFunction resolves ESDT change roles function call
func (e *esdtLocalMint) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	e.mutExecution.RLock()
	defer e.mutExecution.RUnlock()

	err := checkInputArgumentsForLocalAction(acntSnd, acntDst, vmInput, e.funcGasCost)
	if err != nil {
		return nil, err
	}

	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	err = e.rolesHandler.CheckAllowedToExecute(acntSnd, esdtTokenKey, []byte(core.ESDTRoleLocalMint))
	if err != nil {
		return nil, err
	}

	value := big.NewInt(0).SetBytes(vmInput.Arguments[1])
	err = addToESDTBalance(vmInput.CallerAddr, acntSnd, esdtTokenKey, big.NewInt(0).Set(value), e.marshalizer, e.pauseHandler)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{ReturnCode: vmcommon.Ok, GasRemaining: vmInput.GasProvided - e.funcGasCost}
	return vmOutput, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtLocalMint) IsInterfaceNil() bool {
	return e == nil
}
