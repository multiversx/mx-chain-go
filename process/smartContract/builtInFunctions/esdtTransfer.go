//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. esdt.proto
package builtInFunctions

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const esdtKeyIdentifier = "esdt"

var _ process.BuiltinFunction = (*esdtTransfer)(nil)

var zero = big.NewInt(0)

type esdtTransfer struct {
	funcGasCost uint64
	marshalizer marshal.Marshalizer
	keyPrefix   []byte
}

// NewESDTTransferFunc returns the esdt transfer built-in function component
func NewESDTTransferFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
) (*esdtTransfer, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	e := &esdtTransfer{
		funcGasCost: funcGasCost,
		marshalizer: marshalizer,
		keyPrefix:   []byte(core.ElrondProtectedKeyPrefix + esdtKeyIdentifier),
	}

	return e, nil
}

// ProcessBuiltinFunction will transfer the underlying esdt balance of the account
func (e *esdtTransfer) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if len(vmInput.Arguments) != 2 {
		return nil, process.ErrInvalidArguments
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return nil, process.ErrBuiltInFunctionCalledWithValue
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
		err := e.addToESDTBalance(acntSnd, esdtTokenKey, big.NewInt(0).Neg(value))
		if err != nil {
			return nil, err
		}
	}

	vmOutput := &vmcommon.VMOutput{GasRemaining: gasRemaining}
	if !check.IfNil(acntDst) {
		err := e.addToESDTBalance(acntDst, esdtTokenKey, value)
		if err != nil {
			return nil, err
		}

		return vmOutput, nil
	}

	if core.IsSmartContractAddress(vmInput.CallerAddr) {
		// cross-shard ESDT transfer call through a smart contract - needs the storage update in order to create the smart contract result
		esdtTransferTxData := core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString(vmInput.Arguments[0]) + "@" + hex.EncodeToString(vmInput.Arguments[1])
		outTransfer := vmcommon.OutputTransfer{
			Value:    big.NewInt(0),
			GasLimit: 0,
			Data:     []byte(esdtTransferTxData),
			CallType: vmcommon.AsynchronousCall,
		}
		vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
		vmOutput.OutputAccounts[string(vmInput.RecipientAddr)] = &vmcommon.OutputAccount{
			Address:         vmInput.RecipientAddr,
			OutputTransfers: []vmcommon.OutputTransfer{outTransfer},
		}
	}

	return vmOutput, nil
}

func (e *esdtTransfer) addToESDTBalance(userAcnt state.UserAccountHandler, key []byte, value *big.Int) error {
	esdtData, err := e.getESDTDataFromKey(userAcnt, key)
	if err != nil {
		return err
	}

	esdtData.Value.Add(esdtData.Value, value)
	if esdtData.Value.Cmp(zero) < 0 {
		return process.ErrInsufficientFunds
	}

	marshaledData, err := e.marshalizer.Marshal(esdtData)
	if err != nil {
		return err
	}

	log.Trace("esdt after transfer", "addr", userAcnt.AddressBytes(), "value", esdtData.Value, "tokenKey", key)
	userAcnt.DataTrieTracker().SaveKeyValue(key, marshaledData)

	return nil
}

func (e *esdtTransfer) getESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*ESDigitalToken, error) {
	esdtData := &ESDigitalToken{Value: big.NewInt(0)}
	marshaledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil || len(marshaledData) == 0 {
		return esdtData, nil
	}

	err = e.marshalizer.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtTransfer) IsInterfaceNil() bool {
	return e == nil
}
