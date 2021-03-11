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

var _ process.BuiltinFunction = (*esdtNFTAddQuantity)(nil)

type esdtNFTAddQuantity struct {
	keyPrefix    []byte
	marshalizer  marshal.Marshalizer
	pauseHandler process.ESDTPauseHandler
	rolesHandler process.ESDTRoleHandler
	funcGasCost  uint64
	mutExecution sync.RWMutex
}

// NewESDTNFTAddQuantityFunc returns the esdt NFT add quantity built-in function component
func NewESDTNFTAddQuantityFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
	pauseHandler process.ESDTPauseHandler,
	rolesHandler process.ESDTRoleHandler,
) (*esdtNFTAddQuantity, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(pauseHandler) {
		return nil, process.ErrNilPauseHandler
	}
	if check.IfNil(rolesHandler) {
		return nil, process.ErrNilRolesHandler
	}

	e := &esdtNFTAddQuantity{
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
func (e *esdtNFTAddQuantity) SetNewGasConfig(gasCost *process.GasCost) {
	if gasCost == nil {
		return
	}

	e.mutExecution.Lock()
	e.funcGasCost = gasCost.BuiltInCost.ESDTNFTAddQuantity
	e.mutExecution.Unlock()
}

// ProcessBuiltinFunction resolves ESDT NFT add quantity function call
func (e *esdtNFTAddQuantity) ProcessBuiltinFunction(
	acntSnd, _ state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	e.mutExecution.RLock()
	defer e.mutExecution.RUnlock()

	err := checkESDTNFTCreateBurnAddInput(acntSnd, vmInput, e.funcGasCost)
	if err != nil {
		return nil, err
	}
	if len(vmInput.Arguments) < 3 {
		return nil, process.ErrInvalidArguments
	}

	err = e.rolesHandler.CheckAllowedToExecute(acntSnd, vmInput.Arguments[0], []byte(core.ESDTRoleNFTAddQuantity))
	if err != nil {
		return nil, err
	}

	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	nonce := big.NewInt(0).SetBytes(vmInput.Arguments[1]).Uint64()
	esdtData, err := getESDTNFTToken(acntSnd, esdtTokenKey, nonce, e.marshalizer)
	if err != nil {
		return nil, err
	}

	esdtData.Value.Add(esdtData.Value, big.NewInt(0).SetBytes(vmInput.Arguments[2]))

	err = saveESDTNFTToken(acntSnd, esdtTokenKey, esdtData, e.marshalizer, e.pauseHandler)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{
		ReturnCode:   vmcommon.Ok,
		GasRemaining: vmInput.GasProvided - e.funcGasCost,
	}
	return vmOutput, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtNFTAddQuantity) IsInterfaceNil() bool {
	return e == nil
}
