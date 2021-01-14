package utils

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/stretchr/testify/require"
)

// CreateAccountWithESDTBalance -
func CreateAccountWithESDTBalance(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	egldValue *big.Int,
	tokenIdentifier []byte,
	esdtValue *big.Int,
) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	userAccount.IncreaseNonce(0)
	err = userAccount.AddToBalance(egldValue)
	require.Nil(t, err)

	esdtData := &esdt.ESDigitalToken{
		Value:      esdtValue,
		Properties: []byte{},
	}

	esdtDataBytes, err := protoMarshalizer.Marshal(esdtData)
	require.Nil(t, err)

	key := append([]byte("ELRONDesdt"), tokenIdentifier...)
	err = userAccount.DataTrieTracker().SaveKeyValue(key, esdtDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

// CreateESDTTransferTx -
func CreateESDTTransferTx(nonce uint64, sndAddr, rcvAddr []byte, tokenIdentifier []byte, esdtValue *big.Int, gasPrice, gasLimit uint64) *transaction.Transaction {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	esdtValueEncoded := hex.EncodeToString(esdtValue.Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionESDTTransfer), []byte(hexEncodedToken), []byte(esdtValueEncoded)}, []byte("@"))

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  rcvAddr,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}
}

// CheckESDTBalance -
func CheckESDTBalance(t *testing.T, testContext *vm.VMTestContext, addr []byte, tokenIdentifier []byte, expectedBalance *big.Int) {
	account, err := testContext.Accounts.LoadAccount(addr)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	tokenKey := core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + string(tokenIdentifier)
	valueBytes, err := userAccount.DataTrieTracker().RetrieveValue([]byte(tokenKey))
	if err != nil || len(valueBytes) == 0 {
		require.Equal(t, big.NewInt(0), expectedBalance)
		return
	}

	esdtToken := &esdt.ESDigitalToken{}
	err = protoMarshalizer.Unmarshal(esdtToken, valueBytes)
	require.Nil(t, err)

	require.Equal(t, expectedBalance, esdtToken.Value)
}
