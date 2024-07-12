package vm

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/stretchr/testify/require"
)

func TestChainSimulator_EGLD_MultiTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(4)

	baseIssuingCost := "1000"

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   true,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              numOfShards,
		GenesisTimestamp:         startTime,
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.EGLDInMultiTransferEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := createAddresses(t, cs, false)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	log.Info("Initial setup: Create fungible, NFT, SFT and metaESDT tokens (before the activation of DynamicEsdtFlag)")

	// issue metaESDT
	metaESDTTicker := []byte("METATTICKER")
	tx := issueMetaESDTTx(0, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, addrs[0], metaESDTTokenID, roles)

	log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = issueNonFungibleTx(1, addrs[0].Bytes, nftTicker, baseIssuingCost)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	setAddressEsdtRoles(t, cs, addrs[0], nftTokenID, roles)

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = issueSemiFungibleTx(2, addrs[0].Bytes, sftTicker, baseIssuingCost)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	setAddressEsdtRoles(t, cs, addrs[0], sftTokenID, roles)

	log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	nonce := uint64(3)
	for i := range tokenIDs {
		tx = nftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	account0, err := cs.GetAccount(addrs[0])
	require.Nil(t, err)

	beforeBalanceStr0 := account0.Balance

	account1, err := cs.GetAccount(addrs[1])
	require.Nil(t, err)

	beforeBalanceStr1 := account1.Balance

	egldValue := oneEGLD.Mul(oneEGLD, big.NewInt(3))
	tx = multiESDTNFTTransferWithEGLDTx(nonce, addrs[0].Bytes, addrs[1].Bytes, tokenIDs, egldValue)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	// check accounts balance
	account0, err = cs.GetAccount(addrs[0])
	require.Nil(t, err)

	beforeBalance0, _ := big.NewInt(0).SetString(beforeBalanceStr0, 10)

	expectedBalance0 := big.NewInt(0).Sub(beforeBalance0, egldValue)
	txsFee, _ := big.NewInt(0).SetString(txResult.Fee, 10)
	expectedBalanceWithFee0 := big.NewInt(0).Sub(expectedBalance0, txsFee)

	require.Equal(t, expectedBalanceWithFee0.String(), account0.Balance)

	account1, err = cs.GetAccount(addrs[1])
	require.Nil(t, err)

	beforeBalance1, _ := big.NewInt(0).SetString(beforeBalanceStr1, 10)
	expectedBalance1 := big.NewInt(0).Add(beforeBalance1, egldValue)

	require.Equal(t, expectedBalance1.String(), account1.Balance)
}

func multiESDTNFTTransferWithEGLDTx(nonce uint64, sndAdr, rcvAddr []byte, tokens [][]byte, egldValue *big.Int) *transaction.Transaction {
	transferData := make([]*utils.TransferESDTData, 0)

	for _, tokenID := range tokens {
		transferData = append(transferData, &utils.TransferESDTData{
			Token: tokenID,
			Nonce: 1,
			Value: big.NewInt(1),
		})
	}

	numTransfers := len(tokens)
	encodedReceiver := hex.EncodeToString(rcvAddr)
	hexEncodedNumTransfers := hex.EncodeToString(big.NewInt(int64(numTransfers)).Bytes())
	hexEncodedEGLD := hex.EncodeToString([]byte("EGLD-000000"))
	hexEncodedEGLDNonce := "00"

	txDataField := []byte(strings.Join(
		[]string{
			core.BuiltInFunctionMultiESDTNFTTransfer,
			encodedReceiver,
			hexEncodedNumTransfers,
			hexEncodedEGLD,
			hexEncodedEGLDNonce,
			hex.EncodeToString(egldValue.Bytes()),
		}, "@"),
	)

	for _, td := range transferData {
		hexEncodedToken := hex.EncodeToString(td.Token)
		esdtValueEncoded := hex.EncodeToString(td.Value.Bytes())
		hexEncodedNonce := "00"
		if td.Nonce != 0 {
			hexEncodedNonce = hex.EncodeToString(big.NewInt(int64(td.Nonce)).Bytes())
		}

		txDataField = []byte(strings.Join([]string{string(txDataField), hexEncodedToken, hexEncodedNonce, esdtValueEncoded}, "@"))
	}

	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   sndAdr,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Data:      txDataField,
		Value:     big.NewInt(0),
		Version:   1,
		Signature: []byte("dummySig"),
		ChainID:   []byte(configs.ChainID),
	}

	return tx
}
