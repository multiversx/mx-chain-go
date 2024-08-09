package vm

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"

	minGasPrice                            = 1000000000
	maxNumOfBlockToGenerateWhenExecutingTx = 7
)

var oneEGLD = big.NewInt(1000000000000000000)

var log = logger.GetOrCreate("integrationTests/chainSimulator/vm")

// Test scenario #1
//
// Initial setup: Create fungible, NFT,  SFT and metaESDT tokens
// (before the activation of DynamicEsdtFlag)
//
// 1.check that the metadata for all tokens is saved on the system account
// 2. wait for DynamicEsdtFlag activation
// 3. transfer the tokens to another account
// 4. check that the metadata for all tokens is saved on the system account
// 5. make an updateTokenID@tokenID function call on the ESDTSystem SC for all token types
// 6. check that the metadata for all tokens is saved on the system account
// 7. transfer the tokens to another account
// 8. check that the metaData for the NFT was removed from the system account and moved to the user account
// 9. check that the metaData for the rest of the tokens is still present on the system account and not on the userAccount
// 10. do the test for both intra and cross shard txs
func TestChainSimulator_CheckTokensMetadata_TransferTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("transfer and check all tokens - intra shard", func(t *testing.T) {
		transferAndCheckTokensMetaData(t, false, false)
	})

	t.Run("transfer and check all tokens - intra shard - multi transfer", func(t *testing.T) {
		transferAndCheckTokensMetaData(t, false, true)
	})

	t.Run("transfer and check all tokens - cross shard", func(t *testing.T) {
		transferAndCheckTokensMetaData(t, true, false)
	})

	t.Run("transfer and check all tokens - cross shard - multi transfer", func(t *testing.T) {
		transferAndCheckTokensMetaData(t, true, true)
	})
}

func transferAndCheckTokensMetaData(t *testing.T, isCrossShard bool, isMultiTransfer bool) {
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := createAddresses(t, cs, isCrossShard)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	log.Info("Initial setup: Create NFT, SFT and metaESDT tokens (before the activation of DynamicEsdtFlag)")

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

	rolesTransfer := [][]byte{[]byte(core.ESDTRoleTransfer)}
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, metaESDTTokenID, rolesTransfer)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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

	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID, rolesTransfer)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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

	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, sftTokenID, rolesTransfer)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
		tx = nftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Step 1. check that the metadata for all tokens is saved on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
	checkMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	checkMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)

	log.Info("Step 2. wait for DynamicEsdtFlag activation")

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	log.Info("Step 3. transfer the tokens to another account")

	if isMultiTransfer {
		tx = multiESDTNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, tokenIDs)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	} else {
		for _, tokenID := range tokenIDs {
			log.Info("transfering token id", "tokenID", tokenID)

			tx = esdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
			txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
			require.Nil(t, err)
			require.NotNil(t, txResult)
			require.Equal(t, "success", txResult.Status.String())

			nonce++
		}
	}

	log.Info("Step 4. check that the metadata for all tokens is saved on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
	checkMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	checkMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)

	log.Info("Step 5. make an updateTokenID@tokenID function call on the ESDTSystem SC for all token types")

	for _, tokenID := range tokenIDs {
		tx = updateTokenIDTx(nonce, addrs[0].Bytes, tokenID)

		log.Info("updating token id", "tokenID", tokenID)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	log.Info("Step 6. check that the metadata for all tokens is saved on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
	checkMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	checkMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)

	log.Info("Step 7. transfer the tokens to another account")

	nonce = uint64(0)
	if isMultiTransfer {
		tx = multiESDTNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, tokenIDs)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())
	} else {
		for _, tokenID := range tokenIDs {
			log.Info("transfering token id", "tokenID", tokenID)

			tx = esdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, tokenID)
			txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
			require.Nil(t, err)
			require.NotNil(t, txResult)
			require.Equal(t, "success", txResult.Status.String())

			nonce++
		}
	}

	log.Info("Step 8. check that the metaData for the NFT was removed from the system account and moved to the user account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)

	checkMetaData(t, cs, addrs[2].Bytes, nftTokenID, shardID, nftMetaData)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	log.Info("Step 9. check that the metaData for the rest of the tokens is still present on the system account and not on the userAccount")

	checkMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[2].Bytes, sftTokenID, shardID)

	checkMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[2].Bytes, metaESDTTokenID, shardID)
}

func createAddresses(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	isCrossShard bool,
) []dtos.WalletAddress {
	var shardIDs []uint32
	if !isCrossShard {
		shardIDs = []uint32{1, 1, 1}
	} else {
		shardIDs = []uint32{0, 1, 2}
	}

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(oneEGLD, mintValue)

	address, err := cs.GenerateAndMintWalletAddress(shardIDs[0], mintValue)
	require.Nil(t, err)

	address2, err := cs.GenerateAndMintWalletAddress(shardIDs[1], mintValue)
	require.Nil(t, err)

	address3, err := cs.GenerateAndMintWalletAddress(shardIDs[2], mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	return []dtos.WalletAddress{address, address2, address3}
}

func checkMetaData(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
	expectedMetaData *txsFee.MetaData,
) {
	retrievedMetaData := getMetaDataFromAcc(t, cs, addressBytes, token, shardID)

	require.Equal(t, expectedMetaData.Nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, expectedMetaData.Name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, expectedMetaData.Royalties, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Royalties)).Bytes())))
	require.Equal(t, expectedMetaData.Hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expectedMetaData.Uris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, expectedMetaData.Attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))
}

