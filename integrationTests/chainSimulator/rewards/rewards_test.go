package rewards

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	apiCore "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	csUtils "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"

	moveBalanceGasLimit = 50_000
	gasPrice            = 1_000_000_000
)

func TestRewardsAfterAndromedaWithTxs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    200,
	}

	numOfShards := uint32(3)

	tempDir := t.TempDir()
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                tempDir,
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       3,
		MetaChainMinNodes:      3,
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	targetEpoch := 9
	for i := 0; i < targetEpoch; i++ {
		err = cs.ForceChangeOfEpoch()
		require.Nil(t, err)
	}

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	targetShardID := uint32(0)
	numTxs := 10_000
	txs := generateMoveBalance(t, cs, numTxs, targetShardID, targetShardID)

	results, err := cs.SendTxsAndGenerateBlocksTilAreExecuted(txs, 10)
	require.Nil(t, err)

	blockWithTxsHash := results[0].BlockHash
	blocksWithTxs, err := cs.GetNodeHandler(targetShardID).GetFacadeHandler().GetBlockByHash(blockWithTxsHash, apiCore.BlockQueryOptions{})
	require.Nil(t, err)

	prevRandSeed, _ := hex.DecodeString(blocksWithTxs.PrevRandSeed)
	leader, _, err := cs.GetNodeHandler(targetShardID).GetProcessComponents().NodesCoordinator().ComputeConsensusGroup(prevRandSeed, blocksWithTxs.Round, 0, blocksWithTxs.Epoch)
	require.Nil(t, err)

	nodesSetupFile := path.Join(tempDir, "config", "nodesSetup.json")
	validators, err := readValidatorsAndOwners(nodesSetupFile)
	require.Nil(t, err)

	err = cs.GenerateBlocks(210)
	require.Nil(t, err)

	metaBlock := getLastStartOfEpochBlock(t, cs, core.MetachainShardId)
	require.NotNil(t, metaBlock)

	leaderEncoded, _ := cs.GetNodeHandler(0).GetCoreComponents().ValidatorPubKeyConverter().Encode(leader.PubKey())
	leaderOwnerBlockWithTxs := validators[leaderEncoded]

	var anotherOwner string
	found := false
	for _, address := range validators {
		if address != leaderOwnerBlockWithTxs {
			anotherOwner = address
			found = true
		}
	}
	require.True(t, found)

	rewardTxForLeader := getRewardTxForAddress(metaBlock, leaderOwnerBlockWithTxs)
	require.NotNil(t, rewardTxForLeader)
	anotherRewardTx := getRewardTxForAddress(metaBlock, anotherOwner)
	require.NotNil(t, anotherRewardTx)

	rewardTxValueLeaderWithTxs, _ := big.NewInt(0).SetString(rewardTxForLeader.Value, 10)
	rewardTxValueAnotherOwner, _ := big.NewInt(0).SetString(anotherRewardTx.Value, 10)

	coordinator := cs.GetNodeHandler(0).GetProcessComponents().NodesCoordinator()

	rewardsPerShard, err := computeRewardsForShards(metaBlock, coordinator, validators)
	require.Nil(t, err)

	// diff should be equal with 0.1 * moveBalanceCost * num transactions
	// diff = 0.1 * move balance gas limit * gas price * num transactions
	diff := big.NewInt(0).Mul(big.NewInt(moveBalanceGasLimit*0.1), big.NewInt(gasPrice))
	diff.Mul(diff, big.NewInt(int64(numTxs)))

	// check reward tx value
	require.Equal(t, rewardTxValueLeaderWithTxs, big.NewInt(0).Add(rewardTxValueAnotherOwner, diff))

	// rewards for target shard should be rewards for another shard + diff
	require.Equal(t, rewardsPerShard[targetShardID], big.NewInt(0).Add(rewardsPerShard[core.MetachainShardId], diff))
}

func getRewardTxForAddress(block *apiCore.Block, address string) *transaction.ApiTransactionResult {
	for _, mb := range block.MiniBlocks {
		for _, tx := range mb.Transactions {
			if tx.Receiver == address {
				return tx
			}
		}
	}

	return nil
}

func generateMoveBalance(t *testing.T, cs chainSimulator.ChainSimulator, numTxs int, senderShardID, receiverShardID uint32) []*transaction.Transaction {
	numSenders := (numTxs + common.MaxTxNonceDeltaAllowed - 1) / common.MaxTxNonceDeltaAllowed
	tenEGLD := big.NewInt(0).Mul(csUtils.OneEGLD, big.NewInt(10))

	senders := make([]dtos.WalletAddress, 0, numSenders)
	for i := 0; i < numSenders; i++ {

		sender, err := cs.GenerateAndMintWalletAddress(senderShardID, tenEGLD)
		require.Nil(t, err)

		senders = append(senders, sender)
	}

	err := cs.GenerateBlocks(1)
	require.Nil(t, err)

	txs := make([]*transaction.Transaction, 0, numTxs)
	for i := 0; i < numSenders-1; i++ {
		txs = append(txs, generateMoveBalanceTxs(t, cs, senders[i], common.MaxTxNonceDeltaAllowed, receiverShardID)...)
	}

	lastBatchSize := numTxs % common.MaxTxNonceDeltaAllowed
	if lastBatchSize == 0 {
		lastBatchSize = common.MaxTxNonceDeltaAllowed
	}

	txs = append(txs, generateMoveBalanceTxs(t, cs, senders[len(senders)-1], lastBatchSize, receiverShardID)...)

	return txs
}

