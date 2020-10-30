package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestESDTPause_ProcessBuiltInFunction(t *testing.T) {
	t.Parallel()

	acnt, _ := state.NewUserAccount(core.SystemAccountAddress)
	pauseFunc, _ := NewESDTPauseFunc(&mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return acnt, nil
		},
	}, true)
	_, err := pauseFunc.ProcessBuiltinFunction(nil, nil, nil)
	assert.Equal(t, err, process.ErrNilVmInput)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
		},
	}
	_, err = pauseFunc.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrInvalidArguments)

	input = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(1),
		},
	}
	_, err = pauseFunc.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrBuiltInFunctionCalledWithValue)

	input.CallValue = big.NewInt(0)
	key := []byte("key")
	value := []byte("value")
	input.Arguments = [][]byte{key, value}
	_, err = pauseFunc.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrInvalidArguments)

	input.Arguments = [][]byte{key}
	_, err = pauseFunc.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrAddressIsNotESDTSystemSC)

	input.CallerAddr = vm.ESDTSCAddress
	_, err = pauseFunc.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrOnlySystemAccountAccepted)

	input.RecipientAddr = core.SystemAccountAddress
	_, err = pauseFunc.ProcessBuiltinFunction(nil, nil, input)
	assert.Nil(t, err)

	pauseKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + string(key))
	assert.True(t, pauseFunc.IsPaused(pauseKey))

	esdtPauseFalse, _ := NewESDTPauseFunc(&mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return acnt, nil
		},
	}, false)

	_, err = esdtPauseFalse.ProcessBuiltinFunction(nil, nil, input)
	assert.Nil(t, err)

	assert.False(t, pauseFunc.IsPaused(pauseKey))
}
