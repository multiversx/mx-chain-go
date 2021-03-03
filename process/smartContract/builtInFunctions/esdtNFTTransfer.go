package builtInFunctions

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.BuiltinFunction = (*esdtNFTTransfer)(nil)

type esdtNFTTransfer struct {
	keyPrefix        []byte
	marshalizer      marshal.Marshalizer
	pauseHandler     process.ESDTPauseHandler
	rolesHandler     process.ESDTRoleHandler
	funcGasCost      uint64
	accounts         state.AccountsAdapter
	shardCoordinator sharding.Coordinator
	mutExecution     sync.RWMutex
}

// NewESDTNFTTransferFunc returns the esdt nft add quantity built-in function component
func NewESDTNFTTransferFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
	pauseHandler process.ESDTPauseHandler,
	rolesHandler process.ESDTRoleHandler,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
) (*esdtNFTTransfer, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(pauseHandler) {
		return nil, process.ErrNilPauseHandler
	}
	if check.IfNil(rolesHandler) {
		return nil, process.ErrNilRolesHandler
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	e := &esdtNFTTransfer{
		keyPrefix:        []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier),
		marshalizer:      marshalizer,
		pauseHandler:     pauseHandler,
		rolesHandler:     rolesHandler,
		funcGasCost:      funcGasCost,
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtNFTTransfer) SetNewGasConfig(gasCost *process.GasCost) {
	e.mutExecution.Lock()
	e.funcGasCost = gasCost.BuiltInCost.ESDTTransfer
	e.mutExecution.Unlock()
}

// ProcessBuiltinFunction resolves ESDT change roles function call
func (e *esdtNFTTransfer) ProcessBuiltinFunction(
	acntSnd, _ state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	e.mutExecution.RLock()
	defer e.mutExecution.RUnlock()

	err := checkBasicESDTArguments(vmInput)
	if err != nil {
		return nil, err
	}
	if len(vmInput.Arguments) < 4 {
		return nil, process.ErrInvalidArguments
	}

	dstAddress := vmInput.Arguments[3]
	if len(dstAddress) != len(vmInput.CallerAddr) {
		return nil, process.ErrInvalidArguments
	}
	if bytes.Equal(dstAddress, vmInput.CallerAddr) {
		return nil, process.ErrInvalidArguments
	}

	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	err = e.rolesHandler.CheckAllowedToExecute(acntSnd, esdtTokenKey, []byte(core.ESDTRoleNFTBurn))
	if err != nil {
		return nil, err
	}

	nonce := big.NewInt(0).SetBytes(vmInput.Arguments[1]).Uint64()
	esdtData, err := getESDTNFTToken(acntSnd, esdtTokenKey, nonce, e.marshalizer)
	if err != nil {
		return nil, err
	}

	quantityToTransfer := big.NewInt(0).SetBytes(vmInput.Arguments[2])
	if esdtData.Value.Cmp(quantityToTransfer) < 0 {
		return nil, process.ErrInvalidNFTQuantity
	}
	esdtData.Value.Sub(esdtData.Value, quantityToTransfer)

	err = saveESDTNFTToken(acntSnd, esdtTokenKey, esdtData, e.marshalizer, e.pauseHandler)
	if err != nil {
		return nil, err
	}

	esdtData.Value.Set(quantityToTransfer)
	err = e.addNFTToDestination(dstAddress, esdtData, esdtTokenKey)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{ReturnCode: vmcommon.Ok, GasRemaining: vmInput.GasProvided - e.funcGasCost}
	addNFTTransferToVMOutput()

	return vmOutput, nil
}

func (e *esdtNFTTransfer) addNFTToDestination(
	dstAddress []byte,
	esdtDataToTransfer *esdt.ESDigitalToken,
	esdtTokenKey []byte,
) error {
	if e.shardCoordinator.SelfId() == e.shardCoordinator.ComputeId(dstAddress) {
		return nil
	}

	accountHandler, err := e.accounts.LoadAccount(dstAddress)
	if err != nil {
		return err
	}

	userAccount, ok := accountHandler.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	currentESDTData, err := getESDTNFTToken(userAccount, esdtTokenKey, esdtDataToTransfer.TokenMetaData.Nonce, e.marshalizer)
	if err != nil && !errors.Is(err, process.ErrNFTTokenDoesNotExist) {
		return err
	}

	if currentESDTData != nil {
		esdtDataToTransfer.Value.Add(esdtDataToTransfer.Value, currentESDTData.Value)
	}

	err = saveESDTNFTToken(userAccount, esdtTokenKey, esdtDataToTransfer, e.marshalizer, e.pauseHandler)
	if err != nil {
		return err
	}

	return nil
}

func addNFTTransferToVMOutput(
	function string,
	arguments [][]byte,
	recipient []byte,
	gasLocked uint64,
	vmOutput *vmcommon.VMOutput,
) {
	esdtTransferTxData := function
	for _, arg := range arguments {
		esdtTransferTxData += "@" + hex.EncodeToString(arg)
	}
	outTransfer := vmcommon.OutputTransfer{
		Value:     big.NewInt(0),
		GasLimit:  vmOutput.GasRemaining,
		GasLocked: gasLocked,
		Data:      []byte(esdtTransferTxData),
		CallType:  vmcommon.AsynchronousCall,
	}
	vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	vmOutput.OutputAccounts[string(recipient)] = &vmcommon.OutputAccount{
		Address:         recipient,
		OutputTransfers: []vmcommon.OutputTransfer{outTransfer},
	}
	vmOutput.GasRemaining = 0
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtNFTTransfer) IsInterfaceNil() bool {
	return e == nil
}
