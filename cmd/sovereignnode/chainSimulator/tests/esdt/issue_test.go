package esdt

import (
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

func TestSovereignChainSimulator_IssueFungible(t *testing.T) {
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

	nonce := uint64(0)
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	tokenName := "SovereignTkn"
	tokenTicker := "SVN"
	numDecimals := 18
	tokenIdentifier := chainSim.IssueFungible(t, cs, nodeHandler, wallet.Bytes, &nonce, issueCost, tokenName, tokenTicker, numDecimals, supply)

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, supply, tokens[tokenIdentifier].Value)

	setRolesArgs := setSpecialRole(tokenIdentifier, wallet.Bytes, fungibleRoles)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, big.NewInt(0), setRolesArgs, uint64(60000000))

	checkAllRoles(t, nodeHandler, wallet.Bech32, tokenIdentifier, fungibleRoles)
}