func checkMetaDataNotInAcc(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
) {
	esdtData := getESDTDataFromAcc(t, cs, addressBytes, token, shardID)

	require.Nil(t, esdtData.TokenMetaData)
}

func multiESDTNFTTransferTx(nonce uint64, sndAdr, rcvAddr []byte, tokens [][]byte) *transaction.Transaction {
	transferData := make([]*utils.TransferESDTData, 0)

	for _, tokenID := range tokens {
		transferData = append(transferData, &utils.TransferESDTData{
			Token: tokenID,
			Nonce: 1,
			Value: big.NewInt(1),
		})
	}

	tx := utils.CreateMultiTransferTX(
		nonce,
		sndAdr,
		rcvAddr,
		minGasPrice,
		10_000_000,
		transferData...,
	)
	tx.Version = 1
	tx.Signature = []byte("dummySig")
	tx.ChainID = []byte(configs.ChainID)

	return tx
}

func esdtNFTTransferTx(nonce uint64, sndAdr, rcvAddr, token []byte) *transaction.Transaction {
	tx := utils.CreateESDTNFTTransferTx(
		nonce,
		sndAdr,
		rcvAddr,
		token,
		1,
		big.NewInt(1),
		minGasPrice,
		10_000_000,
		"",
	)
	tx.Version = 1
	tx.Signature = []byte("dummySig")
	tx.ChainID = []byte(configs.ChainID)

	return tx
}

func issueTx(nonce uint64, sndAdr []byte, ticker []byte, baseIssuingCost string) *transaction.Transaction {
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issue"),
			[]byte(hex.EncodeToString([]byte("asdname1"))),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func issueMetaESDTTx(nonce uint64, sndAdr []byte, ticker []byte, baseIssuingCost string) *transaction.Transaction {
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerMetaESDT"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func issueNonFungibleTx(nonce uint64, sndAdr []byte, ticker []byte, baseIssuingCost string) *transaction.Transaction {
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issueNonFungible"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString(ticker)),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func issueSemiFungibleTx(nonce uint64, sndAdr []byte, ticker []byte, baseIssuingCost string) *transaction.Transaction {
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issueSemiFungible"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString(ticker)),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func changeToDynamicTx(nonce uint64, sndAdr []byte, tokenID []byte) *transaction.Transaction {
	txDataField := []byte("changeToDynamic@" + hex.EncodeToString(tokenID))

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func updateTokenIDTx(nonce uint64, sndAdr []byte, tokenID []byte) *transaction.Transaction {
	txDataField := []byte("updateTokenID@" + hex.EncodeToString(tokenID))

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func nftCreateTx(
	nonce uint64,
	sndAdr []byte,
	tokenID []byte,
	metaData *txsFee.MetaData,
) *transaction.Transaction {
	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())), // quantity
			metaData.Name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			metaData.Hash,
			metaData.Attributes,
			metaData.Uris[0],
			metaData.Uris[1],
			metaData.Uris[2],
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   sndAdr,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func modifyCreatorTx(
	nonce uint64,
	sndAdr []byte,
	tokenID []byte,
) *transaction.Transaction {
	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.ESDTModifyCreator),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   sndAdr,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func getESDTDataFromAcc(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
) *esdt.ESDigitalToken {
	account, err := cs.GetNodeHandler(shardID).GetStateComponents().AccountsAdapter().LoadAccount(addressBytes)
	require.Nil(t, err)
	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	baseEsdtKeyPrefix := core.ProtectedKeyPrefix + core.ESDTKeyIdentifier
	key := append([]byte(baseEsdtKeyPrefix), token...)

	key2 := append(key, big.NewInt(0).SetUint64(1).Bytes()...)
	esdtDataBytes, _, err := userAccount.RetrieveValue(key2)
	require.Nil(t, err)

	esdtData := &esdt.ESDigitalToken{}
	err = cs.GetNodeHandler(shardID).GetCoreComponents().InternalMarshalizer().Unmarshal(esdtData, esdtDataBytes)
	require.Nil(t, err)

	return esdtData
}

func getMetaDataFromAcc(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
) *esdt.MetaData {
	esdtData := getESDTDataFromAcc(t, cs, addressBytes, token, shardID)

	require.NotNil(t, esdtData.TokenMetaData)

	return esdtData.TokenMetaData
}

func setSpecialRoleTx(
	nonce uint64,
	sndAddr []byte,
	address []byte,
	token []byte,
	roles [][]byte,
) *transaction.Transaction {
	txDataBytes := [][]byte{
		[]byte("setSpecialRole"),
		[]byte(hex.EncodeToString(token)),
		[]byte(hex.EncodeToString(address)),
	}

	for _, role := range roles {
		txDataBytes = append(txDataBytes, []byte(hex.EncodeToString(role)))
	}

	txDataField := bytes.Join(
		txDataBytes,
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAddr,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  60_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func setAddressEsdtRoles(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	nonce uint64,
	address dtos.WalletAddress,
	token []byte,
	roles [][]byte,
) {
	tx := setSpecialRoleTx(nonce, address.Bytes, address.Bytes, token, roles)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())
}

// Test scenario #3
//
// Initial setup: Create NFT,  SFT and metaESDT tokens
// (after the activation of DynamicEsdtFlag)
//
// 1. check that the metaData for the NFT was saved in the user account and not on the system account
// 2. check that the metaData for the other token types is saved on the system account and not at the user account level
func TestChainSimulator_CreateTokensAfterActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, false)

	log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

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

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

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

	log.Info("Step 1. check that the metaData for the NFT was saved in the user account and not on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	log.Info("Step 2. check that the metaData for the other token types is saved on the system account and not at the user account level")

	checkMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)

	checkMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID)
}

