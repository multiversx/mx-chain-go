package accounts

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/stretchr/testify/require"
)

func TestSerializeAccounts(t *testing.T) {
	t.Parallel()

	accs := map[string]*types.AccountInfo{
		"addr1": {
			Address: "addr1",
			Nonce:   1,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, 8000, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "addr1" } }
{"address":"addr1","nonce":1,"balance":"","balanceNum":0}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccountsESDT(t *testing.T) {
	t.Parallel()

	accs := map[string]*types.AccountInfo{
		"addr1": {
			Address:         "addr1",
			Nonce:           1,
			TokenIdentifier: "token-0001",
			Properties:      "000",
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, 8000, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "addr1_token-0001" } }
{"address":"addr1","nonce":1,"balance":"","balanceNum":0,"token":"token-0001","properties":"000"}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccountsHistory(t *testing.T) {
	t.Parallel()

	accsh := map[string]*types.AccountBalanceHistory{
		"account1": {
			Address:         "account1",
			Timestamp:       10,
			Balance:         "123",
			TokenIdentifier: "token-0001",
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccountsHistory(accsh, 8000)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { } }
{"address":"account1","timestamp":10,"balance":"123","token":"token-0001"}
`
	require.Equal(t, expectedRes, res[0].String())
}
