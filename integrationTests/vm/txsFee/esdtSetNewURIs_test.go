package txsFee

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestESDTSetNewURIs(t *testing.T) {
	tokenTypes := getDynamicTokenTypes()
	for _, tokenType := range tokenTypes {
		testName := "ESDTsetNewURIs for " + tokenType
		t.Run(testName, func(t *testing.T) {
			runEsdtSetNewURIsTest(t, tokenType)
		})
	}
}

func runEsdtSetNewURIsTest(t *testing.T, tokenType string) {
	sndAddr := []byte("12345678901234567890123456789012")
	token := []byte("tokenId")
	roles := [][]byte{[]byte(core.ESDTRoleSetNewURI), []byte(core.ESDTRoleNFTCreate)}
	baseEsdtKeyPrefix := core.ProtectedKeyPrefix + core.ESDTKeyIdentifier
	key := append([]byte(baseEsdtKeyPrefix), token...)

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	createAccWithBalance(t, testContext.Accounts, sndAddr, big.NewInt(100000000))
	createAccWithBalance(t, testContext.Accounts, core.ESDTSCAddress, big.NewInt(100000000))
	utils.SetESDTRoles(t, testContext.Accounts, sndAddr, token, roles)

	tx := setTokenTypeTx(core.ESDTSCAddress, 100000, token, tokenType)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	defaultMetaData := getDefaultMetaData()
	tx = createTokenTx(sndAddr, sndAddr, 100000, 1, defaultMetaData)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	defaultMetaData.uris = [][]byte{[]byte(hex.EncodeToString([]byte("newUri1"))), []byte(hex.EncodeToString([]byte("newUri2")))}
	defaultMetaData.nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	tx = esdtSetNewUrisTx(sndAddr, sndAddr, 100000, defaultMetaData)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	retrievedMetaData := getMetaDataFromAcc(t, testContext, core.SystemAccountAddress, key)
	require.Equal(t, [][]byte{[]byte("newUri1"), []byte("newUri2")}, retrievedMetaData.URIs)
}

func esdtSetNewUrisTx(
	sndAddr []byte,
	rcvAddr []byte,
	gasLimit uint64,
	metaData *metaData,
) *transaction.Transaction {
	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.ESDTSetNewURIs),
			metaData.tokenId,
			metaData.nonce,
			metaData.uris[0],
			metaData.uris[1],
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:    1,
		SndAddr:  sndAddr,
		RcvAddr:  rcvAddr,
		GasLimit: gasLimit,
		GasPrice: gasPrice,

		Data:  txDataField,
		Value: big.NewInt(0),
	}
}
