package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestChangeOwnerAddress_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	coa := changeOwnerAddress{}

	owner := []byte("send")
	addr := []byte("addr")

	acc, _ := state.NewUserAccount(addr)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{CallerAddr: owner, CallValue: big.NewInt(0)},
	}

	_, err := coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, process.ErrInvalidArguments, err)

	newAddr := []byte("0000")
	vmInput.Arguments = [][]byte{newAddr}
	_, err = coa.ProcessBuiltinFunction(nil, acc, nil)
	require.Equal(t, process.ErrNilVmInput, err)

	_, err = coa.ProcessBuiltinFunction(nil, nil, vmInput)
	require.Nil(t, err)

	var vmOutput *vmcommon.VMOutput
	acc.OwnerAddress = owner
	vmInput.GasProvided = 10
	vmOutput, err = coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)
	require.Equal(t, vmOutput.GasRemaining, uint64(0))

	coa.gasCost = 1
	vmInput.GasProvided = 10
	acc.OwnerAddress = owner
	vmOutput, err = coa.ProcessBuiltinFunction(acc, acc, vmInput)
	require.Nil(t, err)
	require.Equal(t, vmOutput.GasRemaining, vmInput.GasProvided-coa.gasCost)
}
