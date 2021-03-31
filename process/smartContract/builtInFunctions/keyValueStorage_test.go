package builtInFunctions

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestNewSaveKeyValueStorageFunc(t *testing.T) {
	t.Parallel()

	funcGasCost := uint64(1)
	gasConfig := process.BaseOperationCost{
		StorePerByte: 1,
	}

	kvs, err := NewSaveKeyValueStorageFunc(gasConfig, funcGasCost)
	require.NoError(t, err)
	require.False(t, check.IfNil(kvs))
	require.Equal(t, funcGasCost, kvs.funcGasCost)
	require.Equal(t, gasConfig, kvs.gasConfig)
}

func TestSaveKeyValueStorageFunc_SetNewGasConfig(t *testing.T) {
	t.Parallel()

	funcGasCost := uint64(1)
	gasConfig := process.BaseOperationCost{
		StorePerByte: 1,
	}

	kvs, _ := NewSaveKeyValueStorageFunc(gasConfig, funcGasCost)
	require.NotNil(t, kvs)

	newGasConfig := process.BaseOperationCost{
		StorePerByte: 37,
	}
	newGasCost := &process.GasCost{BaseOperationCost: newGasConfig}

	kvs.SetNewGasConfig(newGasCost)

	require.Equal(t, newGasConfig, kvs.gasConfig)
}

func TestSaveKeyValue_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	funcGasCost := uint64(1)
	gasConfig := process.BaseOperationCost{
		StorePerByte:      1,
		ReleasePerByte:    1,
		DataCopyPerByte:   1,
		PersistPerByte:    1,
		CompilePerByte:    1,
		AoTPreparePerByte: 1,
	}

	skv, _ := NewSaveKeyValueStorageFunc(gasConfig, funcGasCost)

	addr := []byte("addr")
	acc, _ := state.NewUserAccount(addr)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  addr,
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
		RecipientAddr: addr,
	}

	_, err := skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, process.ErrInvalidArguments, err)

	_, err = skv.ProcessBuiltinFunction(nil, acc, nil)
	require.Equal(t, process.ErrNilVmInput, err)

	key := []byte("key")
	value := []byte("value")
	vmInput.Arguments = [][]byte{key, value}

	_, err = skv.ProcessBuiltinFunction(nil, nil, vmInput)
	require.Equal(t, process.ErrNilSCDestAccount, err)

	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)
	retrievedValue, _ := acc.DataTrieTracker().RetrieveValue(key)
	require.True(t, bytes.Equal(retrievedValue, value))

	vmInput.CallerAddr = []byte("other")
	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.True(t, errors.Is(err, process.ErrOperationNotPermitted))

	key = []byte(core.ElrondProtectedKeyPrefix + "is the king")
	value = []byte("value")
	vmInput.Arguments = [][]byte{key, value}

	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.True(t, errors.Is(err, process.ErrOperationNotPermitted))
}

func TestSaveKeyValue_ProcessBuiltinFunctionMultipleKeys(t *testing.T) {
	t.Parallel()

	funcGasCost := uint64(1)
	gasConfig := process.BaseOperationCost{
		StorePerByte:    1,
		ReleasePerByte:  1,
		DataCopyPerByte: 1,
		PersistPerByte:  1,
		CompilePerByte:  1,
	}
	skv, _ := NewSaveKeyValueStorageFunc(gasConfig, funcGasCost)

	addr := []byte("addr")
	acc, _ := state.NewUserAccount(addr)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  addr,
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
		RecipientAddr: addr,
	}

	key := []byte("key")
	value := []byte("value")
	vmInput.Arguments = [][]byte{key, value, key, value, key}

	_, err := skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, err, process.ErrInvalidArguments)

	key2 := []byte("key2")
	value2 := []byte("value2")
	vmInput.Arguments = [][]byte{key, value, key2, value2}

	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)
	retrievedValue, _ := acc.DataTrieTracker().RetrieveValue(key)
	require.True(t, bytes.Equal(retrievedValue, value))
	retrievedValue, _ = acc.DataTrieTracker().RetrieveValue(key2)
	require.True(t, bytes.Equal(retrievedValue, value2))

	vmInput.GasProvided = 1
	vmInput.Arguments = [][]byte{[]byte("key3"), []byte("value")}
	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, err, process.ErrNotEnoughGas)
}
