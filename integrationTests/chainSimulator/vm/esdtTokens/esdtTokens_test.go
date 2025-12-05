package esdtTokens

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/stretchr/testify/require"
)

type esdtTokensCompleteResponseData struct {
	Tokens map[string]groups.ESDTNFTTokenData `json:"esdts"`
}

type esdtTokensCompleteResponse struct {
	Data  esdtTokensCompleteResponseData `json:"data"`
	Error string                         `json:"error"`
	Code  string
}

func TestChainSimulator_Api_TokenType(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	activationEpoch := uint32(2)

	baseIssuingCost := "1000"

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            vm.DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          vm.RoundDurationInMillis,
		SupernovaRoundDurationInMillis: vm.SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 vm.RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        vm.SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewFreePortAPIConfigurator("localhost"),
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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	vm.Log.Info("Initial setup: Create tokens")

	addrs := vm.CreateAddresses(t, cs, false)

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}

	// issue fungible
	fungibleTicker := []byte("FUNTICKER")
	nonce := uint64(0)
	tx := vm.IssueTx(nonce, addrs[0].Bytes, fungibleTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	fungibleTokenID := txResult.Logs.Events[0].Topics[0]
	vm.SetAddressEsdtRoles(t, cs, nonce, addrs[0], fungibleTokenID, roles)
	nonce++

	vm.Log.Info("Issued fungible token id", "tokenID", string(fungibleTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = vm.IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	scrs, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetSCRsByTxHash(txResult.Hash, txResult.SmartContractResults[0].Hash)
	require.Nil(t, err)
	require.NotNil(t, scrs)
	require.Equal(t, len(txResult.SmartContractResults), len(scrs))

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	vm.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	vm.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = vm.IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	vm.SetAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

	vm.Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	fungibleMetaData := txsFee.GetDefaultMetaData()
	fungibleMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
	}

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
	}

	for i := range tokenIDs {
		tx = vm.EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	restAPIInterfaces := cs.GetRestAPIInterfaces()
	require.NotNil(t, restAPIInterfaces)

	url := fmt.Sprintf("http://%s/address/%s/esdt", restAPIInterfaces[shardID], addrs[0].Bech32)
	response := &esdtTokensCompleteResponse{}

	doHTTPClientGetReq(t, url, response)

	allTokens := response.Data.Tokens

	require.Equal(t, 3, len(allTokens))

	expTokenID := string(fungibleTokenID)
	tokenData, ok := allTokens[expTokenID]
	require.True(t, ok)
	require.Equal(t, expTokenID, tokenData.TokenIdentifier)
	require.Equal(t, core.FungibleESDT, tokenData.Type)

	expTokenID = string(nftTokenID) + "-01"
	tokenData, ok = allTokens[expTokenID]
	require.True(t, ok)
	require.Equal(t, expTokenID, tokenData.TokenIdentifier)
	require.Equal(t, core.NonFungibleESDTv2, tokenData.Type)

	expTokenID = string(sftTokenID) + "-01"
	tokenData, ok = allTokens[expTokenID]
	require.True(t, ok)
	require.Equal(t, expTokenID, tokenData.TokenIdentifier)
	require.Equal(t, core.SemiFungibleESDT, tokenData.Type)
}

func TestChainSimulator_Api_NFTToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	activationEpoch := uint32(2)

	baseIssuingCost := "1000"

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            vm.DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          vm.RoundDurationInMillis,
		SupernovaRoundDurationInMillis: vm.SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 vm.RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        vm.SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewFreePortAPIConfigurator("localhost"),
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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	vm.Log.Info("Initial setup: Create NFT token before activation")

	addrs := vm.CreateAddresses(t, cs, false)

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	nonce := uint64(0)
	tx := vm.IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	vm.SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	vm.Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = vm.EsdtNftCreateTx(nonce, addrs[0].Bytes, nftTokenID, nftMetaData, 1)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	restAPIInterfaces := cs.GetRestAPIInterfaces()
	require.NotNil(t, restAPIInterfaces)

	url := fmt.Sprintf("http://%s/address/%s/esdt", restAPIInterfaces[shardID], addrs[0].Bech32)
	response := &esdtTokensCompleteResponse{}

	doHTTPClientGetReq(t, url, response)

	allTokens := response.Data.Tokens

	require.Equal(t, 1, len(allTokens))

	expTokenID := string(nftTokenID) + "-01"
	tokenData, ok := allTokens[expTokenID]
	require.True(t, ok)
	require.Equal(t, expTokenID, tokenData.TokenIdentifier)
	require.Equal(t, "", tokenData.Type)

	vm.Log.Info("Wait for DynamicESDTFlag activation")

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	doHTTPClientGetReq(t, url, response)

	allTokens = response.Data.Tokens

	require.Equal(t, 1, len(allTokens))

	expTokenID = string(nftTokenID) + "-01"
	tokenData, ok = allTokens[expTokenID]
	require.True(t, ok)
	require.Equal(t, expTokenID, tokenData.TokenIdentifier)
	require.Equal(t, "", tokenData.Type)

	vm.Log.Info("Update token id", "tokenID", nftTokenID)

	tx = vm.UpdateTokenIDTx(nonce, addrs[0].Bytes, nftTokenID)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	doHTTPClientGetReq(t, url, response)

	allTokens = response.Data.Tokens

	require.Equal(t, 1, len(allTokens))

	expTokenID = string(nftTokenID) + "-01"
	tokenData, ok = allTokens[expTokenID]
	require.True(t, ok)
	require.Equal(t, expTokenID, tokenData.TokenIdentifier)
	require.Equal(t, "", tokenData.Type)

	vm.Log.Info("Transfer token id", "tokenID", nftTokenID)

	tx = vm.EsdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID)
	nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	url = fmt.Sprintf("http://%s/address/%s/esdt", restAPIInterfaces[1], addrs[1].Bech32)
	doHTTPClientGetReq(t, url, response)

	allTokens = response.Data.Tokens

	require.Equal(t, 1, len(allTokens))

	expTokenID = string(nftTokenID) + "-01"
	tokenData, ok = allTokens[expTokenID]
	require.True(t, ok)
	require.Equal(t, expTokenID, tokenData.TokenIdentifier)
	require.Equal(t, core.NonFungibleESDTv2, tokenData.Type)

	vm.Log.Info("Change to DYNAMIC type")

	tx = vm.ChangeToDynamicTx(nonce, addrs[0].Bytes, nftTokenID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	response = &esdtTokensCompleteResponse{}
	doHTTPClientGetReq(t, url, response)

	allTokens = response.Data.Tokens

	require.Equal(t, 1, len(allTokens))

	expTokenID = string(nftTokenID) + "-01"
	tokenData, ok = allTokens[expTokenID]
	require.True(t, ok)
	require.Equal(t, expTokenID, tokenData.TokenIdentifier)
	require.Equal(t, core.NonFungibleESDTv2, tokenData.Type)
}

func doHTTPClientGetReq(t *testing.T, url string, response interface{}) {
	httpClient := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.Nil(t, err)

	resp, err := httpClient.Do(req)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	jsonParser := json.NewDecoder(resp.Body)
	err = jsonParser.Decode(&response)
	require.Nil(t, err)
}
