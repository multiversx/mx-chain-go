package esdt

import (
	"encoding/hex"
	"fmt"
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

func TestSovereignChainSimulator_RegisterTwoSfts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
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

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	sftName := "SFTNAME"
	sftTicker := "SFTTICKER"
	sftIdentifier := chainSim.RegisterAndSetAllRoles(t, cs, wallet.Bytes, &nonce, issueCost, sftName, sftTicker, core.SemiFungibleESDT, 0)

	checkAllRoles(t, nodeHandler, wallet.Bech32, sftIdentifier, sftRoles)

	initialSupply := big.NewInt(111)
	createArgs := createNftArgs(sftIdentifier, initialSupply, "SFTNAME #1")
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, initialSupply, tokens[sftIdentifier+"-01"].Value)

	sftName = "SFTNAME2"
	sftTicker = "SFTTICKER2"
	sftIdentifier = chainSim.RegisterAndSetAllRoles(t, cs, wallet.Bytes, &nonce, issueCost, sftName, sftTicker, core.SemiFungibleESDT, 0)

	checkAllRoles(t, nodeHandler, wallet.Bech32, sftIdentifier, sftRoles)

	initialSupply = big.NewInt(222)
	createArgs = createNftArgs(sftIdentifier, initialSupply, "SFTNAME #2")
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 3)
	require.Equal(t, initialSupply, tokens[sftIdentifier+"-01"].Value)
}

func TestSovereignChainSimulator_IssueTwoSfts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
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

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	sftName := "SFTNAME"
	sftTicker := "SFTTICKER"
	sftIdentifier := chainSim.IssueSemiFungible(t, cs, wallet.Bytes, &nonce, issueCost, sftName, sftTicker)

	setRolesArgs := setSpecialRole(sftIdentifier, wallet.Bytes, sftRoles)
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, chainSim.ZeroValue, setRolesArgs, uint64(60000000))

	checkAllRoles(t, nodeHandler, wallet.Bech32, sftIdentifier, sftRoles)

	initialSupply := big.NewInt(100)
	createArgs := createNftArgs(sftIdentifier, initialSupply, "SFTNAME #1")
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, initialSupply, tokens[sftIdentifier+"-01"].Value)

	sftName = "SFTNAME2"
	sftTicker = "SFTTICKER2"
	sftIdentifier = chainSim.IssueSemiFungible(t, cs, wallet.Bytes, &nonce, issueCost, sftName, sftTicker)

	setRolesArgs = setSpecialRole(sftIdentifier, wallet.Bytes, sftRoles)
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, chainSim.ZeroValue, setRolesArgs, uint64(60000000))

	checkAllRoles(t, nodeHandler, wallet.Bech32, sftIdentifier, sftRoles)

	initialSupply = big.NewInt(200)
	createArgs = createNftArgs(sftIdentifier, initialSupply, "SFTNAME2 #1")
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 3)
	require.Equal(t, initialSupply, tokens[sftIdentifier+"-01"].Value)
}

func createNftArgs(tokenIdentifier string, initialSupply *big.Int, name string) string {
	return "ESDTNFTCreate" +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(initialSupply.Bytes()) +
		"@" + hex.EncodeToString([]byte(name)) +
		"@" + fmt.Sprintf("%04X", 2500) + // royalties 25%
		"@" + // hash
		"@" + // attributes
		"@" // uri
}

func setSpecialRole(tokenIdentifier string, address []byte, roles []string) string {
	args := "setSpecialRole" +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(address)

	for _, role := range roles {
		args = args + "@" + hex.EncodeToString([]byte(role))
	}

	return args
}