// Test scenario #4
//
// Initial setup: Create NFT, SFT, metaESDT tokens
//
// Call ESDTMetaDataRecreate to rewrite the meta data for the nft
// (The sender must have the ESDTMetaDataRecreate role)
func TestChainSimulator_ESDTMetaDataRecreate(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	addrs := createAddresses(t, cs, false)

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
	}
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	for i := range tokenIDs {
		tx = nftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = changeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleNFTRecreate),
		}
		tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Call ESDTMetaDataRecreate to rewrite the meta data for the nft")

	for i := range tokenIDs {
		newMetaData := txsFee.GetDefaultMetaData()
		newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
		newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
		newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
		newMetaData.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))

		txDataField := bytes.Join(
			[][]byte{
				[]byte(core.ESDTMetaDataRecreate),
				[]byte(hex.EncodeToString(tokenIDs[i])),
				newMetaData.Nonce,
				newMetaData.Name,
				[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
				newMetaData.Hash,
				newMetaData.Attributes,
				newMetaData.Uris[0],
				newMetaData.Uris[1],
				newMetaData.Uris[2],
			},
			[]byte("@"),
		)

		tx = &transaction.Transaction{
			Nonce:     nonce,
			SndAddr:   addrs[0].Bytes,
			RcvAddr:   addrs[0].Bytes,
			GasLimit:  10_000_000,
			GasPrice:  minGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			checkMetaData(t, cs, addrs[0].Bytes, tokenIDs[i], shardID, newMetaData)
		} else {
			checkMetaData(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID, newMetaData)
		}

		nonce++
	}
}

// Test scenario #5
//
// Initial setup: Create NFT, SFT, metaESDT tokens
//
// Call ESDTMetaDataUpdate to update some of the meta data parameters
// (The sender must have the ESDTRoleNFTUpdate role)
func TestChainSimulator_ESDTMetaDataUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	addrs := createAddresses(t, cs, false)

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	for i := range tokenIDs {
		tx = nftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = changeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleNFTUpdate),
		}
		tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	log.Info("Call ESDTMetaDataUpdate to rewrite the meta data for the nft")

	for i := range tokenIDs {
		newMetaData := txsFee.GetDefaultMetaData()
		newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
		newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
		newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
		newMetaData.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))

		txDataField := bytes.Join(
			[][]byte{
				[]byte(core.ESDTMetaDataUpdate),
				[]byte(hex.EncodeToString(tokenIDs[i])),
				newMetaData.Nonce,
				newMetaData.Name,
				[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
				newMetaData.Hash,
				newMetaData.Attributes,
				newMetaData.Uris[0],
				newMetaData.Uris[1],
				newMetaData.Uris[2],
			},
			[]byte("@"),
		)

		tx = &transaction.Transaction{
			Nonce:     nonce,
			SndAddr:   addrs[0].Bytes,
			RcvAddr:   addrs[0].Bytes,
			GasLimit:  10_000_000,
			GasPrice:  minGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			checkMetaData(t, cs, addrs[0].Bytes, tokenIDs[i], shardID, newMetaData)
		} else {
			checkMetaData(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID, newMetaData)
		}

		nonce++
	}
}

