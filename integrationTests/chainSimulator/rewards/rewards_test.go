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
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
)

func TestRewardsTxsAfterEquivalentMessages(t *testing.T) {
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

	metaFacadeHandler := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler()

	nodesSetupFile := path.Join(tempDir, "config", "nodesSetup.json")
	validators, err := readValidatorsAndOwners(nodesSetupFile)
	require.Nil(t, err)

	// find block with rewards transactions, in this range we should find the epoch start block
	var metaBlock *apiCore.Block
	found := false
	for nonce := uint64(210); nonce < 235; nonce++ {
		metaBlock, err = metaFacadeHandler.GetBlockByNonce(nonce, apiCore.BlockQueryOptions{
			WithTransactions: true,
		})
		require.Nil(t, err)

		isEpochStart := metaBlock.EpochStartInfo != nil
		if !isEpochStart {
			continue
		}

		found = true
		break
	}
	require.True(t, found)
	require.NotNil(t, metaBlock)

	coordinator := cs.GetNodeHandler(0).GetProcessComponents().NodesCoordinator()

	rewardsPerShard, err := computeRewardsForShards(metaBlock, coordinator, validators)
	require.Nil(t, err)

	for shardID, reward := range rewardsPerShard {
		fmt.Printf("rewards on shard %d: %s\n", shardID, reward.String())
	}

	require.True(t, allValuesEqual(rewardsPerShard))
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
