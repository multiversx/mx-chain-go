package txsFee

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/stretchr/testify/require"
)

const (
	gasPrice         = uint64(10)
	minGasLimit      = uint64(1)
	gasPriceModifier = float64(0.1)
)

// MetaData defines test meta data struct
type MetaData struct {
	TokenId    []byte
	Nonce      []byte
	Name       []byte
	Royalties  []byte
	Hash       []byte
	Attributes []byte
	Uris       [][]byte
}

// GetDefaultMetaData will return default meta data structure
func GetDefaultMetaData() *MetaData {
	return &MetaData{
		TokenId:    []byte(hex.EncodeToString([]byte("tokenId"))),
		Nonce:      []byte(hex.EncodeToString(big.NewInt(0).Bytes())),
		Name:       []byte(hex.EncodeToString([]byte("name"))),
		Royalties:  []byte(hex.EncodeToString(big.NewInt(10).Bytes())),
		Hash:       []byte(hex.EncodeToString([]byte("hash"))),
		Attributes: []byte(hex.EncodeToString([]byte("attributes"))),
		Uris:       [][]byte{[]byte(hex.EncodeToString([]byte("uri1"))), []byte(hex.EncodeToString([]byte("uri2"))), []byte(hex.EncodeToString([]byte("uri3")))},
	}
}

func getMetaDataFromAcc(t *testing.T, testContext *vm.VMTestContext, accWithMetaData []byte, token []byte) *esdt.MetaData {
	account, err := testContext.Accounts.LoadAccount(accWithMetaData)
	require.Nil(t, err)
	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	key := append(token, big.NewInt(0).SetUint64(1).Bytes()...)
	esdtDataBytes, _, err := userAccount.RetrieveValue(key)
	require.Nil(t, err)
	esdtData := &esdt.ESDigitalToken{}
	err = testContext.Marshalizer.Unmarshal(esdtData, esdtDataBytes)
	require.Nil(t, err)

	return esdtData.TokenMetaData
}

func checkMetaData(t *testing.T, testContext *vm.VMTestContext, accWithMetaData []byte, token []byte, expectedMetaData *MetaData) {
	retrievedMetaData := getMetaDataFromAcc(t, testContext, accWithMetaData, token)

	require.Equal(t, expectedMetaData.Nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, expectedMetaData.Name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, expectedMetaData.Royalties, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Royalties)).Bytes())))
	require.Equal(t, expectedMetaData.Hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expectedMetaData.Uris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, expectedMetaData.Attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))
}

func getDynamicTokenTypes() []string {
	return []string{
		core.DynamicNFTESDT,
		core.DynamicSFTESDT,
		core.DynamicMetaESDT,
	}
}

func createTokenTx(
	sndAddr []byte,
	rcvAddr []byte,
	gasLimit uint64,
	quantity int64,
	metaData *MetaData,
) *transaction.Transaction {
	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			metaData.TokenId,
			[]byte(hex.EncodeToString(big.NewInt(quantity).Bytes())), // quantity
			metaData.Name,
			metaData.Royalties,
			metaData.Hash,
			metaData.Attributes,
			[]byte(hex.EncodeToString([]byte("uri"))),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:    0,
		SndAddr:  sndAddr,
		RcvAddr:  rcvAddr,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}
}

func setTokenTypeTx(
	sndAddr []byte,
	gasLimit uint64,
	tokenId []byte,
	tokenType string,
) *transaction.Transaction {
	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.ESDTSetTokenType),
			[]byte(hex.EncodeToString(tokenId)),
			[]byte(hex.EncodeToString([]byte(tokenType))),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:    0,
		SndAddr:  sndAddr,
		RcvAddr:  core.SystemAccountAddress,
		GasLimit: gasLimit,
		GasPrice: gasPrice,

		Data:  txDataField,
		Value: big.NewInt(0),
	}
}

func getAccount(tb testing.TB, testContext *vm.VMTestContext, scAddress []byte) state.UserAccountHandler {
	scAcc, err := testContext.Accounts.LoadAccount(scAddress)
	require.Nil(tb, err)
	acc, ok := scAcc.(state.UserAccountHandler)
	require.True(tb, ok)

	return acc
}

func getAccountDataTrie(tb testing.TB, testContext *vm.VMTestContext, address []byte) common.Trie {
	acc := getAccount(tb, testContext, address)
	dataTrieInstance, ok := acc.DataTrie().(common.Trie)
	require.True(tb, ok)

	return dataTrieInstance
}

func createAccWithBalance(t *testing.T, accnts state.AccountsAdapter, pubKey []byte, egldValue *big.Int) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	userAccount.IncreaseNonce(0)
	err = userAccount.AddToBalance(egldValue)
	require.Nil(t, err)

	err = accnts.SaveAccount(userAccount)
	require.Nil(t, err)

	_, err = accnts.Commit()
	require.Nil(t, err)
}