// Test scenario #6
//
// Initial setup: Create NFT, SFT, metaESDT tokens
//
// Call ESDTModifyCreator and check that the creator was modified
// (The sender must have the ESDTRoleModifyCreator role)
func TestChainSimulator_ESDTModifyCreator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag). Register NFT directly as dynamic")

	addrs := createAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := issueMetaESDTTx(nonce, addrs[1].Bytes, metaESDTTicker, baseIssuingCost)
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
	tx = setSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// register dynamic NFT
	nftTicker := []byte("NFTTICKER")
	nftTokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(nftTokenName)),
			[]byte(hex.EncodeToString(nftTicker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	tx = &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[1].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = setSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = issueSemiFungibleTx(nonce, addrs[1].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = setSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	for i := range tokenIDs {
		tx = nftCreateTx(nonce, addrs[1].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	log.Info("Change to DYNAMIC type")

	for i := range tokenIDs {
		tx = changeToDynamicTx(nonce, addrs[1].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	log.Info("Call ESDTModifyCreator and check that the creator was modified")

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(oneEGLD, mintValue)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	for i := range tokenIDs {
		log.Info("Modify creator for token", "tokenID", tokenIDs[i])

		newCreatorAddress, err := cs.GenerateAndMintWalletAddress(shardID, mintValue)
		require.Nil(t, err)

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		roles = [][]byte{
			[]byte(core.ESDTRoleModifyCreator),
		}
		tx = setSpecialRoleTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i], roles)
		nonce++

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		tx = modifyCreatorTx(0, newCreatorAddress.Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		retrievedMetaData := getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)

		require.Equal(t, newCreatorAddress.Bytes, retrievedMetaData.Creator)
	}
}

func TestChainSimulator_ESDTModifyCreator_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := issueMetaESDTTx(nonce, addrs[1].Bytes, metaESDTTicker, baseIssuingCost)
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
	tx = setSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// register dynamic NFT
	nftTicker := []byte("NFTTICKER")
	nftTokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(nftTokenName)),
			[]byte(hex.EncodeToString(nftTicker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	tx = &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[1].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = setSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = issueSemiFungibleTx(nonce, addrs[1].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = setSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	for i := range tokenIDs {
		tx = nftCreateTx(nonce, addrs[1].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	log.Info("Change to DYNAMIC type")

	for i := range tokenIDs {
		tx = changeToDynamicTx(nonce, addrs[1].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	log.Info("Call ESDTModifyCreator and check that the creator was modified")

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(oneEGLD, mintValue)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	crossShardID := uint32(2)
	if shardID == uint32(2) {
		crossShardID = uint32(1)
	}

	for i := range tokenIDs {
		log.Info("Modify creator for token", "tokenID", string(tokenIDs[i]))

		newCreatorAddress, err := cs.GenerateAndMintWalletAddress(crossShardID, mintValue)
		require.Nil(t, err)

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		roles = [][]byte{
			[]byte(core.ESDTRoleModifyCreator),
		}
		tx = setSpecialRoleTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i], roles)
		nonce++

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		log.Info("transfering token id", "tokenID", tokenIDs[i])

		tx = esdtNFTTransferTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i])
		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		tx = modifyCreatorTx(0, newCreatorAddress.Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(newCreatorAddress.Bytes)
		retrievedMetaData := getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)

		require.Equal(t, newCreatorAddress.Bytes, retrievedMetaData.Creator)

		nonce++
	}
}

// Test scenario #7
//
// Initial setup: Create NFT, SFT, metaESDT tokens
//
// Call ESDTSetNewURIs and check that the new URIs were set for the token
// (The sender must have the ESDTRoleSetNewURI role)
func TestChainSimulator_ESDTSetNewURIs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, false)

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	for i := range tokenIDs {
		tx = nftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = changeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleSetNewURI),
		}
		tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	log.Info("Call ESDTSetNewURIs and check that the new URIs were set for the tokens")

	metaDataNonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	uris := [][]byte{
		[]byte(hex.EncodeToString([]byte("uri0"))),
		[]byte(hex.EncodeToString([]byte("uri1"))),
		[]byte(hex.EncodeToString([]byte("uri2"))),
	}

	expUris := [][]byte{
		[]byte("uri0"),
		[]byte("uri1"),
		[]byte("uri2"),
	}

	for i := range tokenIDs {
		log.Info("Set new uris for token", "tokenID", string(tokenIDs[i]))

		txDataField := bytes.Join(
			[][]byte{
				[]byte(core.ESDTSetNewURIs),
				[]byte(hex.EncodeToString(tokenIDs[i])),
				metaDataNonce,
				uris[0],
				uris[1],
				uris[2],
			},
			[]byte("@"),
		)

		tx = &transaction.Transaction{
			Nonce:     nonce,
			SndAddr:   addrs[0].Bytes,
			RcvAddr:   addrs[0].Bytes,
			GasLimit:  10_000_000,
			GasPrice:  minGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
		var retrievedMetaData *esdt.MetaData
		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			retrievedMetaData = getMetaDataFromAcc(t, cs, addrs[0].Bytes, tokenIDs[i], shardID)
		} else {
			retrievedMetaData = getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)
		}

		require.Equal(t, expUris, retrievedMetaData.URIs)

		nonce++
	}
}

// Test scenario #8
//
// Initial setup: Create NFT, SFT, metaESDT tokens
//
// Call ESDTModifyRoyalties and check that the royalties were changed
// (The sender must have the ESDTRoleModifyRoyalties role)
func TestChainSimulator_ESDTModifyRoyalties(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, false)

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

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
	tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	for i := range tokenIDs {
		tx = nftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = changeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleModifyRoyalties),
		}
		tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	log.Info("Call ESDTModifyRoyalties and check that the royalties were changed")

	metaDataNonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	royalties := []byte(hex.EncodeToString(big.NewInt(20).Bytes()))

	for i := range tokenIDs {
		log.Info("Set new royalties for token", "tokenID", string(tokenIDs[i]))

		txDataField := bytes.Join(
			[][]byte{
				[]byte(core.ESDTModifyRoyalties),
				[]byte(hex.EncodeToString(tokenIDs[i])),
				metaDataNonce,
				royalties,
			},
			[]byte("@"),
		)

		tx = &transaction.Transaction{
			Nonce:     nonce,
			SndAddr:   addrs[0].Bytes,
			RcvAddr:   addrs[0].Bytes,
			GasLimit:  10_000_000,
			GasPrice:  minGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(addrs[0].Bytes)
		retrievedMetaData := getMetaDataFromAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)

		require.Equal(t, uint32(big.NewInt(20).Uint64()), retrievedMetaData.Royalties)

		nonce++
	}
}

