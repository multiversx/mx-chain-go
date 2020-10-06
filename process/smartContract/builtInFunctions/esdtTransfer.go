//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. esdt.proto
package builtInFunctions

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const esdtKeyIdentifier = "esdt"

var _ process.BuiltinFunction = (*esdtTransfer)(nil)

var zero = big.NewInt(0)

type esdtTransfer struct {
	funcGasCost    uint64
	marshalizer    marshal.Marshalizer
	keyPrefix      []byte
	pauseHandler   process.ESDTPauseHandler
	payableHandler process.PayableHandler
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
		keyPrefix:      []byte(core.ElrondProtectedKeyPrefix + esdtKeyIdentifier),
		pauseHandler:   pauseHandler,
		payableHandler: &disabledPayableHandler{},
	}

	return e, nil
}

// ProcessBuiltinFunction resolves ESDT transfer function calls
func (e *esdtTransfer) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
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

	gasRemaining := uint64(0)
	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	log.Trace("esdtTransfer", "sender", vmInput.CallerAddr, "receiver", vmInput.RecipientAddr, "value", value, "token", esdtTokenKey)

	if !check.IfNil(acntSnd) {
		// gas is paid only by sender
		if vmInput.GasProvided < e.funcGasCost {
			return nil, process.ErrNotEnoughGas
		}

		gasRemaining = vmInput.GasProvided - e.funcGasCost
		err := addToESDTBalance(vmInput.CallerAddr, acntSnd, esdtTokenKey, big.NewInt(0).Neg(value), e.marshalizer, e.pauseHandler)
		if err != nil {
			return nil, err
		}
	}

	vmOutput := &vmcommon.VMOutput{GasRemaining: gasRemaining}
	if !check.IfNil(acntDst) {
		isPayable, err := e.payableHandler.IsPayable(vmInput.RecipientAddr)
		if err != nil {
			return nil, err
		}
		if !isPayable {
			return nil, process.ErrAccountNotPayable
		}

		err = addToESDTBalance(vmInput.CallerAddr, acntDst, esdtTokenKey, value, e.marshalizer, e.pauseHandler)
		if err != nil {
			return nil, err
		}

		return vmOutput, nil
	}

	addOutPutTransferToVMOutput(
		core.BuiltInFunctionESDTTransfer,
		vmInput.Arguments[0],
		vmInput.Arguments[1],
		vmInput.RecipientAddr,
		vmInput.CallerAddr,
		vmOutput)

	return vmOutput, nil
}

func addOutPutTransferToVMOutput(
	function string,
	tokenName []byte,
	tokenValue []byte,
	recipient []byte,
	caller []byte,
	vmOutput *vmcommon.VMOutput,
) {
	if !core.IsSmartContractAddress(caller) {
		return
	}

	// cross-shard ESDT transfer call through a smart contract - needs the storage update in order to create the smart contract result
	esdtTransferTxData := function + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(tokenValue)
	outTransfer := vmcommon.OutputTransfer{
		Value:    big.NewInt(0),
		GasLimit: 0,
		Data:     []byte(esdtTransferTxData),
		CallType: vmcommon.AsynchronousCall,
	}
	vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	vmOutput.OutputAccounts[string(recipient)] = &vmcommon.OutputAccount{
		Address:         recipient,
		OutputTransfers: []vmcommon.OutputTransfer{outTransfer},
	}
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

	if !bytes.Equal(senderAddr, vm.ESDTSCAddress) {
		esdtUserMetaData := ESDTUserMetadataFromBytes(esdtData.Properties)
		if esdtUserMetaData.Frozen {
			return process.ErrESDTIsFrozenForAccount
		}

		if pauseHandler.IsPaused(key) {
			return process.ErrESDTTokenIsPaused
		}
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

func saveESDTData(
	userAcnt state.UserAccountHandler,
	esdtData *ESDigitalToken,
	key []byte,
	marshalizer marshal.Marshalizer,
) error {
	marshaledData, err := marshalizer.Marshal(esdtData)
	if err != nil {
		return err
	}

	log.Trace("esdt after transfer", "addr", userAcnt.AddressBytes(), "value", esdtData.Value, "tokenKey", key)
	userAcnt.DataTrieTracker().SaveKeyValue(key, marshaledData)
	return nil
}

func getESDTDataFromKey(
	userAcnt state.UserAccountHandler,
	key []byte,
	marshalizer marshal.Marshalizer,
) (*ESDigitalToken, error) {
	esdtData := &ESDigitalToken{Value: big.NewInt(0)}
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
