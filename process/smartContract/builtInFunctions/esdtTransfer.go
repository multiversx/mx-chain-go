//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. esdt.proto
package builtInFunctions

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.BuiltinFunction = (*esdtTransfer)(nil)

const esdtKeyPrefix = "esdt"

var zero = big.NewInt(0)

type esdtTransfer struct {
	funcGasCost uint64
	marshalizer marshal.Marshalizer
}

// NewESDTTransferFunc returns the key-value
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

	esdtTokenKey := append([]byte(esdtKeyPrefix), input.Arguments[0]...)
	value := big.NewInt(0).SetBytes(input.Arguments[1])

	if !check.IfNil(acntSnd) {
		// gas is payed only by sender
		if input.GasProvided < e.funcGasCost {
			return nil, input.GasProvided, process.ErrNotEnoughGas
		}

		err := e.addToESDTBalance(acntSnd, esdtTokenKey, value)
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

	return nil, 0, nil
}

func (e *esdtTransfer) addToESDTBalance(userAcnt state.UserAccountHandler, key []byte, value *big.Int) error {
	esdtData := e.getESDTDataFromKey(userAcnt, key)

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

func (e *esdtTransfer) getESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) *ESDigitalToken {
	esdtData := &ESDigitalToken{Value: big.NewInt(0)}
	marshalledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil {
		log.Debug("getESDTDataFromKey retrieve value error", "error", err)
		return esdtData
	}

	err = e.marshalizer.Unmarshal(esdtData, marshalledData)
	if err != nil {
		log.Debug("getESDTDataFromKey unmarshal error", "error", err)
		return esdtData
	}

	return esdtData
}

// IsInterfaceNil return true if underlying object in nil
func (e *esdtTransfer) IsInterfaceNil() bool {
	return e == nil
}