// Test scenario #9
//
// Initial setup: Create NFT
//
// 1. Change the nft to DYNAMIC type - the metadata should be on the system account
// 2. Send the NFT cross shard
// 3. The meta data should still be present on the system account
func TestChainSimulator_NFT_ChangeToDynamicType(t *testing.T) {
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 2)
	require.Nil(t, err)

	log.Info("Initial setup: Create NFT")

	nftTicker := []byte("NFTTICKER")
	nonce := uint64(0)
	tx := issueNonFungibleTx(nonce, addrs[1].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	setAddressEsdtRoles(t, cs, nonce, addrs[1], nftTokenID, roles)
	nonce++

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = nftCreateTx(nonce, addrs[1].Bytes, nftTokenID, nftMetaData)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	log.Info("Step 1. Change the nft to DYNAMIC type - the metadata should be on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	tx = changeToDynamicTx(nonce, addrs[1].Bytes, nftTokenID)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	roles = [][]byte{
		[]byte(core.ESDTRoleNFTUpdate),
	}

	setAddressEsdtRoles(t, cs, nonce, addrs[1], nftTokenID, roles)
	nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	log.Info("Step 2. Send the NFT cross shard")

	tx = esdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Step 3. The meta data should still be present on the system account")

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
}

// Test scenario #10
//
// Initial setup: Create SFT and send in another shard
//
// 1. change the sft meta data (differently from the previous one) in the other shard
// 2. check that the newest metadata is saved
func TestChainSimulator_ChangeMetaData(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("sft change metadata", func(t *testing.T) {
		testChainSimulatorChangeMetaData(t, issueSemiFungibleTx)
	})

	t.Run("metaESDT change metadata", func(t *testing.T) {
		testChainSimulatorChangeMetaData(t, issueMetaESDTTx)
	})
}

type issueTxFunc func(uint64, []byte, []byte, string) *transaction.Transaction

func testChainSimulatorChangeMetaData(t *testing.T, issueFn issueTxFunc) {
	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	log.Info("Initial setup: Create token and send in another shard")

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTAddQuantity),
	}

	ticker := []byte("TICKER")
	nonce := uint64(0)
	tx := issueFn(nonce, addrs[1].Bytes, ticker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	setAddressEsdtRoles(t, cs, nonce, addrs[1], tokenID, roles)
	nonce++

	log.Info("Issued token id", "tokenID", string(tokenID))

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(2).Bytes())), // quantity
			metaData.Name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			metaData.Hash,
			metaData.Attributes,
			metaData.Uris[0],
			metaData.Uris[1],
			metaData.Uris[2],
		},
		[]byte("@"),
	)

	tx = &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[1].Bytes,
		RcvAddr:   addrs[1].Bytes,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	tx = changeToDynamicTx(nonce, addrs[1].Bytes, tokenID)
	nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Send to separate shards")

	tx = esdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, tokenID)
	nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	tx = esdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[0].Bytes, tokenID)
	nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	roles = [][]byte{
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	tx = setSpecialRoleTx(nonce, addrs[1].Bytes, addrs[0].Bytes, tokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Step 1. change the sft meta data in one shard")

	sftMetaData2 := txsFee.GetDefaultMetaData()
	sftMetaData2.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData2.Name = []byte(hex.EncodeToString([]byte("name2")))
	sftMetaData2.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	sftMetaData2.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))

	txDataField = bytes.Join(
		[][]byte{
			[]byte(core.ESDTMetaDataUpdate),
			[]byte(hex.EncodeToString(tokenID)),
			sftMetaData2.Nonce,
			sftMetaData2.Name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			sftMetaData2.Hash,
			sftMetaData2.Attributes,
			sftMetaData2.Uris[0],
			sftMetaData2.Uris[1],
			sftMetaData2.Uris[2],
		},
		[]byte("@"),
	)

	tx = &transaction.Transaction{
		Nonce:     0,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   addrs[0].Bytes,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("Step 2. check that the newest metadata is saved")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
	checkMetaData(t, cs, core.SystemAccountAddress, tokenID, shardID, sftMetaData2)

	shard2ID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)
	checkMetaData(t, cs, core.SystemAccountAddress, tokenID, shard2ID, metaData)

	log.Info("Step 3. create new wallet is shard 2")

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(oneEGLD, mintValue)
	newShard2Addr, err := cs.GenerateAndMintWalletAddress(2, mintValue)
	require.Nil(t, err)
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	log.Info("Step 4. send updated token to shard 2 ")

	tx = esdtNFTTransferTx(1, addrs[0].Bytes, newShard2Addr.Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	err = cs.GenerateBlocks(5)

	log.Info("Step 5. check meta data in shard 2 is updated to latest version ")

	checkMetaData(t, cs, core.SystemAccountAddress, tokenID, shard2ID, sftMetaData2)
}

func TestChainSimulator_NFT_RegisterDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	log.Info("Register dynamic nft token")

	nftTicker := []byte("NFTTICKER")
	nftTokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(nftTokenName)),
			[]byte(hex.EncodeToString(nftTicker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
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

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = nftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	log.Info("Check that token type is Dynamic")

	scQuery := &process.SCQuery{
		ScAddress: vm.ESDTSCAddress,
		FuncName:  "getTokenProperties",
		CallValue: big.NewInt(0),
		Arguments: [][]byte{nftTokenID},
	}
	result, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	tokenType := result.ReturnData[1]
	require.Equal(t, core.DynamicNFTESDT, string(tokenType))
}

func TestChainSimulator_MetaESDT_RegisterDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	log.Info("Register dynamic metaESDT token")

	metaTicker := []byte("METATICKER")
	metaTokenName := []byte("tokenName")

	decimals := big.NewInt(15)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(metaTokenName)),
			[]byte(hex.EncodeToString(metaTicker)),
			[]byte(hex.EncodeToString([]byte("META"))),
			[]byte(hex.EncodeToString(decimals.Bytes())),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
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

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = nftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	log.Info("Check that token type is Dynamic")

	scQuery := &process.SCQuery{
		ScAddress: vm.ESDTSCAddress,
		FuncName:  "getTokenProperties",
		CallValue: big.NewInt(0),
		Arguments: [][]byte{nftTokenID},
	}
	result, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	tokenType := result.ReturnData[1]
	require.Equal(t, core.Dynamic+core.MetaESDT, string(tokenType))
}

