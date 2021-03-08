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
	payableHandler   process.PayableHandler
	funcGasCost      uint64
	accounts         state.AccountsAdapter
	shardCoordinator sharding.Coordinator
	gasConfig        process.BaseOperationCost
	mutExecution     sync.RWMutex
}

// NewESDTNFTTransferFunc returns the esdt nft add quantity built-in function component
func NewESDTNFTTransferFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
	pauseHandler process.ESDTPauseHandler,
	rolesHandler process.ESDTRoleHandler,
	payableHandler process.PayableHandler,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	gasConfig process.BaseOperationCost,
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
		payableHandler:   payableHandler,
		gasConfig:        gasConfig,
		mutExecution:     sync.RWMutex{},
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtNFTTransfer) SetNewGasConfig(gasCost *process.GasCost) {
	e.mutExecution.Lock()
	e.funcGasCost = gasCost.BuiltInCost.ESDTTransfer
	e.gasConfig = gasCost.BaseOperationCost
	e.mutExecution.Unlock()
}

// ProcessBuiltinFunction resolves ESDT NFT transfer roles function call
func (e *esdtNFTTransfer) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
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

	if bytes.Equal(vmInput.CallerAddr, vmInput.RecipientAddr) {
		return e.processNFTTransferOnSenderShard(acntSnd, vmInput)
	}

	// in cross shard NFT transfer the sender account must be nil
	if !check.IfNil(acntSnd) {
		return nil, process.ErrInvalidRcvAddr
	}
	if check.IfNil(acntDst) {
		return nil, process.ErrInvalidRcvAddr
	}

	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	marshalledNFTTransfer := vmInput.Arguments[3]
	esdtTransferData := &esdt.ESDigitalToken{}
	err = e.marshalizer.Unmarshal(esdtTransferData, marshalledNFTTransfer)
	if err != nil {
		return nil, err
	}

	mustVerifyPayable := vmInput.CallType != vmcommon.AsynchronousCallBack && len(vmInput.Arguments) == 4
	err = e.addNFTToDestination(vmInput.RecipientAddr, esdtTransferData, esdtTokenKey, mustVerifyPayable)
	if err != nil {
		return nil, err
	}

	// no need to consume gas on destination - sender already paid for it
	vmOutput := &vmcommon.VMOutput{GasRemaining: vmInput.GasProvided}
	if len(vmInput.Arguments) > 4 && core.IsSmartContractAddress(vmInput.RecipientAddr) {
		var callArgs [][]byte
		if len(vmInput.Arguments) > 5 {
			callArgs = vmInput.Arguments[5:]
		}

		addOutPutTransferToVMOutput(
			string(vmInput.Arguments[4]),
			callArgs,
			vmInput.RecipientAddr,
			vmInput.GasLocked,
			vmOutput)
	}

	return vmOutput, nil
}

func (e *esdtNFTTransfer) processNFTTransferOnSenderShard(
	acntSnd state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	dstAddress := vmInput.Arguments[3]
	if len(dstAddress) != len(vmInput.CallerAddr) {
		return nil, process.ErrInvalidArguments
	}
	if bytes.Equal(dstAddress, vmInput.CallerAddr) {
		return nil, process.ErrInvalidArguments
	}

	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
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
	mustVerifyPayable := vmInput.CallType != vmcommon.AsynchronousCallBack && len(vmInput.Arguments) == 4
	err = e.addNFTToDestination(dstAddress, esdtData, esdtTokenKey, mustVerifyPayable)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{
		ReturnCode:   vmcommon.Ok,
		GasRemaining: vmInput.GasProvided - e.funcGasCost,
	}
	err = e.createNFTOutputTransfers(vmInput, vmOutput, esdtData, dstAddress)
	if err != nil {
		return nil, err
	}

	return vmOutput, nil
}

