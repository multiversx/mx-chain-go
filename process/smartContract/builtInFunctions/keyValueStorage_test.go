package builtInFunctions

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestSaveKeyValue_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	funcGasCost := uint64(1)
	gasConfig := BaseOperationCost{
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
		VMInput:       vmcommon.VMInput{CallerAddr: addr, GasProvided: 50},
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
