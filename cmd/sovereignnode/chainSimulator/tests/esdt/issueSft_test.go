package esdt

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/vm"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

var sftRoles = []string{
	core.ESDTRoleNFTCreate,
	core.ESDTRoleNFTBurn,
	core.ESDTRoleNFTAddQuantity,
}

func TestSovereignChain_RegisterTwoSfts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
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

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	registerArgs := createSftRegisterArgs("SFTNAME", "SFTTICKER")
	registerTx := chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, issueCost, registerArgs, uint64(60000000))
	tokenIdentifier := registerTx.Logs.Events[0].Topics[0]

	checkAllRoles(t, nodeHandler, wallet.Bech32, string(tokenIdentifier), sftRoles)

	initialSupply := big.NewInt(111)
	createArgs := createNftArgs(tokenIdentifier, initialSupply, "SFTNAME #1")
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, initialSupply, tokens[string(tokenIdentifier)+"-01"].Value)

	registerArgs = createSftRegisterArgs("SFTNAME2", "SFTTICKER2")
	registerTx = chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, issueCost, registerArgs, uint64(60000000))
	tokenIdentifier = registerTx.Logs.Events[0].Topics[0]

	checkAllRoles(t, nodeHandler, wallet.Bech32, string(tokenIdentifier), sftRoles)

	initialSupply = big.NewInt(222)
	createArgs = createNftArgs(tokenIdentifier, initialSupply, "SFTNAME #2")
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 3)
	require.Equal(t, initialSupply, tokens[string(tokenIdentifier)+"-01"].Value)
}

func TestSovereignChain_IssueTwoSfts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
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

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	issueArgs := createIssueSftArgs("SFTNAME", "SFTTIKER")
	txResult := chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]

	setRolesArgs := setSpecialRole(tokenIdentifier, wallet.Bytes, sftRoles)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, chainSim.ZeroValue, setRolesArgs, uint64(60000000))

	checkAllRoles(t, nodeHandler, wallet.Bech32, string(tokenIdentifier), sftRoles)

	initialSupply := big.NewInt(100)
	createArgs := createNftArgs(tokenIdentifier, initialSupply, "SFTNAME #1")
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, initialSupply, tokens[string(tokenIdentifier)+"-01"].Value)

	issueArgs = createIssueSftArgs("SFTNAME2", "SFTTIKER2")
	txResult = chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier = txResult.Logs.Events[0].Topics[0]

	setRolesArgs = setSpecialRole(tokenIdentifier, wallet.Bytes, sftRoles)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, chainSim.ZeroValue, setRolesArgs, uint64(60000000))

	checkAllRoles(t, nodeHandler, wallet.Bech32, string(tokenIdentifier), sftRoles)

	initialSupply = big.NewInt(200)
	createArgs = createNftArgs(tokenIdentifier, initialSupply, "SFTNAME2 #1")
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 3)
	require.Equal(t, initialSupply, tokens[string(tokenIdentifier)+"-01"].Value)
}

func createSftRegisterArgs(name string, ticker string) string {
	return "registerAndSetAllRoles" +
		"@" + hex.EncodeToString([]byte(name)) + // name
		"@" + hex.EncodeToString([]byte(ticker)) + // ticker
		"@" + hex.EncodeToString([]byte("SFT")) + // type
		"@" // num decimals
}

func createIssueSftArgs(name string, ticker string) string {
	return "issueSemiFungible" +
		"@" + hex.EncodeToString([]byte(name)) + // name
		"@" + hex.EncodeToString([]byte(ticker)) // ticker
}

func createNftArgs(tokenIdentifier []byte, initialSupply *big.Int, name string) string {
	return "ESDTNFTCreate" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(initialSupply.Bytes()) + // initial quantity
		"@" + hex.EncodeToString([]byte(name)) + // name
		"@09c4" + // royalties 25%
		"@" + // hash
		"@" + // attributes
		"@" // uri
}

func setSpecialRole(tokenIdentifier []byte, address []byte, roles []string) string {
	args := "setSpecialRole" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(address)

	for _, role := range roles {
		args = args + "@" + hex.EncodeToString([]byte(role))
	}

	return args
}
