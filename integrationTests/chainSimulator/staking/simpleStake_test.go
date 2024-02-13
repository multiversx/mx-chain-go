package staking

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

// Test scenarios
// Do 3 stake transactions from 3 different wallets - tx value 2499, 2500, 2501
// testcase1 -- staking v3.5 --> tx1 fail, tx2 - node in queue, tx3 - node in queue with topUp 1
// testcase2 -- staking v4 step1 --> tx1 fail, tx2 - node in auction, tx3 - node in auction with topUp 1
// testcase3 -- staking v4 step2 --> tx1 fail, tx2 - node in auction, tx3 - node in auction with topUp 1
// testcase4 -- staking v3.step3 --> tx1 fail, tx2 - node in auction, tx3 - node in auction with topUp 1

// // Internal test scenario #3
func TestChainSimulator_SimpleStake(t *testing.T) {
	t.Run("staking ph 4 is not active", func(t *testing.T) {
		testChainSimulatorSimpleStake(t, 1, "queued")
	})

	t.Run("staking ph 4 step1", func(t *testing.T) {
		testChainSimulatorSimpleStake(t, 2, "auction")
	})

	t.Run("staking ph 4 step2", func(t *testing.T) {
		testChainSimulatorSimpleStake(t, 3, "auction")
	})

	t.Run("staking ph 4 step3", func(t *testing.T) {
		testChainSimulatorSimpleStake(t, 4, "auction")
	})
}

func testChainSimulatorSimpleStake(t *testing.T, targetEpoch int32, nodesStatus string) {
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
		BypassTxSignatureCheck:   false,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              numOfShards,
		GenesisTimestamp:         startTime,
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  3,
		NumNodesWaitingListShard: 3,
		AlterConfigsFunction: func(cfg *config.Configs) {
			configs.SetStakingV4ActivationEpochs(cfg, 2)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	mintValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(3000))
	wallet1, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)
	wallet2, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)
	wallet3, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)

	_, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(3)
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	dataFieldTx1 := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], mockBLSSignature)
	tx1Value := big.NewInt(0).Mul(big.NewInt(2499), oneEGLD)
	tx1 := generateTransaction(wallet1.Bytes, 0, vm.ValidatorSCAddress, tx1Value, dataFieldTx1, gasLimitForStakeOperation)

	dataFieldTx2 := fmt.Sprintf("stake@01@%s@%s", blsKeys[1], mockBLSSignature)
	tx2 := generateTransaction(wallet3.Bytes, 0, vm.ValidatorSCAddress, minimumStakeValue, dataFieldTx2, gasLimitForStakeOperation)

	dataFieldTx3 := fmt.Sprintf("stake@01@%s@%s", blsKeys[2], mockBLSSignature)
	tx3Value := big.NewInt(0).Mul(big.NewInt(2501), oneEGLD)
	tx3 := generateTransaction(wallet2.Bytes, 0, vm.ValidatorSCAddress, tx3Value, dataFieldTx3, gasLimitForStakeOperation)

	results, err := cs.SendTxsAndGenerateBlockTilTxIsExecuted([]*transaction.Transaction{tx1, tx2, tx3}, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 3, len(results))
	require.NotNil(t, results)

	// tx1 should fail
	require.Equal(t, "insufficient stake value: expected 2500000000000000000000, got 2499000000000000000000", string(results[0].Logs.Events[0].Topics[1]))

	_ = cs.GenerateBlocks(1)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	if targetEpoch < 2 {
		bls1, _ := hex.DecodeString(blsKeys[1])
		bls2, _ := hex.DecodeString(blsKeys[2])

		blsKeyStatus := getBLSKeyStatus(t, metachainNode, bls1)
		require.Equal(t, nodesStatus, blsKeyStatus)

		blsKeyStatus = getBLSKeyStatus(t, metachainNode, bls2)
		require.Equal(t, nodesStatus, blsKeyStatus)
	} else {
		// tx2 -- validator should be in queue
		checkValidatorStatus(t, cs, blsKeys[1], nodesStatus)
		// tx3 -- validator should be in queue
		checkValidatorStatus(t, cs, blsKeys[2], nodesStatus)
	}
}
