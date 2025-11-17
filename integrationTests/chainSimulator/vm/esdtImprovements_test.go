package vm

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
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
)

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
		TransferAndCheckTokensMetaData(t, false, false)
	})

	t.Run("transfer and check all tokens - intra shard - multi transfer", func(t *testing.T) {
		TransferAndCheckTokensMetaData(t, false, true)
	})

	t.Run("transfer and check all tokens - cross shard", func(t *testing.T) {
		TransferAndCheckTokensMetaData(t, true, false)
	})

	t.Run("transfer and check all tokens - cross shard - multi transfer", func(t *testing.T) {
		TransferAndCheckTokensMetaData(t, true, true)
	})
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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	Log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], metaESDTTokenID, roles)
	nonce++

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("Step 1. check that the metaData for the NFT was saved in the user account and not on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	Log.Info("Step 2. check that the metaData for the other token types is saved on the system account and not at the user account level")

	CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)

	CheckMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID)
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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	Log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	addrs := CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleNFTRecreate),
		}
		tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("Call ESDTMetaDataRecreate to rewrite the meta data for the nft")

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
			GasPrice:  MinGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			CheckMetaData(t, cs, addrs[0].Bytes, tokenIDs[i], shardID, newMetaData)
		} else {
			CheckMetaData(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID, newMetaData)
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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	Log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	addrs := CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleNFTUpdate),
		}
		tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	Log.Info("Call ESDTMetaDataUpdate to rewrite the meta data for the nft")

	for i := range tokenIDs {
		newMetaData := txsFee.GetDefaultMetaData()
		newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
		newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
		newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
		newMetaData.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))

		tx = EsdtMetaDataUpdateTx(tokenIDs[i], newMetaData, nonce, addrs[0].Bytes)
		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			CheckMetaData(t, cs, addrs[0].Bytes, tokenIDs[i], shardID, newMetaData)
		} else {
			CheckMetaData(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID, newMetaData)
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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	Log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag). Register NFT directly as dynamic")

	addrs := CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[1].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[1].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = EsdtNftCreateTx(nonce, addrs[1].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	Log.Info("Change to DYNAMIC type")

	for i := range tokenIDs {
		tx = ChangeToDynamicTx(nonce, addrs[1].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	Log.Info("Call ESDTModifyCreator and check that the creator was modified")

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(OneEGLD, mintValue)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	for i := range tokenIDs {
		Log.Info("Modify creator for token", "tokenID", tokenIDs[i])

		newCreatorAddress, err := cs.GenerateAndMintWalletAddress(shardID, mintValue)
		require.Nil(t, err)

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		roles = [][]byte{
			[]byte(core.ESDTRoleModifyCreator),
		}
		tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i], roles)
		nonce++

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		tx = ModifyCreatorTx(0, newCreatorAddress.Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		retrievedMetaData := &esdt.MetaData{}
		if bytes.Equal(tokenIDs[i], nftTokenID) {
			retrievedMetaData = GetMetaDataFromAcc(t, cs, newCreatorAddress.Bytes, tokenIDs[i], shardID)
		} else {
			retrievedMetaData = GetMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)
		}

		require.Equal(t, newCreatorAddress.Bytes, retrievedMetaData.Creator)
	}
}

