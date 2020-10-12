package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestESDTFreezeWipe_ProcessBuiltInFunctionErrors(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	freeze, _ := NewESDTFreezeWipeFunc(marshalizer, true, false)
	_, err := freeze.ProcessBuiltinFunction(nil, nil, nil)
	assert.Equal(t, err, process.ErrNilVmInput)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
		},
	}
	_, err = freeze.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrInvalidArguments)

	input = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(1),
		},
	}
	_, err = freeze.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrBuiltInFunctionCalledWithValue)

	input.CallValue = big.NewInt(0)
	key := []byte("key")
	value := []byte("value")
	input.Arguments = [][]byte{key, value}
	_, err = freeze.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrInvalidArguments)

	input.Arguments = [][]byte{key}
	_, err = freeze.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrAddressIsNotESDTSystemSC)

	input.CallerAddr = vm.ESDTSCAddress
	_, err = freeze.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrNilUserAccount)

	input.RecipientAddr = []byte("dst")
	acnt, _ := state.NewUserAccount(input.RecipientAddr)
	_, err = freeze.ProcessBuiltinFunction(nil, acnt, input)
	assert.Nil(t, err)

	esdtToken := &ESDigitalToken{}
	esdtKey := append(freeze.keyPrefix, key...)
	marshaledData, _ := acnt.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)

	esdtUserData := ESDTUserMetadataFromBytes(esdtToken.Properties)
	assert.True(t, esdtUserData.Frozen)
}

func TestESDTFreezeWipe_ProcessBuiltInFunction(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	freeze, _ := NewESDTFreezeWipeFunc(marshalizer, true, false)
	_, err := freeze.ProcessBuiltinFunction(nil, nil, nil)
	assert.Equal(t, err, process.ErrNilVmInput)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
		},
	}
	key := []byte("key")

	input.Arguments = [][]byte{key}
	input.CallerAddr = vm.ESDTSCAddress
	input.RecipientAddr = []byte("dst")

	acnt, _ := state.NewUserAccount(input.RecipientAddr)
	_, err = freeze.ProcessBuiltinFunction(nil, acnt, input)
	assert.Nil(t, err)

	esdtToken := &ESDigitalToken{}
	esdtKey := append(freeze.keyPrefix, key...)
	marshaledData, _ := acnt.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)

	esdtUserData := ESDTUserMetadataFromBytes(esdtToken.Properties)
	assert.True(t, esdtUserData.Frozen)

	unFreeze, _ := NewESDTFreezeWipeFunc(marshalizer, false, false)
	_, err = unFreeze.ProcessBuiltinFunction(nil, acnt, input)
	assert.Nil(t, err)

	marshaledData, _ = acnt.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)

	esdtUserData = ESDTUserMetadataFromBytes(esdtToken.Properties)
	assert.False(t, esdtUserData.Frozen)

	wipe, _ := NewESDTFreezeWipeFunc(marshalizer, false, true)
	_, err = wipe.ProcessBuiltinFunction(nil, acnt, input)
	assert.Nil(t, err)

	marshaledData, _ = acnt.DataTrieTracker().RetrieveValue(esdtKey)
	assert.Equal(t, 0, len(marshaledData))
}
