package nodesCoordinator

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDummyValidatorsInfo() []*state.ShardValidatorInfo {
	return []*state.ShardValidatorInfo{
		{
			PublicKey:  []byte("pubKey1"),
			ShardId:    0,
			List:       "eligible",
			Index:      0,
			TempRating: 10},
		{
			PublicKey:  []byte("pubKey2"),
			ShardId:    0,
			List:       "eligible",
			Index:      1,
			TempRating: 11},
		{
			PublicKey:  []byte("pubKey3"),
			ShardId:    0,
			List:       "eligible",
			Index:      2,
			TempRating: 12},
		{
			PublicKey:  []byte("pubKey4"),
			ShardId:    core.MetachainShardId,
			List:       "eligible",
			Index:      0,
			TempRating: 13},
		{
			PublicKey:  []byte("pubKey5"),
			ShardId:    core.MetachainShardId,
			List:       "eligible",
			Index:      1,
			TempRating: 14},
		{
			PublicKey:  []byte("pubKey6"),
			ShardId:    core.MetachainShardId,
			List:       "eligible",
			Index:      2,
			TempRating: 15},
	}
}

func verifyEligibleValidators(
	t *testing.T,
	validatorsInfo []*state.ShardValidatorInfo,
	eligibleMap map[uint32][]Validator,
) {
	require.NotNil(t, eligibleMap)

	numEligibleValidators := 0
	for _, validatorsInShard := range eligibleMap {
		for _, val := range validatorsInShard {
			found := false
			for _, validatorInfo := range validatorsInfo {
				samePubKey := bytes.Equal(val.PubKey(), validatorInfo.GetPublicKey())
				sameIndex := val.Index() == validatorInfo.GetIndex()
				if samePubKey && sameIndex {
					found = true
				}
			}
			assert.True(t, found)
			numEligibleValidators += 1
		}
	}

	assert.Equal(t, len(validatorsInfo), numEligibleValidators)
}

func TestIndexHashedNodesCoordinator_SetNodesConfigFromValidatorsInfo(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:          3,
		NodesMeta:           3,
		EnableEpochsHandler: &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, _ := NewHashValidatorsShuffler(shufflerArgs)
	arguments.Shuffler = nodeShuffler

	ihnc, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(1)
	validatorsInfo := createDummyValidatorsInfo()
	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch, []byte("rand seed"), validatorsInfo)
	require.Nil(t, err)

	verifyEligibleValidators(t, validatorsInfo, ihnc.nodesConfig[epoch].eligibleMap)
}

func TestIndexHashedNodesCoordinator_SetNodesConfigFromValidatorsInfoMultipleEpochs(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:          3,
		NodesMeta:           3,
		EnableEpochsHandler: &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, _ := NewHashValidatorsShuffler(shufflerArgs)
	arguments.Shuffler = nodeShuffler

	ihnc, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(1)
	randomness := []byte("rand seed")

	validatorsInfo1 := createDummyValidatorsInfo()
	validatorsInfo2 := createDummyValidatorsInfo()

	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch, randomness, validatorsInfo1)
	require.Nil(t, err)

	epochConfig, ok := ihnc.nodesConfig[epoch]
	assert.True(t, ok)
	verifyEligibleValidators(t, validatorsInfo1, epochConfig.eligibleMap)

	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch+1, randomness, validatorsInfo2)
	require.Nil(t, err)

	epochConfig, ok = ihnc.nodesConfig[epoch+1]
	assert.True(t, ok)
	verifyEligibleValidators(t, validatorsInfo2, epochConfig.eligibleMap)

	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch+numStoredEpochs, randomness, validatorsInfo1)
	assert.Nil(t, err)

	_, ok = ihnc.nodesConfig[epoch]
	assert.False(t, ok)

	epochConfig, ok = ihnc.nodesConfig[epoch+1]
	assert.True(t, ok)
	verifyEligibleValidators(t, validatorsInfo2, epochConfig.eligibleMap)

	validators, err := ihnc.GetAllEligibleValidatorsPublicKeys(epoch)
	require.Nil(t, validators)
	require.True(t, errors.Is(err, ErrEpochNodesConfigDoesNotExist))
}

func TestIndexHashedNodesCoordinator_IsEpochInConfig(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.ValidatorInfoCacher = dataPool.NewCurrentEpochValidatorInfoPool()
	ihnc, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(1)
	ihnc.nodesConfig[epoch] = ihnc.nodesConfig[0]

	ihnc.updateEpochFlags(epoch)

	body := createBlockBodyFromNodesCoordinator(ihnc, epoch, ihnc.validatorInfoCacher)
	validatorsInfo, _ := ihnc.createValidatorInfoFromBody(body, 10, epoch)

	err = ihnc.SetNodesConfigFromValidatorsInfo(epoch, []byte{}, validatorsInfo)
	require.Nil(t, err)

	exists := ihnc.IsEpochInConfig(epoch)
	assert.True(t, exists)

	exists = ihnc.IsEpochInConfig(epoch + 1)
	assert.False(t, exists)
}
