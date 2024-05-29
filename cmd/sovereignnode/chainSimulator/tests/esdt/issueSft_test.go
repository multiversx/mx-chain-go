package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

func TestEsdt_RegisterSft(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ChainSimulatorArgs: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			NumOfShards:            1,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemEsdtAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	registerArgs := "registerAndSetAllRoles" +
		"@" + hex.EncodeToString([]byte("SFTNAME")) + // name
		"@" + hex.EncodeToString([]byte("SFTTIKER")) + // ticker
		"@" + hex.EncodeToString([]byte("SFT")) + // type
		"@" // num decimals
	registerTx := chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, registerArgs, uint64(60000000))
	tokenIdentifier := registerTx.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	createArgs := "ESDTNFTCreate" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + fmt.Sprintf("%X", 100) + // initial quantity
		"@" + hex.EncodeToString([]byte("SFTNAME #1")) + // name
		"@09c4" + // royalties 25%
		"@" + // hash
		"@" + // attributes
		"@" // uri
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), createArgs, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, big.NewInt(100), tokens[string(tokenIdentifier)+"-01"].Value)
}

func TestEsdt_RegisterTwoSfts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ChainSimulatorArgs: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			NumOfShards:            1,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemEsdtAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	registerArgs := "registerAndSetAllRoles" +
		"@" + hex.EncodeToString([]byte("SFTNAME")) + // name
		"@" + hex.EncodeToString([]byte("SFTTIKER")) + // ticker
		"@" + hex.EncodeToString([]byte("SFT")) + // type
		"@" // num decimals
	registerTx := chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, registerArgs, uint64(60000000))
	tokenIdentifier := registerTx.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	createArgs := "ESDTNFTCreate" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + fmt.Sprintf("%X", 100) + // initial quantity
		"@" + hex.EncodeToString([]byte("SFTNAME #1")) + // name
		"@09c4" + // royalties 25%
		"@" + // hash
		"@" + // attributes
		"@" // uri
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), createArgs, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, big.NewInt(100), tokens[string(tokenIdentifier)+"-01"].Value)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	registerArgs = "registerAndSetAllRoles" +
		"@" + hex.EncodeToString([]byte("SFTNAME2")) + // name
		"@" + hex.EncodeToString([]byte("SFTTIKER2")) + // ticker
		"@" + hex.EncodeToString([]byte("SFT")) + // type
		"@" // num decimals
	registerTx = chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, registerArgs, uint64(60000000))
	tokenIdentifier = registerTx.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	createArgs = "ESDTNFTCreate" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + fmt.Sprintf("%X", 222) + // initial quantity
		"@" + hex.EncodeToString([]byte("SFTNAME2 #1")) + // name
		"@09c4" + // royalties 25%
		"@" + // hash
		"@" + // attributes
		"@" // uri
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), createArgs, uint64(60000000))

	tokens, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 3)
	require.Equal(t, big.NewInt(222), tokens[string(tokenIdentifier)+"-01"].Value)
}

func TestEsdt_IssueSft(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ChainSimulatorArgs: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			NumOfShards:            1,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemEsdtAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	issueArgs := "issueSemiFungible" +
		"@" + hex.EncodeToString([]byte("SFTNAME")) + // name
		"@" + hex.EncodeToString([]byte("SFTTIKER")) // ticker
	txResult := chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]

	setRolesArgs := "setSpecialRole" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(wallet.Bytes) +
		"@" + hex.EncodeToString([]byte(core.ESDTRoleNFTCreate)) +
		"@" + hex.EncodeToString([]byte(core.ESDTRoleNFTAddQuantity))
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, big.NewInt(0), setRolesArgs, uint64(60000000))

	esdtsRoles, _, err := nodeHandler.GetFacadeHandler().GetESDTsRoles(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdtsRoles)
	require.True(t, len(esdtsRoles[string(tokenIdentifier)]) == 2)
	require.Equal(t, core.ESDTRoleNFTCreate, esdtsRoles[string(tokenIdentifier)][0])
	require.Equal(t, core.ESDTRoleNFTAddQuantity, esdtsRoles[string(tokenIdentifier)][1])

	createArgs := "ESDTNFTCreate" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + fmt.Sprintf("%X", 100) + // initial quantity
		"@" + hex.EncodeToString([]byte("SFTNAME #1")) + // name
		"@09c4" + // royalties 25%
		"@" + // hash
		"@" + // attributes
		"@" // uri
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), createArgs, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, big.NewInt(100), tokens[string(tokenIdentifier)+"-01"].Value)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	issueArgs = "issueSemiFungible" +
		"@" + hex.EncodeToString([]byte("SFTNAME2")) + // name
		"@" + hex.EncodeToString([]byte("SFTTIKER2")) // ticker
	txResult = chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier = txResult.Logs.Events[0].Topics[0]

	setRolesArgs = "setSpecialRole" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(wallet.Bytes) +
		"@" + hex.EncodeToString([]byte(core.ESDTRoleNFTCreate)) +
		"@" + hex.EncodeToString([]byte(core.ESDTRoleNFTAddQuantity))
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, big.NewInt(0), setRolesArgs, uint64(60000000))

	esdtsRoles, _, err = nodeHandler.GetFacadeHandler().GetESDTsRoles(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdtsRoles)
	require.True(t, len(esdtsRoles[string(tokenIdentifier)]) == 2)
	require.Equal(t, core.ESDTRoleNFTCreate, esdtsRoles[string(tokenIdentifier)][0])
	require.Equal(t, core.ESDTRoleNFTAddQuantity, esdtsRoles[string(tokenIdentifier)][1])

	createArgs = "ESDTNFTCreate" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + fmt.Sprintf("%X", 222) + // initial quantity
		"@" + hex.EncodeToString([]byte("SFTNAME2 #1")) + // name
		"@09c4" + // royalties 25%
		"@" + // hash
		"@" + // attributes
		"@" // uri
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), createArgs, uint64(60000000))

	tokens, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 3)
	require.Equal(t, big.NewInt(222), tokens[string(tokenIdentifier)+"-01"].Value)
}
