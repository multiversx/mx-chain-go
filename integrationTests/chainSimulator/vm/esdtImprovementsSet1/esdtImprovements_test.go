package esdtImprovements

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
	vm2 "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/vm"

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
		vm2.TransferAndCheckTokensMetaData(t, false, false)
	})

	t.Run("transfer and check all tokens - intra shard - multi transfer", func(t *testing.T) {
		vm2.TransferAndCheckTokensMetaData(t, false, true)
	})

	t.Run("transfer and check all tokens - cross shard", func(t *testing.T) {
		vm2.TransferAndCheckTokensMetaData(t, true, false)
	})

	t.Run("transfer and check all tokens - cross shard - multi transfer", func(t *testing.T) {
		vm2.TransferAndCheckTokensMetaData(t, true, true)
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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, false)

	vm2.Log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := vm2.IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], metaESDTTokenID, roles)
	nonce++

	vm2.Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = vm2.IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = vm2.IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

	vm2.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("Step 1. check that the metaData for the NFT was saved in the user account and not on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	vm2.Log.Info("Step 2. check that the metaData for the other token types is saved on the system account and not at the user account level")

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID)
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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	vm2.Log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	addrs := vm2.CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := vm2.IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = vm2.IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = vm2.IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = vm2.ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleNFTRecreate),
		}
		tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("Call ESDTMetaDataRecreate to rewrite the meta data for the nft")

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
			GasPrice:  vm2.MinGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			vm2.CheckMetaData(t, cs, addrs[0].Bytes, tokenIDs[i], shardID, newMetaData)
		} else {
			vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID, newMetaData)
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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	vm2.Log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	addrs := vm2.CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := vm2.IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = vm2.IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = vm2.IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = vm2.ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleNFTUpdate),
		}
		tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	vm2.Log.Info("Call ESDTMetaDataUpdate to rewrite the meta data for the nft")

	for i := range tokenIDs {
		newMetaData := txsFee.GetDefaultMetaData()
		newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
		newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
		newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
		newMetaData.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))

		tx = vm2.EsdtMetaDataUpdateTx(tokenIDs[i], newMetaData, nonce, addrs[0].Bytes)
		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			vm2.CheckMetaData(t, cs, addrs[0].Bytes, tokenIDs[i], shardID, newMetaData)
		} else {
			vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID, newMetaData)
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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	vm2.Log.Info("Initial setup: Create NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag). Register NFT directly as dynamic")

	addrs := vm2.CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := vm2.IssueMetaESDTTx(nonce, addrs[1].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = vm2.SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = vm2.IssueSemiFungibleTx(nonce, addrs[1].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = vm2.EsdtNftCreateTx(nonce, addrs[1].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	vm2.Log.Info("Change to DYNAMIC type")

	for i := range tokenIDs {
		tx = vm2.ChangeToDynamicTx(nonce, addrs[1].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	vm2.Log.Info("Call ESDTModifyCreator and check that the creator was modified")

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(vm2.OneEGLD, mintValue)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	for i := range tokenIDs {
		vm2.Log.Info("Modify creator for token", "tokenID", tokenIDs[i])

		newCreatorAddress, err := cs.GenerateAndMintWalletAddress(shardID, mintValue)
		require.Nil(t, err)

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		roles = [][]byte{
			[]byte(core.ESDTRoleModifyCreator),
		}
		tx = vm2.SetSpecialRoleTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i], roles)
		nonce++

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		tx = vm2.ModifyCreatorTx(0, newCreatorAddress.Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		retrievedMetaData := &esdt.MetaData{}
		if bytes.Equal(tokenIDs[i], nftTokenID) {
			retrievedMetaData = vm2.GetMetaDataFromAcc(t, cs, newCreatorAddress.Bytes, tokenIDs[i], shardID)
		} else {
			retrievedMetaData = vm2.GetMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)
		}

		require.Equal(t, newCreatorAddress.Bytes, retrievedMetaData.Creator)
	}
}

func TestChainSimulator_ESDTModifyCreator_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := vm2.IssueMetaESDTTx(nonce, addrs[1].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = vm2.SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = vm2.IssueSemiFungibleTx(nonce, addrs[1].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[1].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = vm2.EsdtNftCreateTx(nonce, addrs[1].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	vm2.Log.Info("Change to DYNAMIC type")

	for i := range tokenIDs {
		tx = vm2.ChangeToDynamicTx(nonce, addrs[1].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	vm2.Log.Info("Call ESDTModifyCreator and check that the creator was modified")

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(vm2.OneEGLD, mintValue)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	crossShardID := uint32(2)
	if shardID == uint32(2) {
		crossShardID = uint32(1)
	}

	for i := range tokenIDs {
		vm2.Log.Info("Modify creator for token", "tokenID", string(tokenIDs[i]))

		newCreatorAddress, err := cs.GenerateAndMintWalletAddress(crossShardID, mintValue)
		require.Nil(t, err)

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		roles = [][]byte{
			[]byte(core.ESDTRoleModifyCreator),
		}
		tx = vm2.SetSpecialRoleTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i], roles)
		nonce++

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		vm2.Log.Info("transfering token id", "tokenID", tokenIDs[i])

		tx = vm2.EsdtNFTTransferTx(nonce, addrs[1].Bytes, newCreatorAddress.Bytes, tokenIDs[i])
		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		err = cs.GenerateBlocks(10)
		require.Nil(t, err)

		tx = vm2.ModifyCreatorTx(0, newCreatorAddress.Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		retrievedMetaData := &esdt.MetaData{}
		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(newCreatorAddress.Bytes)
		if bytes.Equal(tokenIDs[i], nftTokenID) {
			retrievedMetaData = vm2.GetMetaDataFromAcc(t, cs, newCreatorAddress.Bytes, tokenIDs[i], shardID)
		} else {
			retrievedMetaData = vm2.GetMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)
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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := vm2.IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = vm2.IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = vm2.IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = vm2.ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleSetNewURI),
		}
		tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	vm2.Log.Info("Call ESDTSetNewURIs and check that the new URIs were set for the tokens")

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
		vm2.Log.Info("Set new uris for token", "tokenID", string(tokenIDs[i]))

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
			GasPrice:  vm2.MinGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
		var retrievedMetaData *esdt.MetaData
		if bytes.Equal(tokenIDs[i], tokenIDs[0]) { // nft token
			retrievedMetaData = vm2.GetMetaDataFromAcc(t, cs, addrs[0].Bytes, tokenIDs[i], shardID)
		} else {
			retrievedMetaData = vm2.GetMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenIDs[i], shardID)
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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, false)

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := vm2.IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, metaESDTTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = vm2.IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, nftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = vm2.IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, sftTokenID, roles)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

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
		tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++

		tx = vm2.ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++

		roles := [][]byte{
			[]byte(core.ESDTRoleModifyRoyalties),
		}
		tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[0].Bytes, tokenIDs[i], roles)

		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	vm2.Log.Info("Call ESDTModifyRoyalties and check that the royalties were changed")

	metaDataNonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	royalties := []byte(hex.EncodeToString(big.NewInt(20).Bytes()))

	for i := range tokenIDs {
		vm2.Log.Info("Set new royalties for token", "tokenID", string(tokenIDs[i]))

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
			GasPrice:  vm2.MinGasPrice,
			Signature: []byte("dummySig"),
			Data:      txDataField,
			Value:     big.NewInt(0),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(addrs[0].Bytes)
		retrievedMetaData := vm2.GetMetaDataFromAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)

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
		PathToInitialConfig:            vm2.DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          vm2.RoundDurationInMillis,
		SupernovaRoundDurationInMillis: vm2.SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 vm2.RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        vm2.SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListMeta:        0,
		NumNodesWaitingListShard:       0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
			integrationTests.DeactivateSupernovaInConfig(cfg)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 2)
	require.Nil(t, err)

	vm2.Log.Info("Initial setup: Create NFT")

	nftTicker := []byte("NFTTICKER")
	nonce := uint64(0)
	tx := vm2.IssueNonFungibleTx(nonce, addrs[1].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[1], nftTokenID, roles)
	nonce++

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(nonce, addrs[1].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	vm2.Log.Info("Step 1. Change the nft to DYNAMIC type - the metadata should be on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	tx = vm2.ChangeToDynamicTx(nonce, addrs[1].Bytes, nftTokenID)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	roles = [][]byte{
		[]byte(core.ESDTRoleNFTUpdate),
	}

	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[1], nftTokenID, roles)
	nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	vm2.Log.Info("Step 2. Send the NFT cross shard")

	tx = vm2.EsdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("Step 3. The meta data should still be present on the system account")

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
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
		vm2.TestChainSimulatorChangeMetaData(t, vm2.IssueSemiFungibleTx)
	})

	t.Run("metaESDT change metadata", func(t *testing.T) {
		vm2.TestChainSimulatorChangeMetaData(t, vm2.IssueMetaESDTTx)
	})
}

func TestChainSimulator_NFT_RegisterDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic nft token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	vm2.Log.Info("Check that token type is Dynamic")

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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic metaESDT token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	vm2.Log.Info("Check that token type is Dynamic")

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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic fungible token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
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

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic nft token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	vm2.Log.Info("Check that token type is Dynamic")

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

	vm2.Log.Info("Check token roles")

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

	vm2.CheckTokenRoles(t, result.ReturnData, expectedRoles)
}

func TestChainSimulator_SFT_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic sft token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, sftTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, nftMetaData)

	vm2.Log.Info("Check that token type is Dynamic")

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

	vm2.Log.Info("Check token roles")

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

	vm2.CheckTokenRoles(t, result.ReturnData, expectedRoles)
}

func TestChainSimulator_FNG_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic fungible token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	signalErrorTopic := string(txResult.Logs.Events[0].Topics[1])

	require.Equal(t, fmt.Sprintf("cannot create %s tokens as dynamic", core.FungibleESDT), signalErrorTopic)
}
