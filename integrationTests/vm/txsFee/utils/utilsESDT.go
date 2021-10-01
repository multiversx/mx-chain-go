package utils

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/stretchr/testify/require"
)

// CreateAccountWithESDTBalance -
func CreateAccountWithESDTBalance(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	egldValue *big.Int,
	tokenIdentifier []byte,
	esdtNonce uint64,
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
	if esdtNonce > 0 {
		esdtData.TokenMetaData = &esdt.MetaData{
			Nonce: esdtNonce,
		}
	}

	esdtDataBytes, err := protoMarshalizer.Marshal(esdtData)
	require.Nil(t, err)

	key := append([]byte(core.ElrondProtectedKeyPrefix), []byte(core.ESDTKeyIdentifier)...)
	key = append(key, tokenIdentifier...)
	if esdtNonce > 0 {
		key = append(key, big.NewInt(0).SetUint64(esdtNonce).Bytes()...)
	}

	err = userAccount.DataTrieTracker().SaveKeyValue(key, esdtDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

// CreateAccountWithESDTBalanceAndRoles -
func CreateAccountWithESDTBalanceAndRoles(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	egldValue *big.Int,
	tokenIdentifier []byte,
	esdtNonce uint64,
	esdtValue *big.Int,
	roles [][]byte,
) {
	CreateAccountWithESDTBalance(t, accnts, pubKey, egldValue, tokenIdentifier, esdtNonce, esdtValue)
	SetESDTRoles(t, accnts, pubKey, tokenIdentifier, roles)
}

// SetESDTRoles -
func SetESDTRoles(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	tokenIdentifier []byte,
	roles [][]byte,
) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	key := append([]byte(core.ElrondProtectedKeyPrefix), append([]byte(core.ESDTRoleIdentifier), []byte(core.ESDTKeyIdentifier)...)...)
	key = append(key, tokenIdentifier...)

	if len(roles) == 0 {
		err = userAccount.DataTrieTracker().SaveKeyValue(key, []byte{})
		require.Nil(t, err)

		return
	}

	rolesData := &esdt.ESDTRoles{
		Roles: roles,
	}

	rolesDataBytes, err := protoMarshalizer.Marshal(rolesData)
	require.Nil(t, err)

	err = userAccount.DataTrieTracker().SaveKeyValue(key, rolesDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

// SetLastNFTNonce -
func SetLastNFTNonce(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	tokenIdentifier []byte,
	lastNonce uint64,
) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	key := append([]byte(core.ElrondProtectedKeyPrefix), []byte(core.ESDTNFTLatestNonceIdentifier)...)
	key = append(key, tokenIdentifier...)

	err = userAccount.DataTrieTracker().SaveKeyValue(key, big.NewInt(int64(lastNonce)).Bytes())
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

// CreateESDTNFTTransferTx -
func CreateESDTNFTTransferTx(
	nonce uint64,
	sndAddr []byte,
	rcvAddr []byte,
	tokenIdentifier []byte,
	esdtNonce uint64,
	esdtValue *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	endpointName string,
	arguments ...[]byte) *transaction.Transaction {

	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionESDTNFTTransfer)
	txData.Bytes(tokenIdentifier)
	txData.Int64(int64(esdtNonce))
	txData.BigInt(esdtValue)
	txData.Bytes(rcvAddr)

	if len(endpointName) > 0 {
		txData.Str(endpointName)

		for _, arg := range arguments {
			txData.Bytes(arg)
		}
	}

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  sndAddr, // receiver = sender for ESDTNFTTransfer
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txData.ToBytes(),
		Value:    big.NewInt(0),
	}
}

// CheckESDTBalance -
func CheckESDTBalance(t *testing.T, testContext *vm.VMTestContext, addr []byte, tokenIdentifier []byte, expectedBalance *big.Int) {
	checkEsdtBalance(t, testContext, addr, tokenIdentifier, 0, expectedBalance)
}

// CheckESDTNFTBalance -
func CheckESDTNFTBalance(t *testing.T, testContext *vm.VMTestContext, addr []byte, tokenIdentifier []byte, esdtNonce uint64, expectedBalance *big.Int) {
	checkEsdtBalance(t, testContext, addr, tokenIdentifier, esdtNonce, expectedBalance)
}

// CreateESDTLocalBurnTx -
func CreateESDTLocalBurnTx(nonce uint64, sndAddr, rcvAddr []byte, tokenIdentifier []byte, esdtValue *big.Int, gasPrice, gasLimit uint64) *transaction.Transaction {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	esdtValueEncoded := hex.EncodeToString(esdtValue.Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionESDTLocalBurn), []byte(hexEncodedToken), []byte(esdtValueEncoded)}, []byte("@"))

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

// CreateESDTLocalMintTx -
func CreateESDTLocalMintTx(nonce uint64, sndAddr, rcvAddr []byte, tokenIdentifier []byte, esdtValue *big.Int, gasPrice, gasLimit uint64) *transaction.Transaction {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	esdtValueEncoded := hex.EncodeToString(esdtValue.Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionESDTLocalMint), []byte(hexEncodedToken), []byte(esdtValueEncoded)}, []byte("@"))

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

func checkEsdtBalance(
	t *testing.T,
	testContext *vm.VMTestContext,
	addr []byte,
	tokenIdentifier []byte,
	esdtNonce uint64,
	expectedBalance *big.Int) {

	account, err := testContext.Accounts.LoadAccount(addr)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	tokenKey := core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + string(tokenIdentifier)
	if esdtNonce > 0 {
		tokenKey += string(big.NewInt(int64(esdtNonce)).Bytes())
	}

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
