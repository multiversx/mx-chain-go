package smartContract

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state/accounts"

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

	acc, _ := accounts.NewUserAccount(mock.NewAddressMock([]byte("addr12")))

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
	acc.SetDeveloperReward(big.NewInt(0).Add(acc.GetDeveloperReward(), value))
	reward, err = cdr.ProcessBuiltinFunction(tx, nil, acc, nil)
	require.Nil(t, err)
	require.Equal(t, value, reward)

}
