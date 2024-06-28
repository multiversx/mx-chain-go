package esdt

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	api2 "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

const (
	defaultPathToInitialConfig              = "../../../cmd/node/config/"
	vmTypeHex                               = "0500"
	codeMetadata                            = "0500"
	maxNumOfBlocksToGenerateWhenExecutingTx = 20
)

var oneEGLD = big.NewInt(1000000000000000000)
var zeroValue = big.NewInt(0)

func TestChainSimulator_UpdateCrossShardAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

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
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(0)

	_ = cs.GenerateBlocksUntilEpochIsReached(3)

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(oneEGLD, mintValue)

	address0, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)

	address1, err := cs.GenerateAndMintWalletAddress(1, mintValue)
	require.Nil(t, err)

	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	// deploy SC by address in shard 0
	data := wasm.GetSCCode("testdata/update-attributes.wasm") + "@" + vmTypeHex + "@" + codeMetadata
	tx := chainSim.GenerateTransaction(address0.Bytes, 0, systemScAddress, zeroValue, data, uint64(200000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	scAddress := txResult.Logs.Events[0].Topics[0]
	scAddressBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(scAddress)
	_ = scAddressBech32

	// issue NFT collection by address in shard 0
	issueCost, _ := big.NewInt(0).SetString("5000000000000000000", 10)
	data = "issue@" + hex.EncodeToString([]byte("NFTCOL")) + "@" + hex.EncodeToString([]byte("NCL"))
	tx = chainSim.GenerateTransaction(address0.Bytes, 1, scAddress, issueCost, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	cs.GenerateBlocks(5)

	// create NFT by address in shard 0 ----> will send the minted nft to address in shard 1
	data = "create@" + hex.EncodeToString(address1.Bytes)
	tx = chainSim.GenerateTransaction(address0.Bytes, 2, scAddress, zeroValue, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	issuedESDTs, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetAllIssuedESDTs(core.NonFungibleESDT)
	require.Nil(t, err)
	require.NotNil(t, issuedESDTs)
	require.True(t, len(issuedESDTs) == 1)
	nftCollection := issuedESDTs[0]

	cs.GenerateBlocks(5)

	// check NFT received by address in shard 1 ---> default attributes value in SC is "common"
	accountsEsdts, _, err := cs.GetNodeHandler(1).GetFacadeHandler().GetAllESDTTokens(address1.Bech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountsEsdts)
	require.Equal(t, 1, len(accountsEsdts))
	nftIdentifier := nftCollection + "-01"
	nftAttributes := accountsEsdts[nftIdentifier].TokenMetaData.Attributes
	require.Equal(t, "common", string(nftAttributes))

	//update attribute for NFT -> cross-shard update
	expectedNewAttribute := "attribute1"
	data = "ESDTNFTTransfer" +
		"@" + hex.EncodeToString([]byte(nftCollection)) + //nft collection
		"@01" + // nonce
		"@01" + // amount
		"@" + hex.EncodeToString(scAddress) + // sc address
		"@" + hex.EncodeToString([]byte("update")) +
		"@" + hex.EncodeToString([]byte(expectedNewAttribute))
	tx = chainSim.GenerateTransaction(address1.Bytes, 0, address1.Bytes, zeroValue, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	cs.GenerateBlocks(10)

	accountsEsdts, _, err = cs.GetNodeHandler(1).GetFacadeHandler().GetAllESDTTokens(address1.Bech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountsEsdts)
	require.Equal(t, 1, len(accountsEsdts))
	nftAttributes = accountsEsdts[nftIdentifier].TokenMetaData.Attributes
	assert.Equal(t, expectedNewAttribute, string(nftAttributes))

	cs.GenerateBlocks(5)

	//update attribute for NFT -> cross-shard update
	expectedNewAttribute = "attribute2"
	data = "ESDTNFTTransfer" +
		"@" + hex.EncodeToString([]byte(nftCollection)) + //nft collection
		"@01" + // nonce
		"@01" + // amount
		"@" + hex.EncodeToString(scAddress) + // sc address
		"@" + hex.EncodeToString([]byte("update")) +
		"@" + hex.EncodeToString([]byte(expectedNewAttribute))
	tx = chainSim.GenerateTransaction(address1.Bytes, 1, address1.Bytes, zeroValue, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	cs.GenerateBlocks(10)

	accountsEsdts, _, err = cs.GetNodeHandler(1).GetFacadeHandler().GetAllESDTTokens(address1.Bech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountsEsdts)
	require.Equal(t, 1, len(accountsEsdts))
	nftAttributes = accountsEsdts[nftIdentifier].TokenMetaData.Attributes
	assert.Equal(t, expectedNewAttribute, string(nftAttributes))

	cs.GenerateBlocks(5)

	//update attribute for NFT -> cross-shard update
	expectedNewAttribute = "attribute3"
	data = "ESDTNFTTransfer" +
		"@" + hex.EncodeToString([]byte(nftCollection)) + //nft collection
		"@01" + // nonce
		"@01" + // amount
		"@" + hex.EncodeToString(scAddress) + // sc address
		"@" + hex.EncodeToString([]byte("update")) +
		"@" + hex.EncodeToString([]byte(expectedNewAttribute))
	tx = chainSim.GenerateTransaction(address1.Bytes, 2, address1.Bytes, zeroValue, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	cs.GenerateBlocks(10)

	accountsEsdts, _, err = cs.GetNodeHandler(1).GetFacadeHandler().GetAllESDTTokens(address1.Bech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountsEsdts)
	require.Equal(t, 1, len(accountsEsdts))
	nftAttributes = accountsEsdts[nftIdentifier].TokenMetaData.Attributes
	assert.Equal(t, expectedNewAttribute, string(nftAttributes))
}

func TestChainSimulator_UpdateSameShardAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

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
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(0)

	_ = cs.GenerateBlocksUntilEpochIsReached(3)

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(oneEGLD, mintValue)

	address0, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)

	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	// deploy SC by address in shard 0
	data := wasm.GetSCCode("testdata/update-attributes.wasm") + "@" + vmTypeHex + "@" + codeMetadata
	tx := chainSim.GenerateTransaction(address0.Bytes, 0, systemScAddress, zeroValue, data, uint64(200000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	scAddress := txResult.Logs.Events[0].Topics[0]
	scAddressBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(scAddress)
	_ = scAddressBech32

	// issue NFT collection by address in shard 0
	issueCost, _ := big.NewInt(0).SetString("5000000000000000000", 10)
	data = "issue@" + hex.EncodeToString([]byte("NFTCOL")) + "@" + hex.EncodeToString([]byte("NCL"))
	tx = chainSim.GenerateTransaction(address0.Bytes, 1, scAddress, issueCost, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	cs.GenerateBlocks(5)

	// create NFT by address in shard 0 ----> will send the minted nft to address in shard 1
	data = "create@" + hex.EncodeToString(address0.Bytes)
	tx = chainSim.GenerateTransaction(address0.Bytes, 2, scAddress, zeroValue, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	issuedESDTs, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetAllIssuedESDTs(core.NonFungibleESDT)
	require.Nil(t, err)
	require.NotNil(t, issuedESDTs)
	require.True(t, len(issuedESDTs) == 1)
	nftCollection := issuedESDTs[0]

	cs.GenerateBlocks(5)

	// check NFT received by address in shard 1 ---> default attributes value in SC is "common"
	accountsEsdts, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(address0.Bech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountsEsdts)
	require.Equal(t, 1, len(accountsEsdts))
	nftIdentifier := nftCollection + "-01"
	nftAttributes := accountsEsdts[nftIdentifier].TokenMetaData.Attributes
	require.Equal(t, "common", string(nftAttributes))

	//update attribute for NFT -> same-shard update
	expectedNewAttribute := "attribute1"
	data = "ESDTNFTTransfer" +
		"@" + hex.EncodeToString([]byte(nftCollection)) + //nft collection
		"@01" + // nonce
		"@01" + // amount
		"@" + hex.EncodeToString(scAddress) + // sc address
		"@" + hex.EncodeToString([]byte("update")) +
		"@" + hex.EncodeToString([]byte(expectedNewAttribute))
	tx = chainSim.GenerateTransaction(address0.Bytes, 3, address0.Bytes, zeroValue, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	cs.GenerateBlocks(10)

	accountsEsdts, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(address0.Bech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountsEsdts)
	require.Equal(t, 1, len(accountsEsdts))
	nftAttributes = accountsEsdts[nftIdentifier].TokenMetaData.Attributes
	assert.Equal(t, expectedNewAttribute, string(nftAttributes))

	cs.GenerateBlocks(5)

	//update attribute for NFT -> same-shard update
	expectedNewAttribute = "attribute2"
	data = "ESDTNFTTransfer" +
		"@" + hex.EncodeToString([]byte(nftCollection)) + //nft collection
		"@01" + // nonce
		"@01" + // amount
		"@" + hex.EncodeToString(scAddress) + // sc address
		"@" + hex.EncodeToString([]byte("update")) +
		"@" + hex.EncodeToString([]byte(expectedNewAttribute))
	tx = chainSim.GenerateTransaction(address0.Bytes, 4, address0.Bytes, zeroValue, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	cs.GenerateBlocks(10)

	accountsEsdts, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(address0.Bech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountsEsdts)
	require.Equal(t, 1, len(accountsEsdts))
	nftAttributes = accountsEsdts[nftIdentifier].TokenMetaData.Attributes
	assert.Equal(t, expectedNewAttribute, string(nftAttributes))

	cs.GenerateBlocks(5)

	//update attribute for NFT -> same-shard update
	expectedNewAttribute = "attribute3"
	data = "ESDTNFTTransfer" +
		"@" + hex.EncodeToString([]byte(nftCollection)) + //nft collection
		"@01" + // nonce
		"@01" + // amount
		"@" + hex.EncodeToString(scAddress) + // sc address
		"@" + hex.EncodeToString([]byte("update")) +
		"@" + hex.EncodeToString([]byte(expectedNewAttribute))
	tx = chainSim.GenerateTransaction(address0.Bytes, 5, address0.Bytes, zeroValue, data, uint64(60000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	cs.GenerateBlocks(10)

	accountsEsdts, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(address0.Bech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountsEsdts)
	require.Equal(t, 1, len(accountsEsdts))
	nftAttributes = accountsEsdts[nftIdentifier].TokenMetaData.Attributes
	assert.Equal(t, expectedNewAttribute, string(nftAttributes))
}
