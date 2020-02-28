package smartContract

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestChangeOwnerAddress_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	coa := changeOwnerAddress{}

	owner := []byte("sender")
	tx := &transaction.Transaction{
		SndAddr: owner,
	}

	addr := []byte("addr")
	journalizeWasCalled, saveAccountWasCalled := false, false
	acc, _ := state.NewAccount(mock.NewAddressMock(addr), &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeWasCalled = true
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountWasCalled = true
			return nil
		},
	})
	vmInput := &vmcommon.ContractCallInput{}

	_, err := coa.ProcessBuiltinFunction(tx, nil, acc, vmInput)
	require.Equal(t, process.ErrInvalidArguments, err)

	newAddr := []byte("0000")
	vmInput.Arguments = [][]byte{newAddr}
	_, err = coa.ProcessBuiltinFunction(nil, nil, acc, vmInput)
	require.Equal(t, process.ErrNilTransaction, err)

	_, err = coa.ProcessBuiltinFunction(tx, nil, nil, vmInput)
	require.Equal(t, process.ErrNilSCDestAccount, err)

	acc.OwnerAddress = owner
	_, err = coa.ProcessBuiltinFunction(tx, nil, acc, vmInput)
	require.Nil(t, err)
	require.True(t, journalizeWasCalled)
	require.True(t, saveAccountWasCalled)
}
