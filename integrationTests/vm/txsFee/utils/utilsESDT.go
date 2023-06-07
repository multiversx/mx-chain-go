package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
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

	userAccount, ok := account.(common.UserAccountHandler)
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
			Name:    []byte(fmt.Sprintf("Token %d", esdtNonce)),
			URIs:    [][]byte{[]byte(fmt.Sprintf("URI for token %d", esdtNonce))},
			Creator: pubKey,
			Nonce:   esdtNonce,
		}
	}

	esdtDataBytes, err := integrationTests.TestMarshalizer.Marshal(esdtData)
	require.Nil(t, err)

	key := append([]byte(core.ProtectedKeyPrefix), []byte(core.ESDTKeyIdentifier)...)
	key = append(key, tokenIdentifier...)
	if esdtNonce > 0 {
		key = append(key, big.NewInt(0).SetUint64(esdtNonce).Bytes()...)
	}

	err = userAccount.SaveKeyValue(key, esdtDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	saveNewTokenOnSystemAccount(t, accnts, key, esdtData)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

// CreateAccountWithNFT -
func CreateAccountWithNFT(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	egldValue *big.Int,
	tokenIdentifier []byte,
	attributes []byte,
) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(common.UserAccountHandler)
	require.True(t, ok)

	userAccount.IncreaseNonce(0)
	err = userAccount.AddToBalance(egldValue)
	require.Nil(t, err)

	esdtData := &esdt.ESDigitalToken{
		Value:      big.NewInt(1),
		Properties: []byte{},
		TokenMetaData: &esdt.MetaData{
			Nonce:      1,
			Attributes: attributes,
		},
	}

	esdtDataBytes, err := integrationTests.TestMarshalizer.Marshal(esdtData)
	require.Nil(t, err)

	key := append([]byte(core.ProtectedKeyPrefix), []byte(core.ESDTKeyIdentifier)...)
	key = append(key, tokenIdentifier...)
	key = append(key, big.NewInt(0).SetUint64(1).Bytes()...)

	err = userAccount.SaveKeyValue(key, esdtDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	saveNewTokenOnSystemAccount(t, accnts, key, esdtData)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

func saveNewTokenOnSystemAccount(t *testing.T, accnts state.AccountsAdapter, tokenKey []byte, esdtData *esdt.ESDigitalToken) {
	esdtDataOnSystemAcc := esdtData
	esdtDataOnSystemAcc.Properties = nil
	esdtDataOnSystemAcc.Reserved = []byte{1}
	esdtDataOnSystemAcc.Value.Set(esdtData.Value)

	esdtDataBytes, err := integrationTests.TestMarshalizer.Marshal(esdtData)
	require.Nil(t, err)

	sysAccount, err := accnts.LoadAccount(core.SystemAccountAddress)
	require.Nil(t, err)

	sysUserAccount, ok := sysAccount.(common.UserAccountHandler)
	require.True(t, ok)

	err = sysUserAccount.SaveKeyValue(tokenKey, esdtDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(sysAccount)
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

	userAccount, ok := account.(common.UserAccountHandler)
	require.True(t, ok)

	key := append([]byte(core.ProtectedKeyPrefix), append([]byte(core.ESDTRoleIdentifier), []byte(core.ESDTKeyIdentifier)...)...)
	key = append(key, tokenIdentifier...)

	if len(roles) == 0 {
		err = userAccount.SaveKeyValue(key, []byte{})
		require.Nil(t, err)

		return
	}

	rolesData := &esdt.ESDTRoles{
		Roles: roles,
	}

	rolesDataBytes, err := integrationTests.TestMarshalizer.Marshal(rolesData)
	require.Nil(t, err)

	err = userAccount.SaveKeyValue(key, rolesDataBytes)
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

	userAccount, ok := account.(common.UserAccountHandler)
	require.True(t, ok)

	key := append([]byte(core.ProtectedKeyPrefix), []byte(core.ESDTNFTLatestNonceIdentifier)...)
	key = append(key, tokenIdentifier...)

	err = userAccount.SaveKeyValue(key, big.NewInt(int64(lastNonce)).Bytes())
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

// TransferESDTData -
type TransferESDTData struct {
	Token []byte
	Nonce uint64
	Value *big.Int
}

// CreateMultiTransferTX -
func CreateMultiTransferTX(nonce uint64, sender, dest []byte, gasPrice, gasLimit uint64, tds ...*TransferESDTData) *transaction.Transaction {
	numTransfers := len(tds)
	encodedReceiver := hex.EncodeToString(dest)
	hexEncodedNumTransfers := hex.EncodeToString(big.NewInt(int64(numTransfers)).Bytes())

	txDataField := []byte(strings.Join([]string{core.BuiltInFunctionMultiESDTNFTTransfer, encodedReceiver, hexEncodedNumTransfers}, "@"))
	for _, td := range tds {
		hexEncodedToken := hex.EncodeToString(td.Token)
		esdtValueEncoded := hex.EncodeToString(td.Value.Bytes())
		hexEncodedNonce := "00"
		if td.Nonce != 0 {
			hexEncodedNonce = hex.EncodeToString(big.NewInt(int64(td.Nonce)).Bytes())
		}

		txDataField = []byte(strings.Join([]string{string(txDataField), hexEncodedToken, hexEncodedNonce, esdtValueEncoded}, "@"))
	}

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sender,
		RcvAddr:  sender,
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
func CheckESDTNFTBalance(tb testing.TB, testContext *vm.VMTestContext, addr []byte, tokenIdentifier []byte, esdtNonce uint64, expectedBalance *big.Int) {
	checkEsdtBalance(tb, testContext, addr, tokenIdentifier, esdtNonce, expectedBalance)
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

// CreateESDTNFTBurnTx -
func CreateESDTNFTBurnTx(nonce uint64, sndAddr, rcvAddr []byte, tokenIdentifier []byte, tokenNonce uint64, esdtValue *big.Int, gasPrice, gasLimit uint64) *transaction.Transaction {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	hexEncodedNonce := hex.EncodeToString(big.NewInt(int64(tokenNonce)).Bytes())
	esdtValueEncoded := hex.EncodeToString(esdtValue.Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionESDTNFTBurn), []byte(hexEncodedToken), []byte(hexEncodedNonce), []byte(esdtValueEncoded)}, []byte("@"))

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

// CreateNFTSingleFreezeAndWipeTxs -
func CreateNFTSingleFreezeAndWipeTxs(nonce uint64, tokenManager, addressToFreeze []byte, tokenIdentifier []byte, tokenNonce uint64, gasPrice, gasLimit uint64) (*transaction.Transaction, *transaction.Transaction) {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	hexEncodedNonce := hex.EncodeToString(big.NewInt(int64(tokenNonce)).Bytes())
	addressToFreezeHex := hex.EncodeToString(addressToFreeze)

	txDataField := bytes.Join([][]byte{[]byte("freezeSingleNFT"), []byte(hexEncodedToken), []byte(hexEncodedNonce), []byte(addressToFreezeHex)}, []byte("@"))
	freezeTx := &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  tokenManager,
		RcvAddr:  core.ESDTSCAddress,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}

	txDataField = bytes.Join([][]byte{[]byte("wipeSingleNFT"), []byte(hexEncodedToken), []byte(hexEncodedNonce), []byte(addressToFreezeHex)}, []byte("@"))
	wipeTx := &transaction.Transaction{
		Nonce:    nonce + 1,
		SndAddr:  tokenManager,
		RcvAddr:  core.ESDTSCAddress,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}

	return freezeTx, wipeTx
}

func checkEsdtBalance(
	tb testing.TB,
	testContext *vm.VMTestContext,
	addr []byte,
	tokenIdentifier []byte,
	esdtNonce uint64,
	expectedBalance *big.Int,
) {
	esdtData, err := testContext.BlockchainHook.GetESDTToken(addr, tokenIdentifier, esdtNonce)
	require.Nil(tb, err)
	require.Equal(tb, expectedBalance, esdtData.Value)
}

// CreateESDTNFTUpdateAttributesTx -
func CreateESDTNFTUpdateAttributesTx(
	nonce uint64,
	sndAddr []byte,
	tokenIdentifier []byte,
	gasPrice uint64,
	gasLimit uint64,
	newAttributes []byte,
) *transaction.Transaction {

	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionESDTNFTUpdateAttributes)
	txData.Bytes(tokenIdentifier)
	txData.Int64(1)
	txData.Bytes(newAttributes)

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  sndAddr, // receiver = sender for ESDTNFTUpdateAttributes
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txData.ToBytes(),
		Value:    big.NewInt(0),
	}
}
