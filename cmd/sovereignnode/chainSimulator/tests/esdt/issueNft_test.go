package esdt

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
)

var nftV2Roles = []string{
	core.ESDTRoleNFTCreate,
	core.ESDTRoleNFTBurn,
	core.ESDTRoleNFTUpdateAttributes,
	core.ESDTRoleNFTAddURI,
	core.ESDTRoleNFTRecreate,
	core.ESDTRoleModifyCreator,
	core.ESDTRoleModifyRoyalties,
	core.ESDTRoleSetNewURI,
	core.ESDTRoleNFTUpdate,
}

func TestSovereignChainSimulator_RegisterNftWithPrefix(t *testing.T) {
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
	nftName := "NFTNAME"
	nftTicker := "NFTTICKER"
	nftIdentifier := chainSim.RegisterAndSetAllRoles(t, cs, wallet.Bytes, &nonce, issueCost, nftName, nftTicker, core.NonFungibleESDTv2, 0)

	checkAllRoles(t, nodeHandler, wallet.Bech32, nftIdentifier, nftV2Roles)

	initialSupply := big.NewInt(1)
	createArgs := createNftArgs(nftIdentifier, initialSupply, "NFTNAME #1")
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, createArgs, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, initialSupply, tokens[nftIdentifier+"-01"].Value)
}