func TestChainSimulator_ESDTModifyCreator_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[1].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[1].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = EsdtNftCreateTx(nonce, addrs[1].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	Log.Info("Change to DYNAMIC type")

	for i := range tokenIDs {
		tx = ChangeToDynamicTx(nonce, addrs[1].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	Log.Info("Call ESDTModifyCreator and check that the creator was modified")

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(OneEGLD, mintValue)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	crossShardID := uint32(2)
	if shardID == uint32(2) {
		crossShardID = uint32(1)
	}

	for i := range tokenIDs {
		Log.Info("Modify creator for token", "tokenID", string(tokenIDs[i]))

		newCreatorAddress, err := cs.GenerateAndMintWalletAddress(crossShardID, mintValue)
		require.Nil(t, err)

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		roles = [][]byte{
			[]byte(core.ESDTRoleModifyCreator),
		}
		tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i], roles)
		nonce++

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		Log.Info("transfering token id", "tokenID", tokenIDs[i])

		tx = EsdtNFTTransferTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i])
		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		tx = ModifyCreatorTx(0, newCreatorAddress.Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		retrievedMetaData := &esdt.MetaData{}
		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(newCreatorAddress.Bytes)
		if bytes.Equal(tokenIDs[i], nftTokenID) {
			retrievedMetaData = GetMetaDataFromAcc(t, cs, newCreatorAddress.Bytes, tokenIDs[i], shardID)
		} else {
			retrievedMetaData = GetMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)
		}

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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleSetNewURI),
		}
		tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	Log.Info("Call ESDTSetNewURIs and check that the new URIs were set for the tokens")

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
		Log.Info("Set new uris for token", "tokenID", string(tokenIDs[i]))

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
			GasPrice:  MinGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
		var retrievedMetaData *esdt.MetaData
		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			retrievedMetaData = GetMetaDataFromAcc(t, cs, addrs[0].Bytes, tokenIDs[i], shardID)
		} else {
			retrievedMetaData = GetMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)
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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleModifyRoyalties),
		}
		tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	Log.Info("Call ESDTModifyRoyalties and check that the royalties were changed")

	metaDataNonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	royalties := []byte(hex.EncodeToString(big.NewInt(20).Bytes()))

	for i := range tokenIDs {
		Log.Info("Set new royalties for token", "tokenID", string(tokenIDs[i]))

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
			GasPrice:  MinGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(addrs[0].Bytes)
		retrievedMetaData := GetMetaDataFromAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)

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

	activationEpoch := uint32(4)

	baseIssuingCost := "1000"

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          RoundDurationInMillis,
		SupernovaRoundDurationInMillis: SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListMeta:        0,
		NumNodesWaitingListShard:       0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 2)
	require.Nil(t, err)

	Log.Info("Initial setup: Create NFT")

	nftTicker := []byte("NFTTICKER")
	nonce := uint64(0)
	tx := IssueNonFungibleTx(nonce, addrs[1].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[1], nftTokenID, roles)
	nonce++

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[1].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	Log.Info("Step 1. Change the nft to DYNAMIC type - the metadata should be on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	tx = ChangeToDynamicTx(nonce, addrs[1].Bytes, nftTokenID)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	roles = [][]byte{
		[]byte(core.ESDTRoleNFTUpdate),
	}

	SetAddressEsdtRoles(t, cs, nonce, addrs[1], nftTokenID, roles)
	nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	Log.Info("Step 2. Send the NFT cross shard")

	tx = EsdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Step 3. The meta data should still be present on the system account")

	CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
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
		TestChainSimulatorChangeMetaData(t, IssueSemiFungibleTx)
	})

	t.Run("metaESDT change metadata", func(t *testing.T) {
		TestChainSimulatorChangeMetaData(t, IssueMetaESDTTx)
	})
}

func TestChainSimulator_NFT_RegisterDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic nft token")

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	Log.Info("Check that token type is Dynamic")

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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic metaESDT token")

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	Log.Info("Check that token type is Dynamic")

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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic fungible token")

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic nft token")

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	Log.Info("Check that token type is Dynamic")

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

	Log.Info("Check token roles")

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

	CheckTokenRoles(t, result.ReturnData, expectedRoles)
}

func TestChainSimulator_SFT_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic sft token")

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, sftTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, nftMetaData)

	Log.Info("Check that token type is Dynamic")

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

	Log.Info("Check token roles")

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

	CheckTokenRoles(t, result.ReturnData, expectedRoles)
}

func TestChainSimulator_FNG_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic fungible token")

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
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

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic meta esdt token")

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	metaTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], metaTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, metaTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, core.SystemAccountAddress, metaTokenID, shardID, nftMetaData)

	Log.Info("Check that token type is Dynamic")

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

	Log.Info("Check token roles")

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

	CheckTokenRoles(t, result.ReturnData, expectedRoles)
}

func TestChainSimulator_NFTcreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := GetTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	Log.Info("Initial setup: Create NFT that will have it's metadata saved to the user account")

	nftTicker := []byte("NFTTICKER")
	tx := IssueNonFungibleTx(0, addrs[0].Bytes, nftTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	nftTokenID := txResult.Logs.Events[0].Topics[0]

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	CreateTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, nftTokenID, nftMetaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
	CheckMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_SFTcreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := GetTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	Log.Info("Initial setup: Create SFT that will have it's metadata saved to the user account")

	sftTicker := []byte("SFTTICKER")
	tx := IssueSemiFungibleTx(0, addrs[0].Bytes, sftTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	sftTokenID := txResult.Logs.Events[0].Topics[0]

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	CreateTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, sftTokenID, metaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, metaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)
	CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, sftTokenID, shardID)
}

