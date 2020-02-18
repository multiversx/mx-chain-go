package smartContract

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func TestClaimDeveloperRewards_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	cdr := claimDeveloperRewards{}

	sender := []byte("sender")
	tx := &transaction.Transaction{
		SndAddr: sender,
	}
	count := 0
	acc, _ := state.NewAccount(mock.NewAddressMock([]byte("addr12")), &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			count++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			count++
			return nil
		},
	})

	reward, err := cdr.ProcessBuiltinFunction(nil, nil, acc, nil)
	require.Nil(t, reward)
	require.Equal(t, process.ErrNilTransaction, err)

	reward, err = cdr.ProcessBuiltinFunction(tx, nil, nil, nil)
	require.Nil(t, reward)
	require.Equal(t, process.ErrNilSCDestAccount, err)

	reward, err = cdr.ProcessBuiltinFunction(tx, nil, acc, nil)
	require.Nil(t, reward)
	require.Equal(t, state.ErrOperationNotPermitted, err)

	acc.OwnerAddress = sender
	value := big.NewInt(100)
	_ = acc.AddToDeveloperReward(value)
	reward, err = cdr.ProcessBuiltinFunction(tx, nil, acc, nil)
	require.Nil(t, err)
	require.Equal(t, 4, count)
	require.Equal(t, value, reward)

}
