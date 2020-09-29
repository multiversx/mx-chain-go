package builtInFunctions

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.BuiltinFunction = (*esdtFreezeWipe)(nil)

type esdtFreezeWipe struct {
	funcGasCost uint64
	marshalizer marshal.Marshalizer
	keyPrefix   []byte
	wipe        bool
	freeze      bool
}

// NewESDTFreezeWipeFunc returns the esdt freeze/un-freeze/wipe built-in function component
func NewESDTFreezeWipeFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
	freeze bool,
	wipe bool,
) (*esdtFreezeWipe, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	e := &esdtFreezeWipe{
		funcGasCost: funcGasCost,
		marshalizer: marshalizer,
		keyPrefix:   []byte(core.ElrondProtectedKeyPrefix + esdtKeyIdentifier),
		freeze:      freeze,
		wipe:        wipe,
	}

	return e, nil
}

// ProcessBuiltinFunction resolve ESDT function calls
func (e *esdtFreezeWipe) ProcessBuiltinFunction(
	_, acntDst state.UserAccountHandler,
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
	if !bytes.Equal(vmInput.CallerAddr, vm.ESDTSCAddress) {
		return nil, process.ErrAddressIsNotESDTSystemSC
	}
	if check.IfNil(acntDst) {
		return nil, process.ErrNilUserAccount
	}

	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	log.Trace(vmInput.Function, "sender", vmInput.CallerAddr, "receiver", vmInput.RecipientAddr, "value", value, "token", esdtTokenKey)

	if e.wipe {
		acntDst.DataTrieTracker().SaveKeyValue(esdtTokenKey, nil)
	} else {
		err := e.toggleFreeze(acntDst, esdtTokenKey)
		if err != nil {
			return nil, err
		}
	}

	vmOutput := &vmcommon.VMOutput{}
	return vmOutput, nil
}

func (e *esdtFreezeWipe) toggleFreeze(acntDst state.UserAccountHandler, tokenKey []byte) error {
	tokenData, err := getESDTDataFromKey(acntDst, tokenKey, e.marshalizer)
	if err != nil {
		return err
	}

	tokenData.Frozen = e.freeze

	err = saveESDTData(acntDst, tokenData, tokenKey, e.marshalizer)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtFreezeWipe) IsInterfaceNil() bool {
	return e == nil
}