func TestChainSimulator_MetaESDTCreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := GetTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	Log.Info("Initial setup: Create MetaESDT that will have it's metadata saved to the user account")

	metaTicker := []byte("METATICKER")
	tx := IssueMetaESDTTx(0, addrs[0].Bytes, metaTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	metaTokenID := txResult.Logs.Events[0].Topics[0]

	Log.Info("Issued MetaESDT token id", "tokenID", string(metaTokenID))

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	CreateTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, metaTokenID, metaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
	CheckMetaData(t, cs, core.SystemAccountAddress, metaTokenID, shardID, metaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaTokenID, shardID)
	CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, metaTokenID, shardID)
}

func TestChainSimulator_ChangeToDynamic_OldTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, epochForDynamicNFT := GetTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], metaESDTTokenID, roles)
	nonce++

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	// meta data should be saved on account, since it is before `OptimizeNFTStoreEnableEpoch`
	CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	CheckMetaData(t, cs, addrs[0].Bytes, sftTokenID, shardID, sftMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, sftTokenID, shardID)

	CheckMetaData(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID, esdtMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID)

	err = cs.GenerateBlocksUntilEpochIsReached(epochForDynamicNFT)
	require.Nil(t, err)

	Log.Info("Change to DYNAMIC type")

	// it will not be able to change nft to dynamic type
	for i := range tokenIDs {
		tx = ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	for _, tokenID := range tokenIDs {
		tx = UpdateTokenIDTx(nonce, addrs[0].Bytes, tokenID)

		Log.Info("updating token id", "tokenID", tokenID)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	for _, tokenID := range tokenIDs {
		Log.Info("transfering token id", "tokenID", tokenID)

		tx = EsdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)
	CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, sftTokenID, shardID)

	CheckMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID)
	CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, metaESDTTokenID, shardID)

	CheckMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_CreateAndPause_NFT(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	activationEpoch := uint32(4)

	baseIssuingCost := "1000"

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          RoundDurationInMillis,
		SupernovaRoundDurationInMillis: SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListMeta:        0,
		NumNodesWaitingListShard:       0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	nftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("check that the metadata for all tokens is saved on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	Log.Info("Pause all tokens")

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

	Log.Info("wait for DynamicEsdtFlag activation")

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	Log.Info("make an updateTokenID@tokenID function call on the ESDTSystem SC for all token types")

	tx = UpdateTokenIDTx(nonce, addrs[0].Bytes, nftTokenID)
	nonce++

	Log.Info("updating token id", "tokenID", nftTokenID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("check that the metadata for all tokens is saved on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	Log.Info("transfer the tokens to another account")

	Log.Info("transfering token id", "tokenID", nftTokenID)

	tx = EsdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("check that the metaData for the NFT is still on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)

	CheckMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_CreateAndPauseTokens_DynamicNFT(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	activationEpoch := uint32(4)

	baseIssuingCost := "1000"

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          RoundDurationInMillis,
		SupernovaRoundDurationInMillis: SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListMeta:        0,
		NumNodesWaitingListShard:       0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := CreateAddresses(t, cs, false)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	Log.Info("Step 2. wait for DynamicEsdtFlag activation")

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("Step 1. check that the metadata for the Dynamic NFT is saved on the user account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	Log.Info("Step 1b. Pause all tokens")

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

	Log.Info("check that the metadata for the Dynamic NFT is saved on the user account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	Log.Info("transfering token id", "tokenID", nftTokenID)

	tx = EsdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("check that the metaData for the NFT is on the new user account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)

	CheckMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_CheckRolesWhichHasToBeSingular(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

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
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

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
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	for _, role := range roles {
		tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID, [][]byte{role})
		nonce++

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
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

func TestChainSimulator_metaESDT_mergeMetaDataFromMultipleUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()
	marshaller := cs.GetNodeHandler(0).GetCoreComponents().InternalMarshalizer()
	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic metaESDT token")

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

	shard0Nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     shard0Nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTAddQuantity),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 2)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, shardID, metaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, shardID, []byte{1})

	Log.Info("send metaEsdt cross shard")

	tx = EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("update metaData on shard 0")

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))

	tx = EsdtMetaDataUpdateTx(tokenID, newMetaData, shard0Nonce, addrs[0].Bytes)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	expectedMetaData := txsFee.GetDefaultMetaData()
	expectedMetaData.Nonce = newMetaData.Nonce
	expectedMetaData.Name = newMetaData.Name
	expectedMetaData.Hash = newMetaData.Hash
	expectedMetaData.Attributes = newMetaData.Attributes

	round := cs.GetNodeHandler(0).GetChainHandler().GetCurrentBlockHeader().GetRound()
	reserved := &esdt.MetaDataVersion{
		Name:       round,
		Creator:    round,
		Hash:       round,
		Attributes: round,
	}
	firstVersion, _ := marshaller.Marshal(reserved)

	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, expectedMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, firstVersion)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, metaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, []byte{1})

	Log.Info("send the update role to shard 2")

	shard0Nonce = TransferSpecialRoleToAddr(t, cs, shard0Nonce, tokenID, addrs[0].Bytes, addrs[2].Bytes, []byte(core.ESDTRoleNFTUpdate))

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("update metaData on shard 2")

	newMetaData2 := &txsFee.MetaData{}
	newMetaData2.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData2.Uris = [][]byte{[]byte(hex.EncodeToString([]byte("uri5"))), []byte(hex.EncodeToString([]byte("uri6"))), []byte(hex.EncodeToString([]byte("uri7")))}
	newMetaData2.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = EsdtMetaDataUpdateTx(tokenID, newMetaData2, 0, addrs[2].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, expectedMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, firstVersion)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, metaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, []byte{1})

	retrievedMetaData := GetMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, 2)
	require.Equal(t, uint64(1), retrievedMetaData.Nonce)
	require.Equal(t, 0, len(retrievedMetaData.Name))
	require.Equal(t, addrs[2].Bytes, retrievedMetaData.Creator)
	require.Equal(t, newMetaData2.Royalties, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Royalties)).Bytes())))
	require.Equal(t, 0, len(retrievedMetaData.Hash))
	require.Equal(t, 3, len(retrievedMetaData.URIs))
	for i, uri := range newMetaData2.Uris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, 0, len(retrievedMetaData.Attributes))

	round2 := cs.GetNodeHandler(2).GetChainHandler().GetCurrentBlockHeader().GetRound()
	reserved = &esdt.MetaDataVersion{
		URIs:      round2,
		Creator:   round2,
		Royalties: round2,
	}
	secondVersion, _ := cs.GetNodeHandler(shardID).GetCoreComponents().InternalMarshalizer().Marshal(reserved)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 2, secondVersion)

	Log.Info("transfer from shard 0 to shard 1 - should merge metaData")

	tx = EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, expectedMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, firstVersion)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, expectedMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, firstVersion)

	Log.Info("transfer from shard 1 to shard 2 - should merge metaData")

	tx = SetSpecialRoleTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID, [][]byte{[]byte(core.ESDTRoleTransfer)})
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	shard0Nonce++

	tx = EsdtNFTTransferTx(0, addrs[1].Bytes, addrs[2].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, expectedMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, firstVersion)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, expectedMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, firstVersion)

	latestMetaData := txsFee.GetDefaultMetaData()
	latestMetaData.Nonce = expectedMetaData.Nonce
	latestMetaData.Name = expectedMetaData.Name
	latestMetaData.Royalties = newMetaData2.Royalties
	latestMetaData.Hash = expectedMetaData.Hash
	latestMetaData.Attributes = expectedMetaData.Attributes
	latestMetaData.Uris = newMetaData2.Uris
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 2, latestMetaData)

	reserved = &esdt.MetaDataVersion{
		Name:       round,
		Creator:    round2,
		Royalties:  round2,
		Hash:       round,
		URIs:       round2,
		Attributes: round,
	}
	thirdVersion, _ := cs.GetNodeHandler(shardID).GetCoreComponents().InternalMarshalizer().Marshal(reserved)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 2, thirdVersion)

	Log.Info("transfer from shard 2 to shard 0 - should update metaData")

	tx = SetSpecialRoleTx(shard0Nonce, addrs[0].Bytes, addrs[2].Bytes, tokenID, [][]byte{[]byte(core.ESDTRoleTransfer)})
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	tx = EsdtNFTTransferTx(1, addrs[2].Bytes, addrs[0].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, latestMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, thirdVersion)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, expectedMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, firstVersion)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 2, latestMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 2, thirdVersion)

	Log.Info("transfer from shard 1 to shard 0 - liquidity should be updated")

	tx = EsdtNFTTransferTx(1, addrs[1].Bytes, addrs[0].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, latestMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, thirdVersion)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, expectedMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, firstVersion)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 2, latestMetaData)
	CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 2, thirdVersion)
}

