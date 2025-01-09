package txsFee

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestESDTModifyRoyalties(t *testing.T) {
	tokenTypes := getDynamicTokenTypes()
	for _, tokenType := range tokenTypes {
		testName := "esdtModifyRoyalties for " + tokenType
		t.Run(testName, func(t *testing.T) {
			runEsdtModifyRoyaltiesTest(t, tokenType)
		})
	}
}

func runEsdtModifyRoyaltiesTest(t *testing.T, tokenType string) {
	creatorAddr := []byte("12345678901234567890123456789013")
	token := []byte("tokenId")
	baseEsdtKeyPrefix := core.ProtectedKeyPrefix + core.ESDTKeyIdentifier
	key := append([]byte(baseEsdtKeyPrefix), token...)

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContext.Close()
	testContext.BlockchainHook.(process.BlockChainHookHandler).SetCurrentHeader(&dataBlock.Header{Round: 7})

	createAccWithBalance(t, testContext.Accounts, creatorAddr, big.NewInt(100000000))
	createAccWithBalance(t, testContext.Accounts, core.ESDTSCAddress, big.NewInt(100000000))
	utils.SetESDTRoles(t, testContext.Accounts, creatorAddr, token, [][]byte{[]byte(core.ESDTRoleModifyRoyalties), []byte(core.ESDTRoleNFTCreate)})

	tx := setTokenTypeTx(core.ESDTSCAddress, 100000, token, tokenType)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	defaultMetaData := GetDefaultMetaData()
	tx = createTokenTx(creatorAddr, creatorAddr, 100000, 1, defaultMetaData)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	defaultMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	defaultMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(20).Bytes()))
	tx = esdtModifyRoyaltiesTx(creatorAddr, creatorAddr, 100000, defaultMetaData)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	retrievedMetaData := &esdt.MetaData{}
	if tokenType == core.DynamicNFTESDT {
		retrievedMetaData = getMetaDataFromAcc(t, testContext, creatorAddr, key)
	} else {
		retrievedMetaData = getMetaDataFromAcc(t, testContext, core.SystemAccountAddress, key)
	}
	require.Equal(t, uint32(big.NewInt(20).Uint64()), retrievedMetaData.Royalties)
}

func esdtModifyRoyaltiesTx(
	sndAddr []byte,
	rcvAddr []byte,
	gasLimit uint64,
	metaData *MetaData,
) *transaction.Transaction {
	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.ESDTModifyRoyalties),
			metaData.TokenId,
			metaData.Nonce,
			metaData.Royalties,
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
