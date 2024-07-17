package epochChange

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
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
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         roundsPerEpoch,
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				newCfg := config.EnableEpochs{}
				newCfg.BLSMultiSignerEnableEpoch = cfg.EpochConfig.EnableEpochs.BLSMultiSignerEnableEpoch
				newCfg.MaxNodesChangeEnableEpoch = cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch

				cfg.EpochConfig.EnableEpochs = newCfg
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	trie := cs.GetNodeHandler(0).GetStateComponents().TriesContainer().Get([]byte(dataRetriever.PeerAccountsUnit.String()))
	require.NotNil(t, trie)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)
	require.Equal(t, uint32(1), cs.GetNodeHandler(0).GetCoreComponents().EpochNotifier().CurrentEpoch())

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)
}