func TestChainSimulator_dynamicNFT_mergeMetaDataFromMultipleUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic NFT token")

	ticker := []byte("NFTTICKER")
	tokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(tokenName)),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	shard0Nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     shard0Nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, metaData)

	Log.Info("give update role to another account and update metaData")

	shard0Nonce = TransferSpecialRoleToAddr(t, cs, shard0Nonce, tokenID, addrs[0].Bytes, addrs[1].Bytes, []byte(core.ESDTRoleNFTUpdate))

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = EsdtMetaDataUpdateTx(tokenID, newMetaData, 0, addrs[1].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, metaData)
	newMetaData.Attributes = []byte{}
	CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, newMetaData)

	Log.Info("transfer nft - should merge metaData")

	tx = EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	mergedMetaData := txsFee.GetDefaultMetaData()
	mergedMetaData.Nonce = metaData.Nonce
	mergedMetaData.Name = newMetaData.Name
	mergedMetaData.Hash = newMetaData.Hash
	mergedMetaData.Royalties = newMetaData.Royalties

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, tokenID, 0)
	CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, mergedMetaData)

	Log.Info("transfer nft - should remove metaData from sender")

	tx = SetSpecialRoleTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID, [][]byte{[]byte(core.ESDTRoleTransfer)})
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	tx = EsdtNFTTransferTx(1, addrs[1].Bytes, addrs[2].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 2)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, tokenID, 0)
	CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, tokenID, 1)
	CheckMetaData(t, cs, addrs[2].Bytes, tokenID, 2, mergedMetaData)
}

