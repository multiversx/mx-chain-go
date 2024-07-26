package epochChange

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
)

func TestSovereignChainSimulator_EpochChange(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    uint64(6000),
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         6,
			ConsensusGroupSize:       6,
			NumNodesWaitingListShard: 2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				newCfg := config.EnableEpochs{}
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
				newCfg.BLSMultiSignerEnableEpoch = cfg.EpochConfig.EnableEpochs.BLSMultiSignerEnableEpoch
				newCfg.MaxNodesChangeEnableEpoch = []config.MaxNodesChangeConfig{
					{
						EpochEnable:            0,
						MaxNumNodes:            8,
						NodesToShufflePerShard: 2,
					},
				}

				cfg.EpochConfig.EnableEpochs = newCfg
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	nonce := uint64(0)
	require.Nil(t, err)

	trie := nodeHandler.GetStateComponents().TriesContainer().Get([]byte(dataRetriever.PeerAccountsUnit.String()))
	require.NotNil(t, trie)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)
	require.Equal(t, uint32(1), nodeHandler.GetCoreComponents().EpochNotifier().CurrentEpoch())

	accFeesInEpoch, devFeesInEpoch := getAllFeesInEpoch(nodeHandler)
	require.Empty(t, accFeesInEpoch.Bytes())
	require.Empty(t, devFeesInEpoch.Bytes())

	staking.StakeNodes(t, cs, 10)
	err = nodeHandler.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)

	auctionList, err := nodeHandler.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)
	require.Len(t, auctionList, 10)

	validators := nodeHandler.GetProcessComponents().ValidatorsProvider().GetLatestValidators()
	require.Len(t, validators, 18)

	// check here fees increase after stake txs
	accFeesInEpoch, devFeesInEpoch = getAllFeesInEpoch(nodeHandler)
	require.NotEmpty(t, accFeesInEpoch)
	require.NotEmpty(t, devFeesInEpoch)

	currentEpoch := nodeHandler.GetCoreComponents().EpochNotifier().CurrentEpoch()
	for epoch := currentEpoch + 1; epoch < currentEpoch+6; epoch++ {
		err = cs.GenerateBlocksUntilEpochIsReached(int32(epoch))
		require.Nil(t, err)

		qualified, unqualified := staking.GetQualifiedAndUnqualifiedNodes(t, nodeHandler)
		require.Len(t, qualified, 2)
		require.Len(t, unqualified, 8)

		chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, "data", uint64(10000000))

		accFeesInEpoch, devFeesInEpoch = getAllFeesInEpoch(nodeHandler)
		require.NotEmpty(t, accFeesInEpoch)
		require.Empty(t, devFeesInEpoch)

		accFeesTotal, devFeesTotal := getAllFees(nodeHandler)
		require.NotEmpty(t, accFeesTotal.Bytes())
		require.Empty(t, devFeesTotal.Bytes())
	}
}

func getAllFeesInEpoch(nodeHandler process.NodeHandler) (*big.Int, *big.Int) {
	sovHdr := getCurrSovHdr(nodeHandler)
	return sovHdr.GetAccumulatedFeesInEpoch(), sovHdr.GetDevFeesInEpoch()
}

func getAllFees(nodeHandler process.NodeHandler) (*big.Int, *big.Int) {
	sovHdr := getCurrSovHdr(nodeHandler)
	return sovHdr.GetAccumulatedFees(), sovHdr.GetDeveloperFees()
}

func getCurrSovHdr(nodeHandler process.NodeHandler) data.SovereignChainHeaderHandler {
	return nodeHandler.GetChainHandler().GetCurrentBlockHeader().(data.SovereignChainHeaderHandler)
}