func TestChainSimulator_FNG_RegisterDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	log.Info("Register dynamic fungible token")

	metaTicker := []byte("FNGTICKER")
	metaTokenName := []byte("tokenName")

	decimals := big.NewInt(15)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(metaTokenName)),
			[]byte(hex.EncodeToString(metaTicker)),
			[]byte(hex.EncodeToString([]byte("FNG"))),
			[]byte(hex.EncodeToString(decimals.Bytes())),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	tx := &transaction.Transaction{
		Nonce:     0,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	signalErrorTopic := string(txResult.Logs.Events[0].Topics[1])

	require.Equal(t, fmt.Sprintf("cannot create %s tokens as dynamic", core.FungibleESDT), signalErrorTopic)
}

func TestChainSimulator_NFT_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	log.Info("Register dynamic nft token")

	nftTicker := []byte("NFTTICKER")
	nftTokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerAndSetAllRolesDynamic"),
			[]byte(hex.EncodeToString(nftTokenName)),
			[]byte(hex.EncodeToString(nftTicker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
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

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = nftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	log.Info("Check that token type is Dynamic")

	scQuery := &process.SCQuery{
		ScAddress: vm.ESDTSCAddress,
		FuncName:  "getTokenProperties",
		CallValue: big.NewInt(0),
		Arguments: [][]byte{nftTokenID},
	}
	result, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	tokenType := result.ReturnData[1]
	require.Equal(t, core.DynamicNFTESDT, string(tokenType))

	log.Info("Check token roles")

	scQuery = &process.SCQuery{
		ScAddress: vm.ESDTSCAddress,
		FuncName:  "getAllAddressesAndRoles",
		CallValue: big.NewInt(0),
		Arguments: [][]byte{nftTokenID},
	}
	result, _, err = cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	expectedRoles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTBurn),
		[]byte(core.ESDTRoleNFTUpdateAttributes),
		[]byte(core.ESDTRoleNFTAddURI),
		[]byte(core.ESDTRoleNFTRecreate),
		[]byte(core.ESDTRoleModifyCreator),
		[]byte(core.ESDTRoleModifyRoyalties),
		[]byte(core.ESDTRoleSetNewURI),
		[]byte(core.ESDTRoleNFTUpdate),
	}

	checkTokenRoles(t, result.ReturnData, expectedRoles)
}

func TestChainSimulator_SFT_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	log.Info("Register dynamic sft token")

	sftTicker := []byte("SFTTICKER")
	sftTokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerAndSetAllRolesDynamic"),
			[]byte(hex.EncodeToString(sftTokenName)),
			[]byte(hex.EncodeToString(sftTicker)),
			[]byte(hex.EncodeToString([]byte("SFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = nftCreateTx(nonce, addrs[0].Bytes, sftTokenID, nftMetaData)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, nftMetaData)

	log.Info("Check that token type is Dynamic")

	scQuery := &process.SCQuery{
		ScAddress: vm.ESDTSCAddress,
		FuncName:  "getTokenProperties",
		CallValue: big.NewInt(0),
		Arguments: [][]byte{sftTokenID},
	}
	result, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	tokenType := result.ReturnData[1]
	require.Equal(t, core.DynamicSFTESDT, string(tokenType))

	log.Info("Check token roles")

	scQuery = &process.SCQuery{
		ScAddress: vm.ESDTSCAddress,
		FuncName:  "getAllAddressesAndRoles",
		CallValue: big.NewInt(0),
		Arguments: [][]byte{sftTokenID},
	}
	result, _, err = cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	expectedRoles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTBurn),
		[]byte(core.ESDTRoleNFTUpdateAttributes),
		[]byte(core.ESDTRoleNFTAddURI),
		[]byte(core.ESDTRoleNFTRecreate),
		[]byte(core.ESDTRoleModifyCreator),
		[]byte(core.ESDTRoleModifyRoyalties),
		[]byte(core.ESDTRoleSetNewURI),
		[]byte(core.ESDTRoleNFTUpdate),
	}

	checkTokenRoles(t, result.ReturnData, expectedRoles)
}

func TestChainSimulator_FNG_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	log.Info("Register dynamic fungible token")

	fngTicker := []byte("FNGTICKER")
	fngTokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerAndSetAllRolesDynamic"),
			[]byte(hex.EncodeToString(fngTokenName)),
			[]byte(hex.EncodeToString(fngTicker)),
			[]byte(hex.EncodeToString([]byte("FNG"))),
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	tx := &transaction.Transaction{
		Nonce:     0,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	signalErrorTopic := string(txResult.Logs.Events[0].Topics[1])

	require.Equal(t, fmt.Sprintf("cannot create %s tokens as dynamic", core.FungibleESDT), signalErrorTopic)
}

