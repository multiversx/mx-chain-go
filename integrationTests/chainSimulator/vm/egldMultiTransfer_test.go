package vm

import (
	"encoding/hex"
	"fmt"
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
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
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

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := issueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, nonce, addrs[0], metaESDTTokenID, roles)
	nonce++

	log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = issueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	setAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = issueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	setAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

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

	for i := range tokenIDs {
		tx = esdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		fmt.Println(txResult)
		fmt.Println(string(txResult.Logs.Events[0].Topics[0]))
		fmt.Println(string(txResult.Logs.Events[0].Topics[1]))

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

	egldLog := string(txResult.Logs.Events[0].Topics[0])
	require.Equal(t, "EGLD-000000", egldLog)

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

func TestChainSimulator_EGLD_MultiTransfer_Insufficient_Funds(t *testing.T) {
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

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	nonce := uint64(0)
	tx := issueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = esdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	account0, err := cs.GetAccount(addrs[0])
	require.Nil(t, err)

	beforeBalanceStr0 := account0.Balance

	account1, err := cs.GetAccount(addrs[1])
	require.Nil(t, err)

	beforeBalanceStr1 := account1.Balance

	egldValue, _ := big.NewInt(0).SetString(beforeBalanceStr0, 10)
	egldValue = egldValue.Add(egldValue, big.NewInt(13))
	tx = multiESDTNFTTransferWithEGLDTx(nonce, addrs[0].Bytes, addrs[1].Bytes, [][]byte{nftTokenID}, egldValue)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.NotEqual(t, "success", txResult.Status.String())

	eventLog := string(txResult.Logs.Events[0].Topics[1])
	require.Equal(t, "insufficient funds for token EGLD-000000", eventLog)

	// check accounts balance
	account0, err = cs.GetAccount(addrs[0])
	require.Nil(t, err)

	beforeBalance0, _ := big.NewInt(0).SetString(beforeBalanceStr0, 10)

	txsFee, _ := big.NewInt(0).SetString(txResult.Fee, 10)
	expectedBalanceWithFee0 := big.NewInt(0).Sub(beforeBalance0, txsFee)

	require.Equal(t, expectedBalanceWithFee0.String(), account0.Balance)

	account1, err = cs.GetAccount(addrs[1])
	require.Nil(t, err)

	require.Equal(t, beforeBalanceStr1, account1.Balance)
}

func TestChainSimulator_EGLD_MultiTransfer_Invalid_Value(t *testing.T) {
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

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	nonce := uint64(0)
	tx := issueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = esdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	account0, err := cs.GetAccount(addrs[0])
	require.Nil(t, err)

	beforeBalanceStr0 := account0.Balance

	account1, err := cs.GetAccount(addrs[1])
	require.Nil(t, err)

	beforeBalanceStr1 := account1.Balance

	egldValue := oneEGLD.Mul(oneEGLD, big.NewInt(3))
	tx = multiESDTNFTTransferWithEGLDTx(nonce, addrs[0].Bytes, addrs[1].Bytes, [][]byte{nftTokenID}, egldValue)
	tx.Value = egldValue // invalid value field

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.NotEqual(t, "success", txResult.Status.String())

	eventLog := string(txResult.Logs.Events[0].Topics[1])
	require.Equal(t, "built in function called with tx value is not allowed", eventLog)

	// check accounts balance
	account0, err = cs.GetAccount(addrs[0])
	require.Nil(t, err)

	beforeBalance0, _ := big.NewInt(0).SetString(beforeBalanceStr0, 10)

	txsFee, _ := big.NewInt(0).SetString(txResult.Fee, 10)
	expectedBalanceWithFee0 := big.NewInt(0).Sub(beforeBalance0, txsFee)

	require.Equal(t, expectedBalanceWithFee0.String(), account0.Balance)

	account1, err = cs.GetAccount(addrs[1])
	require.Nil(t, err)

	require.Equal(t, beforeBalanceStr1, account1.Balance)
}

func TestChainSimulator_Multiple_EGLD_Transfers(t *testing.T) {
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

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	nonce := uint64(0)
	tx := issueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = esdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	account0, err := cs.GetAccount(addrs[0])
	require.Nil(t, err)

	beforeBalanceStr0 := account0.Balance

	account1, err := cs.GetAccount(addrs[1])
	require.Nil(t, err)

	beforeBalanceStr1 := account1.Balance

	// multi nft transfer with multiple EGLD-000000 tokens
	numTransfers := 3
	encodedReceiver := hex.EncodeToString(addrs[1].Bytes)
	egldValue := oneEGLD.Mul(oneEGLD, big.NewInt(3))

	txDataField := []byte(strings.Join(
		[]string{
			core.BuiltInFunctionMultiESDTNFTTransfer,
			encodedReceiver,
			hex.EncodeToString(big.NewInt(int64(numTransfers)).Bytes()),
			hex.EncodeToString([]byte("EGLD-000000")),
			"00",
			hex.EncodeToString(egldValue.Bytes()),
			hex.EncodeToString(nftTokenID),
			hex.EncodeToString(big.NewInt(1).Bytes()),
			hex.EncodeToString(big.NewInt(int64(1)).Bytes()),
			hex.EncodeToString([]byte("EGLD-000000")),
			"00",
			hex.EncodeToString(egldValue.Bytes()),
		}, "@"),
	)

	tx = &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   addrs[0].Bytes,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Data:      txDataField,
		Value:     big.NewInt(0),
		Version:   1,
		Signature: []byte("dummySig"),
		ChainID:   []byte(configs.ChainID),
	}

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	// check accounts balance
	account0, err = cs.GetAccount(addrs[0])
	require.Nil(t, err)

	beforeBalance0, _ := big.NewInt(0).SetString(beforeBalanceStr0, 10)

	expectedBalance0 := big.NewInt(0).Sub(beforeBalance0, egldValue)
	expectedBalance0 = big.NewInt(0).Sub(expectedBalance0, egldValue)
	txsFee, _ := big.NewInt(0).SetString(txResult.Fee, 10)
	expectedBalanceWithFee0 := big.NewInt(0).Sub(expectedBalance0, txsFee)

	require.Equal(t, expectedBalanceWithFee0.String(), account0.Balance)

	account1, err = cs.GetAccount(addrs[1])
	require.Nil(t, err)

	beforeBalance1, _ := big.NewInt(0).SetString(beforeBalanceStr1, 10)
	expectedBalance1 := big.NewInt(0).Add(beforeBalance1, egldValue)
	expectedBalance1 = big.NewInt(0).Add(expectedBalance1, egldValue)

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

func TestChainSimulator_IssueToken_EGLDTicker(t *testing.T) {
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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	log.Info("Initial setup: Issue token (before the activation of EGLDInMultiTransferFlag)")

	// issue NFT
	nftTicker := []byte("EGLD")
	nonce := uint64(0)
	tx := issueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = esdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Issue token (after activation of EGLDInMultiTransferFlag)")

	// should fail issuing token with EGLD ticker
	tx = issueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	errMessage := string(txResult.Logs.Events[0].Topics[1])
	require.Equal(t, vm.ErrCouldNotCreateNewTokenIdentifier.Error(), errMessage)

	require.Equal(t, "success", txResult.Status.String())
}

func TestScCallTransferValueESDT(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   true,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              3,
		GenesisTimestamp:         time.Now().Unix(),
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpochOpt,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  3,
		NumNodesWaitingListShard: 3,

		InitialEpoch: 1700,
		InitialNonce: 1700,
		InitialRound: 1700,
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	nonce := uint64(0)
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address:          "erd1qqqqqqqqqqqqqpgqw7vrdzlqg4f8zja8qgpw2cdqpcp5xhvrvcqs984824",
			Nonce:            &nonce,
			Balance:          "0",
			Code:             "0061736d01000000016b116000017f60000060027f7f017f60027f7f0060017f0060017f017f60047f7f7f7f0060037f7f7f017f60027f7e0060037f7f7f0060047f7f7f7f017f60067e7f7f7f7f7f017f60057f7f7e7f7f017f6000017e60057f7f7f7f7f0060047f7f7f7e0060057e7f7f7f7f0002bd041803656e760a6d4275666665724e6577000003656e760d6d427566666572417070656e64000203656e760d6d616e6167656443616c6c6572000403656e76126d427566666572417070656e644279746573000703656e76126d616e616765645369676e616c4572726f72000403656e76126d427566666572476574417267756d656e74000203656e76106d4275666665724765744c656e677468000503656e7619626967496e74476574556e7369676e6564417267756d656e74000303656e760f6765744e756d417267756d656e7473000003656e760b7369676e616c4572726f72000303656e761b6d616e61676564457865637574654f6e44657374436f6e74657874000b03656e760f6d4275666665725365744279746573000703656e76196d42756666657246726f6d426967496e74556e7369676e6564000203656e760e626967496e74536574496e743634000803656e76106d616e61676564534341646472657373000403656e760e636865636b4e6f5061796d656e74000103656e761b6d616e616765645472616e7366657256616c756545786563757465000c03656e761c6d616e616765644765744d756c74694553445443616c6c56616c7565000403656e76096d4275666665724571000203656e7609626967496e74416464000903656e760a6765744761734c656674000d03656e760f636c65616e52657475726e44617461000103656e76136d42756666657253746f7261676553746f7265000203656e76136d42756666657247657442797465536c696365000a031e1d05000002040509000605030e05030a06060f081000030300010101010105030100030616037f01418080080b7f00419582080b7f0041a082080b075608066d656d6f7279020004696e697400300775706772616465003107666f7277617264003208726563656976656400330863616c6c4261636b00340a5f5f646174615f656e6403010b5f5f686561705f6261736503020a84151d0f01017f10002201200010011a20010b0c01017f101a2200100220000b1901017f419082084190820828020041016b220036020020000b1101017f101a220220002001100b1a20020b1400100820004604400f0b41d6800841191009000b2b01027f2000419482082d0000220171200041ff01714622024504404194820820002001723a00000b20020b180020012002101b21012000101f360204200020013602000b080041014100101b0b1e00101f1a200220032802001021102220002002360204200020013602000b0f01017f101a22012000100c1a20010b4601017f230041106b220224002002200141187420014180fe03714108747220014108764180fe03712001411876727236020c20002002410c6a410410031a200241106a24000b8e0101037f230041106b220524000240200310240d00200220031025200410062107410021030340200320074f0d012005410036020c200420032005410c6a410410261a2002200528020c220641187420064180fe03714108747220064108764180fe0371200641187672721025200341046a21030c000b000b2000200236020420002001360200200541106a24000b070020001006450b0d00101f1a20002001101810220b0f00200020012003200210174100470b1e00101f1a200220032802001018102220002002360204200020013602000b1b00101f1a200220031018102220002002360204200020013602000b2001017f101f22042003102a20022004102220002002360204200020013602000bff0102027f017e230041106b220324002003200142388620014280fe0383422886842001428080fc0783421886200142808080f80f834208868484200142088842808080f80f832001421888428080fc078384200142288822044280fe03832001423888848484370308200041002001428080808080808080015422002001423088a741ff01711b220220006a410020022004a741ff01711b22006a410020002001422088a741ff01711b22006a410020002001a722004118761b22026a41002002200041107641ff01711b22026a41002002200041087641ff01711b22006a200041002001501b6a2200200341086a6a410820006b100b1a200341106a24000b110020002001200220032004101a100a1a0b0a0041764200100d41760b7001037f230041106b22022400200020012802042204200128020849047f200241086a2203420037030020024200370300200128020020042002411010261a2001200441106a36020420002002290300370001200041096a200329030037000041010541000b3a0000200241106a24000bb90102017f017e2000200128000c220241187420024180fe03714108747220024108764180fe03712002411876727236020c20002001280000220241187420024180fe03714108747220024108764180fe03712002411876727236020820002001290004220342388620034280fe0383422886842003428080fc0783421886200342808080f80f834208868484200342088842808080f80f832003421888428080fc07838420034228884280fe038320034238888484843703000b0c01017f101a2200100e20000b0800100f4100101c0b3301037f100f4100101c1019210141e58108412a101b2100101f2102200010244504402001102c42a0c21e2000200210101a0b0b930b020a7f027e230041d0016b220024004101101c4100101a220110051a20011006412047044041bc80084117101b220041d58108411010031a200041d38008410310031a200041bb8108411010031a20001004000b200121034102101d450440415a10110b02404104101d0d00415841b18008410b100b1a2000415a100636029801200042daffffff0f370290010340200041b8016a20004190016a102d20002d00b8014101470d01415820002800b901220141187420014180fe03714108747220014108764180fe037120014118767272101241004c0d000b4199800841181009000b1019101821010240200341feffffff07470440200041f8006a41d181084104101e200041f0006a2000280278200028027c200110282000280274210520002802702106101f21022000415a100636028c01200042daffffff0f37028401200041c0016a210820004191016a2107034020004190016a20004184016a102d20002d0090014101460440200041b0016a200741086a290000370300200020072900003703a8012008200041a8016a102e20002903c001210a20002802cc01210920002802c80110182104101a22014200100d20012001200910132000200a423886200a4280fe038342288684200a428080fc0783421886200a42808080f80f834208868484200a42088842808080f80f83200a421888428080fc078384200a4228884280fe0383200a423888848484370294012000200441187420044180fe03714108747220044108764180fe037120044118767272360290012000200141187420014180fe03714108747220014108764180fe03712001411876727236029c01200220004190016a411010031a0c010b0b1014220a42a08d067d200a200a42a08d06561b210b0240024002400240200210064104760e020102000b200041186a41ef80084114101e20002802182104200028021c2101101f1a2001200310181022200210062103101f22072003410476ad102a20012007102220002002100636029801200041003602940120002002360290010340200041b8016a20004190016a102d20002d00b801410146044020002800c501210220002900bd01210a200120002800b901220341187420034180fe03714108747220034108764180fe0371200341187672721025200041086a20042001200a423886200a4280fe038342288684200a428080fc0783421886200a42808080f80f834208868484200a42088842808080f80f83200a421888428080fc078384200a4228884280fe0383200a423888848484102920002802082104200028020c2101101f1a2001200241187420024180fe03714108747220024108764180fe037120024118767272102110220c010b0b200041106a200420012006200510232000280214210120002802102102200b102f102c20022001102b0c020b200b2003102c20062005102b0c010b20004198016a420037030020004200370390012002410020004190016a411010260d02200041c0016a20004190016a102e200041b0016a2201200041c8016a290300370300200020002903c001220a3703a801200041b4016a2102200a500440200041386a41928108410c101e200041306a2000280238200028023c20011027200041286a2000280230200028023420021020200041206a2000280228200028022c2006200510232000280224210120002802202102200b2003102c20022001102b0c010b200041e8006a41838108410f101e200041e0006a2000280268200028026c20011027200041d8006a20002802602000280264200a1029200041d0006a2000280258200028025c20021020200041c8006a2000280250200028025420031028200041406b2000280248200028024c2006200510232000280244210120002802402102200b102f102c20022001102b0b1015200041d0016a24000f0b4180800841191009000b419e8108411d1009000b2101017f100f4101101c4100101a2200100741cb81084106101b2000102110161a0b0300010b0ba3020200418080080b8f02726563697069656e742061646472657373206e6f7420736574756e65787065637465642045474c44207472616e7366657245474c442d303030303030617267756d656e74206465636f6465206572726f722028293a2077726f6e67206e756d626572206f6620617267756d656e74734d756c7469455344544e46545472616e73666572455344544e46545472616e73666572455344545472616e736665724d616e6167656456656320696e646578206f7574206f662072616e6765626164206172726179206c656e677468616d6f756e747465737464756d6d795f73635f61646472657373455344545472616e7366657240353535333434343332443333333533303633333436354030463432343000419082080b0438ffffff",
			RootHash:         "",
			CodeMetadata:     "BQQ=",
			CodeHash:         "KQsZ3JMD3ojAR5MfScgQH0o3XLUpvP1H7flxxt0qe80=",
			DeveloperRewards: "4354035000000",
			Owner:            "erd1tkc62psh0flcj6anm6gt227gqqu7sp4xc3c3cc0fcmgk9ax6vcqs2w8h2s",
		},
		{
			Address: "erd1tkc62psh0flcj6anm6gt227gqqu7sp4xc3c3cc0fcmgk9ax6vcqs2w8h2s",
			Nonce:   &nonce,
			Balance: "191060078069461323",
		},
	})
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	pubKeyConverter := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()
	sndBech := "erd1tkc62psh0flcj6anm6gt227gqqu7sp4xc3c3cc0fcmgk9ax6vcqs2w8h2s"
	snd, _ := pubKeyConverter.Decode(sndBech)
	rcv, _ := pubKeyConverter.Decode("erd1qqqqqqqqqqqqqpgqw7vrdzlqg4f8zja8qgpw2cdqpcp5xhvrvcqs984824")

	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(0),
		SndAddr:   snd,
		RcvAddr:   rcv,
		Data:      []byte("upgradeContract@0061736d01000000016b116000017f60000060027f7f017f60027f7f0060017f0060017f017f60047f7f7f7f0060037f7f7f017f60027f7e0060037f7f7f0060047f7f7f7f017f60067e7f7f7f7f7f017f60057f7f7e7f7f017f6000017e60057f7f7f7f7f0060047f7f7f7e0060057e7f7f7f7f0002bd041803656e760a6d4275666665724e6577000003656e760d6d427566666572417070656e64000203656e760d6d616e6167656443616c6c6572000403656e76126d427566666572417070656e644279746573000703656e76126d616e616765645369676e616c4572726f72000403656e76126d427566666572476574417267756d656e74000203656e76106d4275666665724765744c656e677468000503656e7619626967496e74476574556e7369676e6564417267756d656e74000303656e760f6765744e756d417267756d656e7473000003656e760b7369676e616c4572726f72000303656e761b6d616e61676564457865637574654f6e44657374436f6e74657874000b03656e760f6d4275666665725365744279746573000703656e76196d42756666657246726f6d426967496e74556e7369676e6564000203656e760e626967496e74536574496e743634000803656e76106d616e61676564534341646472657373000403656e760e636865636b4e6f5061796d656e74000103656e761b6d616e616765645472616e7366657256616c756545786563757465000c03656e761c6d616e616765644765744d756c74694553445443616c6c56616c7565000403656e76096d4275666665724571000203656e7609626967496e74416464000903656e760a6765744761734c656674000d03656e760f636c65616e52657475726e44617461000103656e76136d42756666657253746f7261676553746f7265000203656e76136d42756666657247657442797465536c696365000a031e1d05000002040509000605030e05030a06060f081000030300010101010105030100030616037f01418080080b7f00419582080b7f0041a082080b075608066d656d6f7279020004696e697400300775706772616465003107666f7277617264003208726563656976656400330863616c6c4261636b00340a5f5f646174615f656e6403010b5f5f686561705f6261736503020a84151d0f01017f10002201200010011a20010b0c01017f101a2200100220000b1901017f419082084190820828020041016b220036020020000b1101017f101a220220002001100b1a20020b1400100820004604400f0b41d6800841191009000b2b01027f2000419482082d0000220171200041ff01714622024504404194820820002001723a00000b20020b180020012002101b21012000101f360204200020013602000b080041014100101b0b1e00101f1a200220032802001021102220002002360204200020013602000b0f01017f101a22012000100c1a20010b4601017f230041106b220224002002200141187420014180fe03714108747220014108764180fe03712001411876727236020c20002002410c6a410410031a200241106a24000b8e0101037f230041106b220524000240200310240d00200220031025200410062107410021030340200320074f0d012005410036020c200420032005410c6a410410261a2002200528020c220641187420064180fe03714108747220064108764180fe0371200641187672721025200341046a21030c000b000b2000200236020420002001360200200541106a24000b070020001006450b0d00101f1a20002001101810220b0f00200020012003200210174100470b1e00101f1a200220032802001018102220002002360204200020013602000b1b00101f1a200220031018102220002002360204200020013602000b2001017f101f22042003102a20022004102220002002360204200020013602000bff0102027f017e230041106b220324002003200142388620014280fe0383422886842001428080fc0783421886200142808080f80f834208868484200142088842808080f80f832001421888428080fc078384200142288822044280fe03832001423888848484370308200041002001428080808080808080015422002001423088a741ff01711b220220006a410020022004a741ff01711b22006a410020002001422088a741ff01711b22006a410020002001a722004118761b22026a41002002200041107641ff01711b22026a41002002200041087641ff01711b22006a200041002001501b6a2200200341086a6a410820006b100b1a200341106a24000b110020002001200220032004101a100a1a0b0a0041764200100d41760b7001037f230041106b22022400200020012802042204200128020849047f200241086a2203420037030020024200370300200128020020042002411010261a2001200441106a36020420002002290300370001200041096a200329030037000041010541000b3a0000200241106a24000bb90102017f017e2000200128000c220241187420024180fe03714108747220024108764180fe03712002411876727236020c20002001280000220241187420024180fe03714108747220024108764180fe03712002411876727236020820002001290004220342388620034280fe0383422886842003428080fc0783421886200342808080f80f834208868484200342088842808080f80f832003421888428080fc07838420034228884280fe038320034238888484843703000b0c01017f101a2200100e20000b0800100f4100101c0b3301037f100f4100101c1019210141e58108412a101b2100101f2102200010244504402001102c42a0c21e2000200210101a0b0b930b020a7f027e230041d0016b220024004101101c4100101a220110051a20011006412047044041bc80084117101b220041d58108411010031a200041d38008410310031a200041bb8108411010031a20001004000b200121034102101d450440415a10110b02404104101d0d00415841b18008410b100b1a2000415a100636029801200042daffffff0f370290010340200041b8016a20004190016a102d20002d00b8014101470d01415820002800b901220141187420014180fe03714108747220014108764180fe037120014118767272101241004c0d000b4199800841181009000b1019101821010240200341feffffff07470440200041f8006a41d181084104101e200041f0006a2000280278200028027c200110282000280274210520002802702106101f21022000415a100636028c01200042daffffff0f37028401200041c0016a210820004191016a2107034020004190016a20004184016a102d20002d0090014101460440200041b0016a200741086a290000370300200020072900003703a8012008200041a8016a102e20002903c001210a20002802cc01210920002802c80110182104101a22014200100d20012001200910132000200a423886200a4280fe038342288684200a428080fc0783421886200a42808080f80f834208868484200a42088842808080f80f83200a421888428080fc078384200a4228884280fe0383200a423888848484370294012000200441187420044180fe03714108747220044108764180fe037120044118767272360290012000200141187420014180fe03714108747220014108764180fe03712001411876727236029c01200220004190016a411010031a0c010b0b1014220a42a08d067d200a200a42a08d06561b210b0240024002400240200210064104760e020102000b200041186a41ef80084114101e20002802182104200028021c2101101f1a2001200310181022200210062103101f22072003410476ad102a20012007102220002002100636029801200041003602940120002002360290010340200041b8016a20004190016a102d20002d00b801410146044020002800c501210220002900bd01210a200120002800b901220341187420034180fe03714108747220034108764180fe0371200341187672721025200041086a20042001200a423886200a4280fe038342288684200a428080fc0783421886200a42808080f80f834208868484200a42088842808080f80f83200a421888428080fc078384200a4228884280fe0383200a423888848484102920002802082104200028020c2101101f1a2001200241187420024180fe03714108747220024108764180fe037120024118767272102110220c010b0b200041106a200420012006200510232000280214210120002802102102200b102f102c20022001102b0c020b200b2003102c20062005102b0c010b20004198016a420037030020004200370390012002410020004190016a411010260d02200041c0016a20004190016a102e200041b0016a2201200041c8016a290300370300200020002903c001220a3703a801200041b4016a2102200a500440200041386a41928108410c101e200041306a2000280238200028023c20011027200041286a2000280230200028023420021020200041206a2000280228200028022c2006200510232000280224210120002802202102200b2003102c20022001102b0c010b200041e8006a41838108410f101e200041e0006a2000280268200028026c20011027200041d8006a20002802602000280264200a1029200041d0006a2000280258200028025c20021020200041c8006a2000280250200028025420031028200041406b2000280248200028024c2006200510232000280244210120002802402102200b102f102c20022001102b0b1015200041d0016a24000f0b4180800841191009000b419e8108411d1009000b2101017f100f4101101c4100101a2200100741cb81084106101b2000102110161a0b0300010b0ba3020200418080080b8f02726563697069656e742061646472657373206e6f7420736574756e65787065637465642045474c44207472616e7366657245474c442d303030303030617267756d656e74206465636f6465206572726f722028293a2077726f6e67206e756d626572206f6620617267756d656e74734d756c7469455344544e46545472616e73666572455344544e46545472616e73666572455344545472616e736665724d616e6167656456656320696e646578206f7574206f662072616e6765626164206172726179206c656e677468616d6f756e747465737464756d6d795f73635f61646472657373455344545472616e7366657240353535333434343332443333333533303633333436354030463432343000419082080b0438ffffff@0504"),
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
		Signature: []byte("dummy"),
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	require.Equal(t, core.SignalErrorOperation, txResult.Logs.Events[0].Identifier)
	require.Equal(t, "transfer value on esdt call", string(txResult.Logs.Events[0].Topics[1]))
}