func generateMoveBalanceTxs(t *testing.T, cs chainSimulator.ChainSimulator, sender dtos.WalletAddress, numTxs int, receiverShardID uint32) []*transaction.Transaction {
	senderShardID := cs.GetNodeHandler(core.MetachainShardId).GetShardCoordinator().ComputeId(sender.Bytes)

	res, _, err := cs.GetNodeHandler(senderShardID).GetFacadeHandler().GetAccount(sender.Bech32, apiCore.AccountQueryOptions{})
	require.Nil(t, err)

	txs := make([]*transaction.Transaction, numTxs)
	initialNonce := res.Nonce
	for i := 0; i < numTxs; i++ {
		rcv := cs.GenerateAddressInShard(receiverShardID)

		txs[i] = &transaction.Transaction{
			Nonce:     initialNonce + uint64(i),
			Value:     big.NewInt(1),
			RcvAddr:   rcv.Bytes,
			SndAddr:   sender.Bytes,
			GasPrice:  gasPrice,
			GasLimit:  moveBalanceGasLimit,
			ChainID:   []byte(configs.ChainID),
			Version:   2,
			Signature: []byte("sig"),
		}
	}

	return txs
}

func TestRewardsTxsAfterAndromeda(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    200,
	}

	numOfShards := uint32(3)

	tempDir := t.TempDir()
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                tempDir,
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       3,
		MetaChainMinNodes:      3,
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	targetEpoch := 9
	for i := 0; i < targetEpoch; i++ {
		err = cs.ForceChangeOfEpoch()
		require.Nil(t, err)
	}

	err = cs.GenerateBlocks(210)
	require.Nil(t, err)

	nodesSetupFile := path.Join(tempDir, "config", "nodesSetup.json")
	validators, err := readValidatorsAndOwners(nodesSetupFile)
	require.Nil(t, err)

	metaBlock := getLastStartOfEpochBlock(t, cs, core.MetachainShardId)
	require.NotNil(t, metaBlock)

	coordinator := cs.GetNodeHandler(0).GetProcessComponents().NodesCoordinator()

	rewardsPerShard, err := computeRewardsForShards(metaBlock, coordinator, validators)
	require.Nil(t, err)

	for shardID, reward := range rewardsPerShard {
		fmt.Printf("rewards on shard %d: %s\n", shardID, reward.String())
	}

	require.True(t, allValuesEqual(rewardsPerShard))
}

func getLastStartOfEpochBlock(t *testing.T, cs chainSimulator.ChainSimulator, shardID uint32) *apiCore.Block {
	metachainHandler := cs.GetNodeHandler(shardID).GetFacadeHandler()

	networkStatus, err := metachainHandler.StatusMetrics().NetworkMetrics()
	require.Nil(t, err)

	epochStartBlocKNonce, ok := networkStatus[common.MetricNonceAtEpochStart].(uint64)
	require.True(t, ok)

	metaBlock, err := metachainHandler.GetBlockByNonce(epochStartBlocKNonce, apiCore.BlockQueryOptions{
		WithTransactions: true,
	})
	require.Nil(t, err)

	return metaBlock
}

func computeRewardsForShards(
	metaBlock *apiCore.Block,
	coordinator nodesCoordinator.NodesCoordinator,
	validators map[string]string,
) (map[uint32]*big.Int, error) {
	shards := []uint32{0, 1, 2, core.MetachainShardId}
	rewardsPerShard := make(map[uint32]*big.Int)

	for _, shardID := range shards {
		rewardsPerShard[shardID] = big.NewInt(0) // Initialize reward entry
		err := computeRewardsForShard(metaBlock, coordinator, validators, shardID, rewardsPerShard)
		if err != nil {
			return nil, err
		}
	}

	return rewardsPerShard, nil
}

func computeRewardsForShard(metaBlock *apiCore.Block,
	coordinator nodesCoordinator.NodesCoordinator,
	validators map[string]string,
	shardID uint32,
	rewardsPerShard map[uint32]*big.Int,
) error {
	validatorsPerShard, _ := coordinator.GetAllEligibleValidatorsPublicKeysForShard(8, shardID)

	for _, validator := range validatorsPerShard {
		owner, exists := validators[hex.EncodeToString([]byte(validator))]
		if !exists {
			continue
		}
		err := accumulateShardRewards(metaBlock, shardID, owner, rewardsPerShard)
		if err != nil {
			return err
		}
	}

	return nil
}

func accumulateShardRewards(metaBlock *apiCore.Block, shardID uint32, owner string, rewardsPerShard map[uint32]*big.Int) error {
	var firstValue *big.Int
	for _, mb := range metaBlock.MiniBlocks {
		if mb.Type != block.RewardsBlock.String() {
			continue
		}

		for _, tx := range mb.Transactions {
			if tx.Receiver != owner {
				continue
			}

			valueBig, _ := new(big.Int).SetString(tx.Value, 10)
			if firstValue == nil {
				firstValue = valueBig
			}
			if valueBig.Cmp(firstValue) != 0 {
				return errors.New("different values in rewards transactions")
			}

			rewardsPerShard[shardID].Add(rewardsPerShard[shardID], valueBig)
		}
	}

	return nil
}

func readValidatorsAndOwners(filePath string) (map[string]string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var nodesSetup struct {
		InitialNodes []struct {
			PubKey  string `json:"pubkey"`
			Address string `json:"address"`
		} `json:"initialNodes"`
	}

	err = json.Unmarshal(file, &nodesSetup)
	if err != nil {
		return nil, err
	}

	validators := make(map[string]string)
	for _, node := range nodesSetup.InitialNodes {
		validators[node.PubKey] = node.Address
	}

	return validators, nil
}

func allValuesEqual(m map[uint32]*big.Int) bool {
	if len(m) == 0 {
		return true
	}
	expectedValue := m[0]
	for _, v := range m {
		if expectedValue.Cmp(v) != 0 {
			return false
		}
	}
	return true
}