func TestChainSimulator_MetaESDT_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	log.Info("Register dynamic meta esdt token")

	ticker := []byte("META" + "TICKER")
	tokenName := []byte("tokenName")

	decimals := big.NewInt(10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerAndSetAllRolesDynamic"),
			[]byte(hex.EncodeToString(tokenName)),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString([]byte("META"))),
			[]byte(hex.EncodeToString(decimals.Bytes())),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	metaTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, nonce, addrs[0], metaTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = nftCreateTx(nonce, addrs[0].Bytes, metaTokenID, nftMetaData)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, metaTokenID, shardID, nftMetaData)

	log.Info("Check that token type is Dynamic")

	scQuery := &process.SCQuery{
		ScAddress: vm.ESDTSCAddress,
		FuncName:  "getTokenProperties",
		CallValue: big.NewInt(0),
		Arguments: [][]byte{metaTokenID},
	}
	result, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	tokenType := result.ReturnData[1]
	require.Equal(t, core.Dynamic+core.MetaESDT, string(tokenType))

	log.Info("Check token roles")

	scQuery = &process.SCQuery{
		ScAddress: vm.ESDTSCAddress,
		FuncName:  "getAllAddressesAndRoles",
		CallValue: big.NewInt(0),
		Arguments: [][]byte{metaTokenID},
	}
	result, _, err = cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	expectedRoles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTBurn),
		[]byte(core.ESDTRoleNFTAddQuantity),
		[]byte(core.ESDTRoleNFTUpdateAttributes),
		[]byte(core.ESDTRoleNFTAddURI),
	}

	checkTokenRoles(t, result.ReturnData, expectedRoles)
}

func checkTokenRoles(t *testing.T, returnData [][]byte, expectedRoles [][]byte) {
	for _, expRole := range expectedRoles {
		found := false

		for _, item := range returnData {
			if bytes.Equal(expRole, item) {
				found = true
			}
		}

		require.True(t, found)
	}
}

func TestChainSimulator_NFTcreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := getTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, false)

	log.Info("Initial setup: Create NFT that will have it's metadata saved to the user account")

	nftTicker := []byte("NFTTICKER")
	tx := issueNonFungibleTx(0, addrs[0].Bytes, nftTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	nftTokenID := txResult.Logs.Events[0].Topics[0]

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	createTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, nftTokenID, nftMetaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
	checkMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_SFTcreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := getTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, false)

	log.Info("Initial setup: Create SFT that will have it's metadata saved to the user account")

	sftTicker := []byte("SFTTICKER")
	tx := issueSemiFungibleTx(0, addrs[0].Bytes, sftTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	sftTokenID := txResult.Logs.Events[0].Topics[0]

	log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	createTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, sftTokenID, metaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, metaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)
	checkMetaDataNotInAcc(t, cs, addrs[1].Bytes, sftTokenID, shardID)
}

func TestChainSimulator_MetaESDTCreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := getTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, false)

	log.Info("Initial setup: Create MetaESDT that will have it's metadata saved to the user account")

	metaTicker := []byte("METATICKER")
	tx := issueMetaESDTTx(0, addrs[0].Bytes, metaTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	metaTokenID := txResult.Logs.Events[0].Topics[0]

	log.Info("Issued MetaESDT token id", "tokenID", string(metaTokenID))

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	createTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, metaTokenID, metaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
	checkMetaData(t, cs, core.SystemAccountAddress, metaTokenID, shardID, metaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaTokenID, shardID)
	checkMetaDataNotInAcc(t, cs, addrs[1].Bytes, metaTokenID, shardID)
}

func getTestChainSimulatorWithDynamicNFTEnabled(t *testing.T, baseIssuingCost string) (testsChainSimulator.ChainSimulator, int32) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpochForDynamicNFT := uint32(2)

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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpochForDynamicNFT
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpochForDynamicNFT))
	require.Nil(t, err)

	return cs, int32(activationEpochForDynamicNFT)
}

func getTestChainSimulatorWithSaveToSystemAccountDisabled(t *testing.T, baseIssuingCost string) (testsChainSimulator.ChainSimulator, int32) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpochForSaveToSystemAccount := uint32(4)
	activationEpochForDynamicNFT := uint32(6)

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
			cfg.EpochConfig.EnableEpochs.OptimizeNFTStoreEnableEpoch = activationEpochForSaveToSystemAccount
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpochForDynamicNFT
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpochForSaveToSystemAccount) - 2)
	require.Nil(t, err)

	return cs, int32(activationEpochForDynamicNFT)
}

