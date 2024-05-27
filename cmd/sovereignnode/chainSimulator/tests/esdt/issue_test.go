package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(100))
var issueCost = big.NewInt(5000000000000000000)

func TestEsdt_Issue(t *testing.T) {
	epochConfig, economicsConfig, sovereignExtraConfig, err := sovereignChainSimulator.LoadSovereignConfigs(sovereignConfigPath)
	require.Nil(t, err)

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignExtraConfig: *sovereignExtraConfig,
		ChainSimulatorArgs: chainSimulator.ArgsChainSimulator{
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
				cfg.EconomicsConfig = economicsConfig
				cfg.EpochConfig = epochConfig
				cfg.GeneralConfig.SovereignConfig = *sovereignExtraConfig
				cfg.GeneralConfig.VirtualMachine.Execution.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
				cfg.GeneralConfig.VirtualMachine.Querying.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemEsdtAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")

	nonce := uint64(0)
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)

	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	issueArgs := "issue" +
		"@" + hex.EncodeToString([]byte("SovereignTkn")) +
		"@" + hex.EncodeToString([]byte("SVN")) +
		"@" + hex.EncodeToString(supply.Bytes()) +
		"@" + fmt.Sprintf("%X", 18) +
		"@" + hex.EncodeToString([]byte("canAddSpecialRoles")) +
		"@" + hex.EncodeToString([]byte("true"))
	txResult := chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]
	require.True(t, len(tokenIdentifier) > 7)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	esdts, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs(core.FungibleESDT)
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, string(tokenIdentifier) == esdts[0])

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, supply, tokens[string(tokenIdentifier)].Value)

	setRolesArgs := "setSpecialRole" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(wallet.Bytes) +
		"@" + hex.EncodeToString([]byte(core.ESDTRoleLocalMint)) +
		"@" + hex.EncodeToString([]byte(core.ESDTRoleLocalBurn)) +
		"@" + hex.EncodeToString([]byte(core.ESDTRoleTransfer))
	txResult = chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, big.NewInt(0), setRolesArgs, uint64(60000000))

	esdtsRoles, _, err := nodeHandler.GetFacadeHandler().GetESDTsRoles(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdtsRoles)
	require.True(t, len(esdtsRoles[string(tokenIdentifier)]) == 3)
	require.Equal(t, core.ESDTRoleLocalMint, esdtsRoles[string(tokenIdentifier)][0])
	require.Equal(t, core.ESDTRoleLocalBurn, esdtsRoles[string(tokenIdentifier)][1])
	require.Equal(t, core.ESDTRoleTransfer, esdtsRoles[string(tokenIdentifier)][2])
}