func TestChainSimulator_dynamicNFT_changeMetaDataForOneNFTShouldNotChangeOtherNonces(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic NFT token")

	ticker := []byte("NFTTICKER")
	tokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(tokenName)),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	shard0Nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     shard0Nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(2).Bytes()))
	tx = EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("give update role to another account and update metaData for nonce 2")

	shard0Nonce = TransferSpecialRoleToAddr(t, cs, shard0Nonce, tokenID, addrs[0].Bytes, addrs[1].Bytes, []byte(core.ESDTRoleNFTUpdate))

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(2).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = EsdtMetaDataUpdateTx(tokenID, newMetaData, 0, addrs[1].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("transfer nft with nonce 1 - should not merge metaData")

	tx = EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, metaData)
}

func TestChainSimulator_dynamicNFT_updateBeforeCreateOnSameAccountShouldOverwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic NFT token")

	ticker := []byte("NFTTICKER")
	tokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(tokenName)),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	shard0Nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     shard0Nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("update meta data for a token that is not yet created")

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = EsdtMetaDataUpdateTx(tokenID, newMetaData, shard0Nonce, addrs[0].Bytes)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	newMetaData.Attributes = []byte{}
	newMetaData.Uris = [][]byte{}
	CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, newMetaData)

	Log.Info("create nft with the same nonce - should overwrite the metadata")

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, metaData)
}

func TestChainSimulator_dynamicNFT_updateBeforeCreateOnDifferentAccountsShouldMergeMetaDataWhenTransferred(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Register dynamic NFT token")

	ticker := []byte("NFTTICKER")
	tokenName := []byte("tokenName")

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerDynamic"),
			[]byte(hex.EncodeToString(tokenName)),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString([]byte("NFT"))),
		},
		[]byte("@"),
	)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	shard0Nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     shard0Nonce,
		SndAddr:   addrs[0].Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("transfer update role to another address")

	shard0Nonce = TransferSpecialRoleToAddr(t, cs, shard0Nonce, tokenID, addrs[0].Bytes, addrs[1].Bytes, []byte(core.ESDTRoleNFTUpdate))

	Log.Info("update meta data for a token that is not yet created")

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = EsdtMetaDataUpdateTx(tokenID, newMetaData, 0, addrs[1].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, tokenID, 0)
	newMetaData.Attributes = []byte{}
	newMetaData.Uris = [][]byte{}
	CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, newMetaData)

	Log.Info("create nft with the same nonce on different account")

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, metaData)
	CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, newMetaData)

	Log.Info("transfer dynamic NFT to the updated account")

	tx = EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, tokenID, 0)
	newMetaData.Attributes = metaData.Attributes
	newMetaData.Uris = metaData.Uris
	CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, newMetaData)
}