func (e *esdtNFTTransfer) createNFTOutputTransfers(
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
	esdtTransferData *esdt.ESDigitalToken,
	dstAddress []byte,
) error {
	marshalledNFTTransfer, err := e.marshalizer.Marshal(esdtTransferData)
	if err != nil {
		return err
	}

	gasForTransfer := uint64(len(marshalledNFTTransfer)) * e.gasConfig.DataCopyPerByte
	if gasForTransfer > vmOutput.GasRemaining {
		return process.ErrNotEnoughGas
	}
	vmOutput.GasRemaining = vmOutput.GasRemaining - gasForTransfer

	nftTransferCallArgs := make([][]byte, 0)
	nftTransferCallArgs = append(nftTransferCallArgs, vmInput.Arguments[:2]...)
	nftTransferCallArgs = append(nftTransferCallArgs, marshalledNFTTransfer)
	if len(vmInput.Arguments) > 4 {
		nftTransferCallArgs = append(nftTransferCallArgs, vmInput.Arguments[4:]...)
	}

	isSCCallAfter := len(vmInput.Arguments) > 4 && core.IsSmartContractAddress(dstAddress)

	if e.shardCoordinator.SelfId() != e.shardCoordinator.ComputeId(dstAddress) {
		gasToTransfer := uint64(0)
		if isSCCallAfter {
			gasToTransfer = vmOutput.GasRemaining
			vmOutput.GasRemaining = 0
		}
		addNFTTransferToVMOutput(
			dstAddress,
			nftTransferCallArgs,
			vmInput.GasLocked,
			gasToTransfer,
			vmOutput)

		return nil
	}

	addNFTTransferToVMOutput(
		dstAddress,
		nftTransferCallArgs,
		0,
		0,
		vmOutput)

	if isSCCallAfter {
		var callArgs [][]byte
		if len(vmInput.Arguments) > 5 {
			callArgs = vmInput.Arguments[5:]
		}

		addOutPutTransferToVMOutput(
			string(vmInput.Arguments[4]),
			callArgs,
			vmInput.RecipientAddr,
			vmInput.GasLocked,
			vmOutput)
	}

	return nil
}

func (e *esdtNFTTransfer) addNFTToDestination(
	dstAddress []byte,
	esdtDataToTransfer *esdt.ESDigitalToken,
	esdtTokenKey []byte,
	mustVerifyPayable bool,
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

	if mustVerifyPayable {
		isPayable, errIsPayable := e.payableHandler.IsPayable(dstAddress)
		if errIsPayable != nil {
			return errIsPayable
		}
		if !isPayable {
			return process.ErrAccountNotPayable
		}
	}

	currentESDTData, err := getESDTNFTToken(userAccount, esdtTokenKey, esdtDataToTransfer.TokenMetaData.Nonce, e.marshalizer)
	if err != nil && !errors.Is(err, process.ErrNFTTokenDoesNotExist) {
		return err
	}
	if currentESDTData != nil {
		if currentESDTData.TokenMetaData == nil {
			return process.ErrWrongNFTOnDestination
		}
		if !bytes.Equal(currentESDTData.TokenMetaData.Hash, esdtDataToTransfer.TokenMetaData.Hash) {
			return process.ErrWrongNFTOnDestination
		}
		esdtDataToTransfer.Value.Add(esdtDataToTransfer.Value, currentESDTData.Value)
	}

	err = saveESDTNFTToken(userAccount, esdtTokenKey, esdtDataToTransfer, e.marshalizer, e.pauseHandler)
	if err != nil {
		return err
	}

	err = e.accounts.SaveAccount(userAccount)
	if err != nil {
		return err
	}

	return nil
}

func addNFTTransferToVMOutput(
	recipient []byte,
	arguments [][]byte,
	gasLocked uint64,
	gasLimit uint64,
	vmOutput *vmcommon.VMOutput,
) {
	nftTransferTxData := core.BuiltInFunctionESDTNFTTransfer
	for _, arg := range arguments {
		nftTransferTxData += "@" + hex.EncodeToString(arg)
	}
	outTransfer := vmcommon.OutputTransfer{
		Value:     big.NewInt(0),
		GasLimit:  gasLimit,
		GasLocked: gasLocked,
		Data:      []byte(nftTransferTxData),
		CallType:  vmcommon.AsynchronousCall,
	}
	vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	vmOutput.OutputAccounts[string(recipient)] = &vmcommon.OutputAccount{
		Address:         recipient,
		OutputTransfers: []vmcommon.OutputTransfer{outTransfer},
	}
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtNFTTransfer) IsInterfaceNil() bool {
	return e == nil
}
