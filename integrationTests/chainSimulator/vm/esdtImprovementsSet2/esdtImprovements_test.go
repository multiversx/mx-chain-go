package esdtImprovementsSet2

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	vm2 "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
)

func TestChainSimulator_MetaESDT_RegisterAndSetAllRolesDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic meta esdt token")

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

	metaTokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], metaTokenID, roles)
	nonce++

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, metaTokenID, nftMetaData, 1)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, metaTokenID, shardID, nftMetaData)

	vm2.Log.Info("Check that token type is Dynamic")

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

	vm2.Log.Info("Check token roles")

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

	vm2.CheckTokenRoles(t, result.ReturnData, expectedRoles)
}

func TestChainSimulator_NFTcreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := vm2.GetTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, false)

	vm2.Log.Info("Initial setup: Create NFT that will have it's metadata saved to the user account")

	nftTicker := []byte("NFTTICKER")
	tx := vm2.IssueNonFungibleTx(0, addrs[0].Bytes, nftTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	nftTokenID := txResult.Logs.Events[0].Topics[0]

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	vm2.CreateTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, nftTokenID, nftMetaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
	vm2.CheckMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_SFTcreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := vm2.GetTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, false)

	vm2.Log.Info("Initial setup: Create SFT that will have it's metadata saved to the user account")

	sftTicker := []byte("SFTTICKER")
	tx := vm2.IssueSemiFungibleTx(0, addrs[0].Bytes, sftTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	sftTokenID := txResult.Logs.Events[0].Topics[0]

	vm2.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	vm2.CreateTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, sftTokenID, metaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, metaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, sftTokenID, shardID)
}

func TestChainSimulator_MetaESDTCreatedBeforeSaveToSystemAccountEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, epochForDynamicNFT := vm2.GetTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, false)

	vm2.Log.Info("Initial setup: Create MetaESDT that will have it's metadata saved to the user account")

	metaTicker := []byte("METATICKER")
	tx := vm2.IssueMetaESDTTx(0, addrs[0].Bytes, metaTicker, baseIssuingCost)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	metaTokenID := txResult.Logs.Events[0].Topics[0]

	vm2.Log.Info("Issued MetaESDT token id", "tokenID", string(metaTokenID))

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	vm2.CreateTokenUpdateTokenIDAndTransfer(t, cs, addrs[0].Bytes, addrs[1].Bytes, metaTokenID, metaData, epochForDynamicNFT, addrs[0])

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, metaTokenID, shardID, metaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaTokenID, shardID)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, metaTokenID, shardID)
}