func createTokenUpdateTokenIDAndTransfer(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	originAddress []byte,
	targetAddress []byte,
	tokenID []byte,
	metaData *txsFee.MetaData,
	epochForDynamicNFT int32,
	walletWithRoles dtos.WalletAddress,
) {
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	setAddressEsdtRoles(t, cs, 1, walletWithRoles, tokenID, roles)

	tx := nftCreateTx(2, originAddress, tokenID, metaData)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	log.Info("check that the metadata is saved on the user account")
	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(originAddress)
	checkMetaData(t, cs, originAddress, tokenID, shardID, metaData)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	err = cs.GenerateBlocksUntilEpochIsReached(epochForDynamicNFT)
	require.Nil(t, err)

	tx = updateTokenIDTx(3, originAddress, tokenID)

	log.Info("updating token id", "tokenID", tokenID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("transferring token id", "tokenID", tokenID)

	tx = esdtNFTTransferTx(4, originAddress, targetAddress, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
}

func TestChainSimulator_ChangeToDynamic_OldTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, epochForDynamicNFT := getTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, false)

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

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	for i := range tokenIDs {
		tx = nftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	// meta data should be saved on account, since it is before `OptimizeNFTStoreEnableEpoch`
	checkMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	checkMetaData(t, cs, addrs[0].Bytes, sftTokenID, shardID, sftMetaData)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, sftTokenID, shardID)

	checkMetaData(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID, esdtMetaData)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(epochForDynamicNFT))
	require.Nil(t, err)

	log.Info("Change to DYNAMIC type")

	// it will not be able to change nft to dynamic type
	for i := range tokenIDs {
		tx = changeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	for _, tokenID := range tokenIDs {
		tx = updateTokenIDTx(nonce, addrs[0].Bytes, tokenID)

		log.Info("updating token id", "tokenID", tokenID)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	for _, tokenID := range tokenIDs {
		log.Info("transfering token id", "tokenID", tokenID)

		tx = esdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	checkMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)
	checkMetaDataNotInAcc(t, cs, addrs[1].Bytes, sftTokenID, shardID)

	checkMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID)
	checkMetaDataNotInAcc(t, cs, addrs[1].Bytes, metaESDTTokenID, shardID)

	checkMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_CreateAndPause_NFT(t *testing.T) {
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := createAddresses(t, cs, false)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issueNonFungible"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString(nftTicker)),
			[]byte(hex.EncodeToString([]byte("canPause"))),
			[]byte(hex.EncodeToString([]byte("true"))),
		},
		[]byte("@"),
	)

	nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	nftTokenID := txResult.Logs.Events[0].Topics[0]
	setAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = nftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("check that the metadata for all tokens is saved on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	log.Info("Pause all tokens")

	scQuery := &process.SCQuery{
		ScAddress:  vm.ESDTSCAddress,
		CallerAddr: addrs[0].Bytes,
		FuncName:   "pause",
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{nftTokenID},
	}
	result, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	log.Info("wait for DynamicEsdtFlag activation")

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	log.Info("make an updateTokenID@tokenID function call on the ESDTSystem SC for all token types")

	tx = updateTokenIDTx(nonce, addrs[0].Bytes, nftTokenID)
	nonce++

	log.Info("updating token id", "tokenID", nftTokenID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("check that the metadata for all tokens is saved on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	log.Info("transfer the tokens to another account")

	log.Info("transfering token id", "tokenID", nftTokenID)

	tx = esdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("check that the metaData for the NFT is still on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)

	checkMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	checkMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_CreateAndPauseTokens_DynamicNFT(t *testing.T) {
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := createAddresses(t, cs, false)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	log.Info("Step 2. wait for DynamicEsdtFlag activation")

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	// register dynamic NFT
	nftTicker := []byte("NFTTICKER")
	nftTokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(nftTokenName)),
			[]byte(hex.EncodeToString(nftTicker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
			[]byte(hex.EncodeToString([]byte("canPause"))),
			[]byte(hex.EncodeToString([]byte("true"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	setAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = nftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Step 1. check that the metadata for all tokens is saved on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	log.Info("Step 1b. Pause all tokens")

	scQuery := &process.SCQuery{
		ScAddress:  vm.ESDTSCAddress,
		CallerAddr: addrs[0].Bytes,
		FuncName:   "pause",
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{nftTokenID},
	}
	result, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "", result.ReturnMessage)
	require.Equal(t, testsChainSimulator.OkReturnCode, result.ReturnCode)

	tx = updateTokenIDTx(nonce, addrs[0].Bytes, nftTokenID)
	nonce++

	log.Info("updating token id", "tokenID", nftTokenID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("change to dynamic token")

	tx = changeToDynamicTx(nonce, addrs[0].Bytes, nftTokenID)
	nonce++

	log.Info("updating token id", "tokenID", nftTokenID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("check that the metadata for all tokens is saved on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	log.Info("transfering token id", "tokenID", nftTokenID)

	tx = esdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	log.Info("check that the metaData for the NFT is still on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)

	checkMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
	checkMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	checkMetaDataNotInAcc(t, cs, addrs[1].Bytes, nftTokenID, shardID)
}

func TestChainSimulator_CheckRolesWhichHasToBeSingular(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := getTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := createAddresses(t, cs, true)

	// register dynamic NFT
	nftTicker := []byte("NFTTICKER")
	nftTokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(nftTokenName)),
			[]byte(hex.EncodeToString(nftTicker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
			[]byte(hex.EncodeToString([]byte("canPause"))),
			[]byte(hex.EncodeToString([]byte("true"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]

	log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTUpdateAttributes),
		[]byte(core.ESDTRoleNFTAddURI),
		[]byte(core.ESDTRoleSetNewURI),
		[]byte(core.ESDTRoleModifyCreator),
		[]byte(core.ESDTRoleModifyRoyalties),
		[]byte(core.ESDTRoleNFTRecreate),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	setAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	for _, role := range roles {
		tx = setSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID, [][]byte{role})
		nonce++

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		if txResult.Logs != nil && len(txResult.Logs.Events) > 0 {
			returnMessage := string(txResult.Logs.Events[0].Topics[1])
			require.True(t, strings.Contains(returnMessage, "already exists"))
		} else {
			require.Fail(t, "should have been return error message")
		}
	}
}
