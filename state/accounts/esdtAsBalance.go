package accounts

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/errors"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"math/big"
)

const baseESDTKeyPrefix = core.ProtectedKeyPrefix + core.ESDTKeyIdentifier

type esdtAsBalance struct {
	keyPrefix  []byte
	marshaller marshal.Marshalizer
}

// NewESDTAsBalance creates the esdtAsBalance component
func NewESDTAsBalance(
	baseTokenID string,
	marshaller marshal.Marshalizer,
) (*esdtAsBalance, error) {
	if check.IfNil(marshaller) {
		return nil, errors.ErrNilMarshalizer
	}
	if len(baseTokenID) == 0 {
		return nil, errors.ErrEmptyBaseToken
	}

	e := &esdtAsBalance{keyPrefix: []byte(baseESDTKeyPrefix + baseTokenID)}

	return e, nil
}

func (e *esdtAsBalance) GetBalance(accountDataHandler vmcommon.AccountDataHandler) *big.Int {
	esdtData, err := e.getESDTData(accountDataHandler)
	if err != nil {
		return big.NewInt(0)
	}

	return esdtData.Value
}

func (e *esdtAsBalance) AddToBalance(accountDataHandler vmcommon.AccountDataHandler, value *big.Int) error {
	esdtData, err := e.getESDTData(accountDataHandler)
	if err != nil {
		return err
	}

	newBalance := big.NewInt(0).Add(esdtData.Value, value)
	if newBalance.Cmp(zero) < 0 {
		return errors.ErrInsufficientFunds
	}

	esdtData.Value.Set(newBalance)
	err = e.saveESDTData(accountDataHandler, esdtData)
	if err != nil {
		return err
	}

	return nil
}

func (e *esdtAsBalance) SubFromBalance(accountDataHandler vmcommon.AccountDataHandler, value *big.Int) error {
	esdtData, err := e.getESDTData(accountDataHandler)
	if err != nil {
		return err
	}

	newBalance := big.NewInt(0).Sub(esdtData.Value, value)
	if newBalance.Cmp(zero) < 0 {
		return errors.ErrInsufficientFunds
	}

	esdtData.Value.Set(newBalance)
	err = e.saveESDTData(accountDataHandler, esdtData)
	if err != nil {
		return err
	}

	return nil
}

func (e *esdtAsBalance) getESDTData(accountDataHandler vmcommon.AccountDataHandler) (*esdt.ESDigitalToken, error) {
	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(0), Type: uint32(core.Fungible)}
	marshaledData, _, err := accountDataHandler.RetrieveValue(e.keyPrefix)
	if err != nil {
		return nil, err
	}

	err = e.marshaller.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}

func (e *esdtAsBalance) saveESDTData(accountDataHandler vmcommon.AccountDataHandler, esdtData *esdt.ESDigitalToken) error {
	marshaledData, err := e.marshaller.Marshal(esdtData)
	if err != nil {
		return err
	}

	err = accountDataHandler.SaveKeyValue(e.keyPrefix, marshaledData)
	if err != nil {
		return err
	}

	return nil
}

func (e *esdtAsBalance) IsInterfaceNil() bool {
	return e == nil
}
