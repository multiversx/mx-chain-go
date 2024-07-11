package epochChange

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
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
			MinNodesPerShard:       8,
			ConsensusGroupSize:     8,
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

	trie := cs.GetNodeHandler(0).GetStateComponents().TriesContainer().Get([]byte(dataRetriever.PeerAccountsUnit.String()))
	require.NotNil(t, trie)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)
	require.Equal(t, uint32(1), cs.GetNodeHandler(0).GetCoreComponents().EpochNotifier().CurrentEpoch())

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	logger.SetLogLevel("*:DEBUG")
	stakeNodes(t, cs, 10)

	err = cs.GenerateBlocksUntilEpochIsReached(6)
	require.Nil(t, err)
}

func stakeNodes(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, numNodesToStake int) {
	txs := make([]*transaction.Transaction, numNodesToStake)
	for i := 0; i < numNodesToStake; i++ {
		txs[i] = createStakeTransaction(t, cs)
	}

	stakeTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted(txs, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTxs)
	require.Len(t, stakeTxs, numNodesToStake)

	require.Nil(t, cs.GenerateBlocks(1))
}

func stakeOneNode(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator) {
	txStake := createStakeTransaction(t, cs)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	require.Nil(t, cs.GenerateBlocks(1))
}

func createStakeTransaction(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator) *transaction.Transaction {
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Add(chainSimulatorIntegrationTests.MinimumStakeValue, chainSimulatorIntegrationTests.OneEGLD)
	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	return chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
}
