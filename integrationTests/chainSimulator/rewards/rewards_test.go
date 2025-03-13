package rewards

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	apiCore "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
)

func TestRewardsTxsAfterEquivalentMessages(t *testing.T) {
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
		AlterConfigsFunction: func(cfg *config.Configs) {
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	targetEpoch := 9
	for i := 0; i < targetEpoch; i++ {
		if i == 8 {
			fmt.Println("here")
		}
		err = cs.ForceChangeOfEpoch()
		require.Nil(t, err)
	}

	err = cs.GenerateBlocks(210)
	require.Nil(t, err)

	metaFacadeHandler := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler()

	nodesSetupFile := path.Join(tempDir, "config", "nodesSetup.json")
	validators, err := readValidatorsAndOwners(nodesSetupFile)
	require.Nil(t, err)
	fmt.Println(validators)

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

	nodesCoordinator := cs.GetNodeHandler(0).GetProcessComponents().NodesCoordinator()

	shards := []uint32{0, 1, 2, core.MetachainShardId}

	rewardsPerShard := make(map[uint32]*big.Int)
	for _, shardID := range shards {
		validatorsPerShard, errG := nodesCoordinator.GetAllEligibleValidatorsPublicKeysForShard(8, shardID)
		require.Nil(t, errG)

		for _, validator := range validatorsPerShard {
			owner, ok := validators[hex.EncodeToString([]byte(validator))]
			if !ok {
				continue
			}

			for _, mb := range metaBlock.MiniBlocks {
				if mb.Type != block.RewardsBlock.String() {
					continue
				}

				for _, tx := range mb.Transactions {
					if tx.Receiver != owner {
						continue
					}
					_, ok = rewardsPerShard[shardID]
					if !ok {
						rewardsPerShard[shardID] = big.NewInt(0)
					}

					valueBig, okR := rewardsPerShard[shardID].SetString(tx.Value, 10)
					require.True(t, okR)
					rewardsPerShard[shardID].Add(rewardsPerShard[shardID], valueBig)
				}
			}
		}

	}

	for shardID, reward := range rewardsPerShard {
		fmt.Printf("rewards on shard %d: %s\n", shardID, reward.String())
	}

	require.True(t, allValuesEqual(rewardsPerShard))
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
	var firstValue *big.Int
	firstSet := false

	for _, v := range m {
		if !firstSet {
			firstValue = v
			firstSet = true
		} else if firstValue.Cmp(v) != 0 {
			return false
		}
	}
	return true
}