func TestChainSimulator_ChangeToDynamic_OldTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, epochForDynamicNFT := vm2.GetTestChainSimulatorWithSaveToSystemAccountDisabled(t, baseIssuingCost)
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

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	// meta data should be saved on account, since it is before `OptimizeNFTStoreEnableEpoch`
	vm2.CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	vm2.CheckMetaData(t, cs, addrs[0].Bytes, sftTokenID, shardID, sftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, sftTokenID, shardID)

	vm2.CheckMetaData(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID, esdtMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID)

	err = cs.GenerateBlocksUntilEpochIsReached(epochForDynamicNFT)
	require.Nil(t, err)

	vm2.Log.Info("Change to DYNAMIC type")

	// it will not be able to change nft to dynamic type
	for i := range tokenIDs {
		tx = vm2.ChangeToDynamicTx(nonce, addrs[0].Bytes, tokenIDs[i])

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	for _, tokenID := range tokenIDs {
		tx = vm2.UpdateTokenIDTx(nonce, addrs[0].Bytes, tokenID)

		vm2.Log.Info("updating token id", "tokenID", tokenID)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	for _, tokenID := range tokenIDs {
		vm2.Log.Info("transfering token id", "tokenID", tokenID)

		tx = vm2.EsdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, sftTokenID, shardID)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, sftTokenID, shardID)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, metaESDTTokenID, shardID)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, metaESDTTokenID, shardID)

	vm2.CheckMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
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

	addrs := vm2.CreateAddresses(t, cs, false)

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

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	nftTokenID := txResult.Logs.Events[0].Topics[0]
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("check that the metadata for all tokens is saved on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	vm2.Log.Info("Pause all tokens")

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

	vm2.Log.Info("wait for DynamicEsdtFlag activation")

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	vm2.Log.Info("make an updateTokenID@tokenID function call on the ESDTSystem SC for all token types")

	tx = vm2.UpdateTokenIDTx(nonce, addrs[0].Bytes, nftTokenID)
	nonce++

	vm2.Log.Info("updating token id", "tokenID", nftTokenID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("check that the metadata for all tokens is saved on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)

	vm2.Log.Info("transfer the tokens to another account")

	vm2.Log.Info("transfering token id", "tokenID", nftTokenID)

	tx = vm2.EsdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("check that the metaData for the NFT is still on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)

	vm2.CheckMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
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

	addrs := vm2.CreateAddresses(t, cs, false)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	vm2.Log.Info("Step 2. wait for DynamicEsdtFlag activation")

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

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("Step 1. check that the metadata for the Dynamic NFT is saved on the user account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	vm2.Log.Info("Step 1b. Pause all tokens")

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

	vm2.Log.Info("check that the metadata for the Dynamic NFT is saved on the user account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaData(t, cs, addrs[0].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	vm2.Log.Info("transfering token id", "tokenID", nftTokenID)

	tx = vm2.EsdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("check that the metaData for the NFT is on the new user account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)

	vm2.CheckMetaData(t, cs, addrs[1].Bytes, nftTokenID, shardID, nftMetaData)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, nftTokenID, shardID)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)
}

func TestChainSimulator_CheckRolesWhichHasToBeSingular(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"

	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

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

	vm2.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

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
	vm2.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	for _, role := range roles {
		tx = vm2.SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID, [][]byte{role})
		nonce++

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
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
	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()
	marshaller := cs.GetNodeHandler(0).GetCoreComponents().InternalMarshalizer()
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

	shard0Nonce := uint64(0)
	tx := &transaction.Transaction{
		Nonce:     shard0Nonce,
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
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
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
	vm2.SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 2)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, shardID, metaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, shardID, []byte{1})

	vm2.Log.Info("send metaEsdt cross shard")

	tx = vm2.EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("update metaData on shard 0")

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))

	tx = vm2.EsdtMetaDataUpdateTx(tokenID, newMetaData, shard0Nonce, addrs[0].Bytes)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
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

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, expectedMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, firstVersion)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, metaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, []byte{1})

	vm2.Log.Info("send the update role to shard 2")

	shard0Nonce = vm2.TransferSpecialRoleToAddr(t, cs, shard0Nonce, tokenID, addrs[0].Bytes, addrs[2].Bytes, []byte(core.ESDTRoleNFTUpdate))

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("update metaData on shard 2")

	newMetaData2 := &txsFee.MetaData{}
	newMetaData2.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData2.Uris = [][]byte{[]byte(hex.EncodeToString([]byte("uri5"))), []byte(hex.EncodeToString([]byte("uri6"))), []byte(hex.EncodeToString([]byte("uri7")))}
	newMetaData2.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = vm2.EsdtMetaDataUpdateTx(tokenID, newMetaData2, 0, addrs[2].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, expectedMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, firstVersion)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, metaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, []byte{1})

	retrievedMetaData := vm2.GetMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, 2)
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
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 2, secondVersion)

	vm2.Log.Info("transfer from shard 0 to shard 1 - should merge metaData")

	tx = vm2.EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, expectedMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, firstVersion)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, expectedMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, firstVersion)

	vm2.Log.Info("transfer from shard 1 to shard 2 - should merge metaData")

	tx = vm2.SetSpecialRoleTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID, [][]byte{[]byte(core.ESDTRoleTransfer)})
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	shard0Nonce++

	tx = vm2.EsdtNFTTransferTx(0, addrs[1].Bytes, addrs[2].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, expectedMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, firstVersion)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, expectedMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, firstVersion)

	latestMetaData := txsFee.GetDefaultMetaData()
	latestMetaData.Nonce = expectedMetaData.Nonce
	latestMetaData.Name = expectedMetaData.Name
	latestMetaData.Royalties = newMetaData2.Royalties
	latestMetaData.Hash = expectedMetaData.Hash
	latestMetaData.Attributes = expectedMetaData.Attributes
	latestMetaData.Uris = newMetaData2.Uris
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 2, latestMetaData)

	reserved = &esdt.MetaDataVersion{
		Name:       round,
		Creator:    round2,
		Royalties:  round2,
		Hash:       round,
		URIs:       round2,
		Attributes: round,
	}
	thirdVersion, _ := cs.GetNodeHandler(shardID).GetCoreComponents().InternalMarshalizer().Marshal(reserved)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 2, thirdVersion)

	vm2.Log.Info("transfer from shard 2 to shard 0 - should update metaData")

	tx = vm2.SetSpecialRoleTx(shard0Nonce, addrs[0].Bytes, addrs[2].Bytes, tokenID, [][]byte{[]byte(core.ESDTRoleTransfer)})
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	tx = vm2.EsdtNFTTransferTx(1, addrs[2].Bytes, addrs[0].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, latestMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, thirdVersion)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, expectedMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, firstVersion)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 2, latestMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 2, thirdVersion)

	vm2.Log.Info("transfer from shard 1 to shard 0 - liquidity should be updated")

	tx = vm2.EsdtNFTTransferTx(1, addrs[1].Bytes, addrs[0].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 0, latestMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 0, thirdVersion)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 1, expectedMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 1, firstVersion)
	vm2.CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, 2, latestMetaData)
	vm2.CheckReservedField(t, cs, core.SystemAccountAddress, tokenID, 2, thirdVersion)
}

