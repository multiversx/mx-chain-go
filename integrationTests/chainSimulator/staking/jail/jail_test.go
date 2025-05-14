package jail

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/vm"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	stakingV4JailUnJailStep1EnableEpoch = 9
	defaultPathToInitialConfig          = "../../../../cmd/node/config/"
	epochWhenNodeIsJailed               = 8
)

// Test description
// All test cases will do a stake transaction and wait till the new node is jailed
// testcase1 -- unJail transaction will be sent when staking v3.5 is still action --> node status should be `new` after unjail
// testcase2 -- unJail transaction will be sent when staking v4 step1 is action --> node status should be `auction` after unjail
// testcase3 -- unJail transaction will be sent when staking v4 step2 is action --> node status should be `auction` after unjail
// testcase4 -- unJail transaction will be sent when staking v4 step3 is action --> node status should be `auction` after unjail
func TestChainSimulator_ValidatorJailUnJail(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		testChainSimulatorJailAndUnJail(t, 8, "new")
	})

	t.Run("staking ph 4 step 1 active", func(t *testing.T) {
		testChainSimulatorJailAndUnJail(t, 9, "auction")
	})

	t.Run("staking ph 4 step 2 active", func(t *testing.T) {
		testChainSimulatorJailAndUnJail(t, 10, "auction")
	})

	t.Run("staking ph 4 step 3 active", func(t *testing.T) {
		testChainSimulatorJailAndUnJail(t, 11, "auction")
	})
}

func testChainSimulatorJailAndUnJail(t *testing.T, targetEpoch int32, nodeStatusAfterUnJail string) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	numOfShards := uint32(3)

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       4,
		MetaChainMinNodes:      4,
		AlterConfigsFunction: func(cfg *config.Configs) {
			configs.SetStakingV4ActivationEpochs(cfg, stakingV4JailUnJailStep1EnableEpoch)
			cfg.EpochConfig.EnableEpochs.AndromedaEnableEpoch = 100
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 8 // 8 nodes until new nodes will be placed on queue
			configs.SetMaxNumberOfNodesInConfigs(cfg, uint32(newNumNodes), 0, numOfShards)
			configs.SetQuickJailRatingConfig(cfg)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)

	_, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(3000))
	walletAddress, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 0, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	// wait node to be jailed
	err = cs.GenerateBlocksUntilEpochIsReached(epochWhenNodeIsJailed)
	require.Nil(t, err)

	decodedBLSKey, _ := hex.DecodeString(blsKeys[0])
	status := staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey)
	require.Equal(t, "jailed", status)

	// do an unjail transaction
	unJailValue, _ := big.NewInt(0).SetString("2500000000000000000", 10)
	txUnJailDataField := fmt.Sprintf("unJail@%s", blsKeys[0])
	txUnJail := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 1, vm.ValidatorSCAddress, unJailValue, txUnJailDataField, staking.GasLimitForStakeOperation)

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	unJailTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnJail, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unJailTx)
	require.Equal(t, transaction.TxStatusSuccess, unJailTx.Status)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	status = staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey)
	require.Equal(t, "staked", status)

	staking.CheckValidatorStatus(t, cs, blsKeys[0], nodeStatusAfterUnJail)

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	staking.CheckValidatorStatus(t, cs, blsKeys[0], "waiting")

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 2)
	require.Nil(t, err)

	staking.CheckValidatorStatus(t, cs, blsKeys[0], "eligible")
}

// Test description
// Add a new node and wait until the node get jailed
// Add a second node to take the place of the jailed node
// UnJail the first node --> should go in queue
// Activate staking v4 step 1 --> node should be unstaked as the queue is cleaned up

// Internal test scenario #2
func TestChainSimulator_FromQueueToAuctionList(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	numOfShards := uint32(3)

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       4,
		MetaChainMinNodes:      4,
		AlterConfigsFunction: func(cfg *config.Configs) {
			configs.SetStakingV4ActivationEpochs(cfg, stakingV4JailUnJailStep1EnableEpoch)
			configs.SetQuickJailRatingConfig(cfg)

			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 1
			configs.SetMaxNumberOfNodesInConfigs(cfg, uint32(newNumNodes), 0, numOfShards)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = cs.GenerateBlocks(30)
	require.Nil(t, err)

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys([][]byte{privateKeys[1]})
	require.Nil(t, err)

	mintValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(6000))
	walletAddress, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 0, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	// wait node to be jailed
	err = cs.GenerateBlocksUntilEpochIsReached(epochWhenNodeIsJailed)
	require.Nil(t, err)

	decodedBLSKey0, _ := hex.DecodeString(blsKeys[0])
	status := staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey0)
	require.Equal(t, "jailed", status)

	// add one more node
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[1], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	decodedBLSKey1, _ := hex.DecodeString(blsKeys[1])
	status = staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey1)
	require.Equal(t, "staked", status)

	// unJail the first node
	unJailValue, _ := big.NewInt(0).SetString("2500000000000000000", 10)
	txUnJailDataField := fmt.Sprintf("unJail@%s", blsKeys[0])
	txUnJail := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 2, vm.ValidatorSCAddress, unJailValue, txUnJailDataField, staking.GasLimitForStakeOperation)

	unJailTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnJail, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unJailTx)
	require.Equal(t, transaction.TxStatusSuccess, unJailTx.Status)

	status = staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey0)
	require.Equal(t, "queued", status)

	err = cs.GenerateBlocksUntilEpochIsReached(stakingV4JailUnJailStep1EnableEpoch)
	require.Nil(t, err)

	status = staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey0)
	require.Equal(t, staking.UnStakedStatus, status)

	staking.CheckValidatorStatus(t, cs, blsKeys[0], string(common.InactiveList))
}

func TestJailNodes(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	numOfShards := uint32(3)

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   true,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              numOfShards,
		GenesisTimestamp:         startTime,
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         4,
		MetaChainMinNodes:        4,
		NumNodesWaitingListMeta:  1,
		NumNodesWaitingListShard: 1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			configs.SetQuickJailRatingConfig(cfg)
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 1
			configs.SetMaxNumberOfNodesInConfigs(cfg, uint32(newNumNodes), 0, numOfShards)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	err = cs.GenerateBlocks(30)
	require.Nil(t, err)

	_, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(6000))
	walletAddress, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	for i := 0; i < 10; i++ {
		err = cs.ForceChangeOfEpoch()
		require.Nil(t, err)
	}

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 0, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(200)
	require.Nil(t, err)

	decodedBLSKey0, _ := hex.DecodeString(blsKeys[0])
	status := staking.GetBLSKeyStatus(t, cs.GetNodeHandler(core.MetachainShardId), decodedBLSKey0)
	require.Equal(t, "jailed", status)
}
