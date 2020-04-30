package builtInFunctions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestChangeOwnerAddress_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	coa := changeOwnerAddress{}

	owner := []byte("sender")
	addr := []byte("addr")

	acc, _ := state.NewUserAccount(mock.NewAddressMock(addr))
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{CallerAddr: owner},
	}

	_, err := coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, process.ErrInvalidArguments, err)

	newAddr := []byte("0000")
	vmInput.Arguments = [][]byte{newAddr}
	_, err = coa.ProcessBuiltinFunction(nil, acc, nil)
	require.Equal(t, process.ErrNilVmInput, err)

	_, err = coa.ProcessBuiltinFunction(nil, nil, vmInput)
	require.Equal(t, process.ErrNilSCDestAccount, err)

	acc.OwnerAddress = owner
	_, err = coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)
}
