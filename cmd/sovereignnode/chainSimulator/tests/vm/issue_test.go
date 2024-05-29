package vm

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	issueWasmPath              = "../testdata/issue.wasm"
	issuePrice                 = "5000000000000000000"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(100))

func TestSmartContract_IssueToken(t *testing.T) {
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
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(200)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	deployedContractAddress := chainSimulatorIntegrationTests.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", issueWasmPath)
	deployedContractAddressBech32, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(deployedContractAddress)
	require.Nil(t, err)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, issueCost, "issue", uint64(60000000))

	accountSC, _, err := nodeHandler.GetFacadeHandler().GetAccount(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.True(t, accountSC.Balance == "0")

	esdts, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs(core.FungibleESDT)
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, len(esdts) == 1)

	tokenIdentifier := esdts[0]
	setRolesArgs := "setRoles@" + hex.EncodeToString([]byte(tokenIdentifier))
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, big.NewInt(0), setRolesArgs, uint64(60000000))

	esdtsRoles, _, err := nodeHandler.GetFacadeHandler().GetESDTsRoles(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdtsRoles)
	require.True(t, len(esdtsRoles[tokenIdentifier]) == 3)
	require.Equal(t, core.ESDTRoleLocalMint, esdtsRoles[tokenIdentifier][0])
	require.Equal(t, core.ESDTRoleLocalBurn, esdtsRoles[tokenIdentifier][1])
	require.Equal(t, core.ESDTRoleTransfer, esdtsRoles[tokenIdentifier][2])

	expectedMintedAmount, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	mintTxArgs := "mint" +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(expectedMintedAmount.Bytes())
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, big.NewInt(0), mintTxArgs, uint64(20000000))

	accountSC, _, err = nodeHandler.GetFacadeHandler().GetAccount(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.True(t, accountSC.Balance == "0")

	esdtSC, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotEmpty(t, esdtSC)
	require.Equal(t, expectedMintedAmount, esdtSC[esdts[0]].Value)
}

func TestSmartContract_IssueToken_MainChain(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 1,
		GenesisTimestamp:            time.Now().Unix(),
		RoundDurationInMillis:       uint64(6000),
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            2,
		MetaChainMinNodes:           2,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t",
		},
	})
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(5)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	wallet, err := cs.GenerateAndMintWalletAddress(0, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	deployedContractAddress := chainSimulatorIntegrationTests.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", issueWasmPath)
	deployedContractAddressBech32, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(deployedContractAddress)
	require.Nil(t, err)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, issueCost, "issue", uint64(60000000))

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	accountSC, _, err := nodeHandler.GetFacadeHandler().GetAccount(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.True(t, accountSC.Balance == "0")

	esdts, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetAllIssuedESDTs(core.FungibleESDT)
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, len(esdts) == 1)

	tokenIdentifier := esdts[0]
	setRolesArgs := "setRoles@" + hex.EncodeToString([]byte(tokenIdentifier))
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, big.NewInt(0), setRolesArgs, uint64(60000000))

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	esdtsRoles, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetESDTsRoles(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdtsRoles)
	require.True(t, len(esdtsRoles[tokenIdentifier]) == 3)
	require.Equal(t, core.ESDTRoleLocalMint, esdtsRoles[tokenIdentifier][0])
	require.Equal(t, core.ESDTRoleLocalBurn, esdtsRoles[tokenIdentifier][1])
	require.Equal(t, core.ESDTRoleTransfer, esdtsRoles[tokenIdentifier][2])

	expectedMintedAmount, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	mintTxArgs := "mint" +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(expectedMintedAmount.Bytes())
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, big.NewInt(0), mintTxArgs, uint64(20000000))

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	accountSC, _, err = nodeHandler.GetFacadeHandler().GetAccount(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.True(t, accountSC.Balance == "0")

	esdtSC, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotEmpty(t, esdtSC)
	require.Equal(t, expectedMintedAmount, esdtSC[esdts[0]].Value)
}
