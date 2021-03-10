package builtInFunctions

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

var _ process.BuiltinFunction = (*esdtTransfer)(nil)

var zero = big.NewInt(0)

type esdtTransfer struct {
	funcGasCost    uint64
	marshalizer    marshal.Marshalizer
	keyPrefix      []byte
	pauseHandler   process.ESDTPauseHandler
	payableHandler process.PayableHandler
	mutExecution   sync.RWMutex
}

// NewESDTTransferFunc returns the esdt transfer built-in function component
func NewESDTTransferFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
	pauseHandler process.ESDTPauseHandler,
) (*esdtTransfer, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(pauseHandler) {
		return nil, process.ErrNilPauseHandler
	}

	e := &esdtTransfer{
		funcGasCost:    funcGasCost,
		marshalizer:    marshalizer,
		keyPrefix:      []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier),
		pauseHandler:   pauseHandler,
		payableHandler: &disabledPayableHandler{},
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtTransfer) SetNewGasConfig(gasCost *process.GasCost) {
	if gasCost == nil {
		return
	}

	e.mutExecution.Lock()
	e.funcGasCost = gasCost.BuiltInCost.ESDTTransfer
	e.mutExecution.Unlock()
}

// ProcessBuiltinFunction resolves ESDT transfer function calls
func (e *esdtTransfer) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	e.mutExecution.RLock()
	defer e.mutExecution.RUnlock()

	err := checkBasicESDTArguments(vmInput)
	if err != nil {
		return nil, err
	}

	value := big.NewInt(0).SetBytes(vmInput.Arguments[1])
	if value.Cmp(zero) <= 0 {
		return nil, process.ErrNegativeValue
	}

	gasRemaining := computeGasRemaining(acntSnd, vmInput.GasProvided, e.funcGasCost)
	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	log.Trace("esdtTransfer", "sender", vmInput.CallerAddr, "receiver", vmInput.RecipientAddr, "value", value, "token", esdtTokenKey)

	if !check.IfNil(acntSnd) {
		// gas is paid only by sender
		if vmInput.GasProvided < e.funcGasCost {
			return nil, process.ErrNotEnoughGas
		}

		err := addToESDTBalance(vmInput.CallerAddr, acntSnd, esdtTokenKey, big.NewInt(0).Neg(value), e.marshalizer, e.pauseHandler)
		if err != nil {
			return nil, err
		}
	}

	isSCCallAfter := core.IsSmartContractAddress(vmInput.RecipientAddr) && len(vmInput.Arguments) > 2

	vmOutput := &vmcommon.VMOutput{GasRemaining: gasRemaining, ReturnCode: vmcommon.Ok}
	if !check.IfNil(acntDst) {
		mustVerifyPayable := vmInput.CallType != vmcommon.AsynchronousCallBack && !bytes.Equal(vmInput.CallerAddr, vm.ESDTSCAddress)
		if mustVerifyPayable && len(vmInput.Arguments) == 2 {
			isPayable, err := e.payableHandler.IsPayable(vmInput.RecipientAddr)
			if err != nil {
				return nil, err
			}
			if !isPayable {
				if !check.IfNil(acntSnd) {
					err = addToESDTBalance(vmInput.CallerAddr, acntSnd, esdtTokenKey, value, e.marshalizer, e.pauseHandler)
					if err != nil {
						return nil, err
					}
				}

				return nil, process.ErrAccountNotPayable
			}
		}

		err = addToESDTBalance(vmInput.CallerAddr, acntDst, esdtTokenKey, value, e.marshalizer, e.pauseHandler)
		if err != nil {
			return nil, err
		}

		if isSCCallAfter {
			vmOutput.GasRemaining, err = core.SafeSubUint64(vmInput.GasProvided, e.funcGasCost)
			log.LogIfError(err, "esdtTransfer", "isSCCallAfter")
			var callArgs [][]byte
			if len(vmInput.Arguments) > 3 {
				callArgs = vmInput.Arguments[3:]
			}

			addOutputTransferToVMOutput(
				string(vmInput.Arguments[2]),
				callArgs,
				vmInput.RecipientAddr,
				vmInput.GasLocked,
				vmOutput)

			return vmOutput, nil
		}

		if vmInput.CallType == vmcommon.AsynchronousCallBack && check.IfNil(acntSnd) {
			// gas was already consumed on sender shard
			vmOutput.GasRemaining = vmInput.GasProvided
		}

		return vmOutput, nil
	}

	// cross-shard ESDT transfer call through a smart contract
	if core.IsSmartContractAddress(vmInput.CallerAddr) {
		addOutputTransferToVMOutput(
			core.BuiltInFunctionESDTTransfer,
			vmInput.Arguments,
			vmInput.RecipientAddr,
			vmInput.GasLocked,
			vmOutput)
	}

	return vmOutput, nil
}

func addOutputTransferToVMOutput(
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

func addToESDTBalance(
	senderAddr []byte,
	userAcnt state.UserAccountHandler,
	key []byte,
	value *big.Int,
	marshalizer marshal.Marshalizer,
	pauseHandler process.ESDTPauseHandler,
) error {
	esdtData, err := getESDTDataFromKey(userAcnt, key, marshalizer)
	if err != nil {
		return err
	}

	if esdtData.Type != uint32(core.Fungible) {
		return process.ErrOnlyFungibleTokensHaveBalanceTransfer
	}

	err = checkFrozeAndPause(senderAddr, key, esdtData, pauseHandler)
	if err != nil {
		return err
	}

	esdtData.Value.Add(esdtData.Value, value)
	if esdtData.Value.Cmp(zero) < 0 {
		return process.ErrInsufficientFunds
	}

	err = saveESDTData(userAcnt, esdtData, key, marshalizer)
	if err != nil {
		return err
	}

	return nil
}

func checkFrozeAndPause(
	senderAddr []byte,
	key []byte,
	esdtData *esdt.ESDigitalToken,
	pauseHandler process.ESDTPauseHandler,
) error {
	if bytes.Equal(senderAddr, vm.ESDTSCAddress) {
		return nil
	}

	esdtUserMetaData := ESDTUserMetadataFromBytes(esdtData.Properties)
	if esdtUserMetaData.Frozen {
		return process.ErrESDTIsFrozenForAccount
	}

	if pauseHandler.IsPaused(key) {
		return process.ErrESDTTokenIsPaused
	}

	return nil
}

func saveESDTData(
	userAcnt state.UserAccountHandler,
	esdtData *esdt.ESDigitalToken,
	key []byte,
	marshalizer marshal.Marshalizer,
) error {
	marshaledData, err := marshalizer.Marshal(esdtData)
	if err != nil {
		return err
	}

	log.Trace("esdt after transfer", "addr", userAcnt.AddressBytes(), "value", esdtData.Value, "tokenKey", key)
	return userAcnt.DataTrieTracker().SaveKeyValue(key, marshaledData)
}

func getESDTDataFromKey(
	userAcnt state.UserAccountHandler,
	key []byte,
	marshalizer marshal.Marshalizer,
) (*esdt.ESDigitalToken, error) {
	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(0), Type: uint32(core.Fungible)}
	marshaledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil || len(marshaledData) == 0 {
		return esdtData, nil
	}

	err = marshalizer.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}

func (e *esdtTransfer) setPayableHandler(payableHandler process.PayableHandler) error {
	if check.IfNil(payableHandler) {
		return process.ErrNilPayableHandler
	}

	e.payableHandler = payableHandler
	return nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtTransfer) IsInterfaceNil() bool {
	return e == nil
}