func TestChainSimulator_dynamicNFT_mergeMetaDataFromMultipleUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic NFT token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	vm2.SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, metaData)

	vm2.Log.Info("give update role to another account and update metaData")

	shard0Nonce = vm2.TransferSpecialRoleToAddr(t, cs, shard0Nonce, tokenID, addrs[0].Bytes, addrs[1].Bytes, []byte(core.ESDTRoleNFTUpdate))

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = vm2.EsdtMetaDataUpdateTx(tokenID, newMetaData, 0, addrs[1].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	vm2.CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, metaData)
	newMetaData.Attributes = []byte{}
	vm2.CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, newMetaData)

	vm2.Log.Info("transfer nft - should merge metaData")

	tx = vm2.EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
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

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, tokenID, 0)
	vm2.CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, mergedMetaData)

	vm2.Log.Info("transfer nft - should remove metaData from sender")

	tx = vm2.SetSpecialRoleTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID, [][]byte{[]byte(core.ESDTRoleTransfer)})
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	tx = vm2.EsdtNFTTransferTx(1, addrs[1].Bytes, addrs[2].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 2)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, tokenID, 0)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[1].Bytes, tokenID, 1)
	vm2.CheckMetaData(t, cs, addrs[2].Bytes, tokenID, 2, mergedMetaData)
}

func TestChainSimulator_dynamicNFT_changeMetaDataForOneNFTShouldNotChangeOtherNonces(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic NFT token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	vm2.SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(2).Bytes()))
	tx = vm2.EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("give update role to another account and update metaData for nonce 2")

	shard0Nonce = vm2.TransferSpecialRoleToAddr(t, cs, shard0Nonce, tokenID, addrs[0].Bytes, addrs[1].Bytes, []byte(core.ESDTRoleNFTUpdate))

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(2).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = vm2.EsdtMetaDataUpdateTx(tokenID, newMetaData, 0, addrs[1].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.Log.Info("transfer nft with nonce 1 - should not merge metaData")

	tx = vm2.EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	vm2.CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, metaData)
}

func TestChainSimulator_dynamicNFT_updateBeforeCreateOnSameAccountShouldOverwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic NFT token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	vm2.SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("update meta data for a token that is not yet created")

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = vm2.EsdtMetaDataUpdateTx(tokenID, newMetaData, shard0Nonce, addrs[0].Bytes)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	newMetaData.Attributes = []byte{}
	newMetaData.Uris = [][]byte{}
	vm2.CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, newMetaData)

	vm2.Log.Info("create nft with the same nonce - should overwrite the metadata")

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, metaData)
}

func TestChainSimulator_dynamicNFT_updateBeforeCreateOnDifferentAccountsShouldMergeMetaDataWhenTransferred(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	baseIssuingCost := "1000"
	cs, _ := vm2.GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := vm2.CreateAddresses(t, cs, true)

	vm2.Log.Info("Register dynamic NFT token")

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
		GasPrice:  vm2.MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	shard0Nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	vm2.SetAddressEsdtRoles(t, cs, shard0Nonce, addrs[0], tokenID, roles)
	shard0Nonce++

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.Log.Info("transfer update role to another address")

	shard0Nonce = vm2.TransferSpecialRoleToAddr(t, cs, shard0Nonce, tokenID, addrs[0].Bytes, addrs[1].Bytes, []byte(core.ESDTRoleNFTUpdate))

	vm2.Log.Info("update meta data for a token that is not yet created")

	newMetaData := &txsFee.MetaData{}
	newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
	newMetaData.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	newMetaData.Royalties = []byte(hex.EncodeToString(big.NewInt(15).Bytes()))

	tx = vm2.EsdtMetaDataUpdateTx(tokenID, newMetaData, 0, addrs[1].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, tokenID, 0)
	newMetaData.Attributes = []byte{}
	newMetaData.Uris = [][]byte{}
	vm2.CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, newMetaData)

	vm2.Log.Info("create nft with the same nonce on different account")

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm2.EsdtNftCreateTx(shard0Nonce, addrs[0].Bytes, tokenID, metaData, 1)
	shard0Nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	vm2.CheckMetaData(t, cs, addrs[0].Bytes, tokenID, 0, metaData)
	vm2.CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, newMetaData)

	vm2.Log.Info("transfer dynamic NFT to the updated account")

	tx = vm2.EsdtNFTTransferTx(shard0Nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm2.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 0)
	vm2.CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, 1)
	vm2.CheckMetaDataNotInAcc(t, cs, addrs[0].Bytes, tokenID, 0)
	newMetaData.Attributes = metaData.Attributes
	newMetaData.Uris = metaData.Uris
	vm2.CheckMetaData(t, cs, addrs[1].Bytes, tokenID, 1, newMetaData)
}
