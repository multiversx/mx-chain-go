package nodesCoordinator

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDummyValidatorsInfo() []*state.ShardValidatorInfo {
	return []*state.ShardValidatorInfo{
		{
			PublicKey:  []uint8{1, 2, 3, 4},
			ShardId:    0,
			List:       "eligible",
			Index:      0,
			TempRating: 10},
		{
			PublicKey:  []uint8{2, 3, 4, 5},
			ShardId:    0,
			List:       "eligible",
			Index:      1,
			TempRating: 10},
		{
			PublicKey:  []uint8{3, 4, 5, 1},
			ShardId:    0,
			List:       "eligible",
			Index:      2,
			TempRating: 10},
		{
			PublicKey:  []uint8{4, 5, 1, 2},
			ShardId:    core.MetachainShardId,
			List:       "eligible",
			Index:      0,
			TempRating: 10},
		{
			PublicKey:  []uint8{5, 1, 2, 3},
			ShardId:    core.MetachainShardId,
			List:       "eligible",
			Index:      1,
			TempRating: 10},
		{
			PublicKey:  []uint8{1, 2, 3, 5},
			ShardId:    core.MetachainShardId,
			List:       "eligible",
			Index:      2,
			TempRating: 10},
	}
}

func TestIndexHashedNodesCoordinator_SetNodesConfigFromValidatorsInfo(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           3,
		NodesMeta:            3,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, _ := NewHashValidatorsShuffler(shufflerArgs)
	arguments.Shuffler = nodeShuffler

	ihnc, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(1)

	validatorsInfo := createDummyValidatorsInfo()

	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch, []byte("rand seed"), validatorsInfo)
	require.Nil(t, err)

	verifyValidatorsPubKeysByIndex(t, validatorsInfo, ihnc.nodesConfig[epoch])
}

func TestIndexHashedNodesCoordinator_SetNodesConfigFromValidatorsInfoMultipleEpochs(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           3,
		NodesMeta:            3,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, _ := NewHashValidatorsShuffler(shufflerArgs)
	arguments.Shuffler = nodeShuffler

	ihnc, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(1)

	randomness := []byte("rand seed")

	validatorsInfo := createDummyValidatorsInfo()

	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch, randomness, validatorsInfo)
	require.Nil(t, err)
	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch+1, randomness, validatorsInfo)
	require.Nil(t, err)

	_, ok := ihnc.nodesConfig[epoch+1]
	assert.True(t, ok)
	verifyValidatorsPubKeysByIndex(t, validatorsInfo, ihnc.nodesConfig[epoch+1])

	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch+nodesCoordinatorStoredEpochs, randomness, validatorsInfo)
	assert.Nil(t, err)

	_, ok = ihnc.nodesConfig[epoch]
	assert.False(t, ok)

	_, ok = ihnc.nodesConfig[epoch+1]
	assert.True(t, ok)
	verifyValidatorsPubKeysByIndex(t, validatorsInfo, ihnc.nodesConfig[epoch+1])

	validators, err := ihnc.GetAllEligibleValidatorsPublicKeys(epoch)
	require.Nil(t, validators)
	require.True(t, errors.Is(err, ErrEpochNodesConfigDoesNotExist))
}

func verifyValidatorsPubKeysByIndex(
	t *testing.T,
	validatorsInfo []*state.ShardValidatorInfo,
	nodesConfig *epochNodesConfig,
) {
	for i := range validatorsInfo {
		shardId := validatorsInfo[i].ShardId
		for j := range nodesConfig.eligibleMap[shardId] {
			if validatorsInfo[i].Index == nodesConfig.eligibleMap[shardId][j].Index() {
				assert.Equal(t, validatorsInfo[i].PublicKey, nodesConfig.eligibleMap[shardId][j].PubKey())
			}
		}
	}
}

func TestIndexHashedNodesCoordinator_IsEpochInConfig(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihnc, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(1)

	ihnc.nodesConfig[epoch] = ihnc.nodesConfig[0]

	body := createBlockBodyFromNodesCoordinator(ihnc, epoch)
	validatorsInfo, _ := createValidatorInfoFromBody(body, arguments.Marshalizer, 10)

	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch, []byte{}, validatorsInfo)

	exists := ihnc.IsEpochInConfig(epoch)
	assert.True(t, exists)

	exists = ihnc.IsEpochInConfig(epoch + 1)
	assert.False(t, exists)
}
