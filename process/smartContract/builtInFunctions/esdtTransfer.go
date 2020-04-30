//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. esdt.proto
package builtInFunctions

import (
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
	input *vmcommon.ContractCallInput,
) (*big.Int, uint64, error) {
	if input == nil {
		return nil, 0, process.ErrNilVmInput
	}
	if len(input.Arguments) != 2 {
		return nil, input.GasProvided, process.ErrInvalidArguments
	}

	gasRemaining := input.GasProvided
	esdtTokenKey := append(e.keyPrefix, input.Arguments[0]...)
	value := big.NewInt(0).SetBytes(input.Arguments[1])

	if !check.IfNil(acntSnd) {
		// gas is payed only by sender
		if input.GasProvided < e.funcGasCost {
			return nil, input.GasProvided, process.ErrNotEnoughGas
		}

		gasRemaining = input.GasProvided - e.funcGasCost
		err := e.addToESDTBalance(acntSnd, esdtTokenKey, big.NewInt(0).Neg(value))
		if err != nil {
			return nil, input.GasProvided, err
		}
	}

	if !check.IfNil(acntDst) {
		err := e.addToESDTBalance(acntDst, esdtTokenKey, value)
		if err != nil {
			return nil, input.GasProvided, err
		}
	}

	return nil, gasRemaining, nil
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

	marshalledData, err := e.marshalizer.Marshal(esdtData)
	if err != nil {
		return err
	}

	userAcnt.DataTrieTracker().SaveKeyValue(key, marshalledData)

	return nil
}

func (e *esdtTransfer) getESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*ESDigitalToken, error) {
	esdtData := &ESDigitalToken{Value: big.NewInt(0)}
	marshalledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil {
		log.Debug("getESDTDataFromKey retrieve value error", "error", err)
		return esdtData, nil
	}

	err = e.marshalizer.Unmarshal(esdtData, marshalledData)
	if err != nil {
		log.Debug("getESDTDataFromKey unmarshal error", "error", err)
		return nil, err
	}

	return esdtData, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtTransfer) IsInterfaceNil() bool {
	return e == nil
}
