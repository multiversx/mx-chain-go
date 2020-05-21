package sharding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------- CrossShardValidatorDistributor --------------------------------
func TestCrossShardValidatorDistributor_DistributeValidators_NilDestination(t *testing.T) {
	vd := &CrossShardValidatorDistributor{}

	nbWaiting := 3
	nbShards := uint32(2)
	random := []byte{0}

	waitingMap := generateValidatorMap(nbWaiting, nbShards)

	err := vd.DistributeValidators(nil, waitingMap, random)
	assert.Equal(t, ErrNilOrEmptyDestinationForDistribute, err)
}

func TestCrossShardValidatorDistributor_DistributeValidators_DistributesEqually(t *testing.T) {
	vd := &CrossShardValidatorDistributor{}

	nbEligible := 100
	nbWaiting := 30
	nbShards := uint32(2)
	random := []byte{0}

	eligibleMap := generateValidatorMap(nbEligible, nbShards)
	waitingMap := generateValidatorMap(nbWaiting, nbShards)

	err := vd.DistributeValidators(eligibleMap, waitingMap, random)
	assert.Nil(t, err)

	totalNumberAfterDistribution := nbEligible + nbWaiting
	for _, shardValidators := range eligibleMap {
		assert.Equal(t, totalNumberAfterDistribution, len(shardValidators))
	}
}

func TestCrossShardValidatorDistributor_DistributeValidators_ShufflesBetweenShards(t *testing.T) {
	vd := &CrossShardValidatorDistributor{}

	nbEligible := 100
	nbWaiting := 30
	nbShards := uint32(2)
	random := generateRandomByteArray(32)

	eligibleMap := generateValidatorMap(nbEligible, nbShards)
	waitingMap := generateValidatorMap(nbWaiting, nbShards)

	waitingCopy := copyValidatorMap(waitingMap)

	err := vd.DistributeValidators(eligibleMap, waitingMap, random)
	assert.Nil(t, err)

	differentShardCounter := 0
	for shardId, initialWaiting := range waitingCopy {
		for _, waitingValidator := range initialWaiting {
			found, newShardId := searchInMap(eligibleMap, waitingValidator.PubKey())
			assert.True(t, found)
			if newShardId != shardId {
				differentShardCounter++
			}
		}
	}

	assert.True(t, differentShardCounter > 0)
}

func TestCrossShardValidatorDistributor_IsInterfaceNil(t *testing.T) {
	vd := &CrossShardValidatorDistributor{}

	assert.False(t, vd.IsInterfaceNil())
}

// ---------------------- IntraShardValidatorDistributor --------------------------------

func TestIntraShardValidatorDistributor_DistributeValidators_NilDestination(t *testing.T) {
	vd := &IntraShardValidatorDistributor{}

	nbWaiting := 3
	nbShards := uint32(2)
	random := []byte{0}

	waitingMap := generateValidatorMap(nbWaiting, nbShards)

	err := vd.DistributeValidators(nil, waitingMap, random)
	assert.Equal(t, ErrNilOrEmptyDestinationForDistribute, err)
}

func TestIntraShardValidatorDistributor_DistributeValidators_DistributesEqually(t *testing.T) {
	vd := &IntraShardValidatorDistributor{}

	nbEligible := 10
	nbWaiting := 3
	nbShards := uint32(2)
	random := []byte{0}

	eligibleMap := generateValidatorMap(nbEligible, nbShards)
	waitingMap := generateValidatorMap(nbWaiting, nbShards)

	err := vd.DistributeValidators(eligibleMap, waitingMap, random)
	assert.Nil(t, err)

	totalNumberAfterDistribution := nbEligible + nbWaiting
	for _, shardValidators := range eligibleMap {
		assert.Equal(t, totalNumberAfterDistribution, len(shardValidators))
	}
}

func TestIntraShardValidatorDistributor_DistributeValidators_ShufflesIntraShard(t *testing.T) {
	vd := &IntraShardValidatorDistributor{}

	nbEligible := 100
	nbWaiting := 30
	nbShards := uint32(2)
	random := generateRandomByteArray(32)

	eligibleMap := generateValidatorMap(nbEligible, nbShards)
	waitingMap := generateValidatorMap(nbWaiting, nbShards)

	waitingCopy := copyValidatorMap(waitingMap)

	err := vd.DistributeValidators(eligibleMap, waitingMap, random)
	assert.Nil(t, err)

	differentShardCounter := 0
	for shardId, initialWaiting := range waitingCopy {
		for _, waitingValidator := range initialWaiting {
			found, newShardId := searchInMap(eligibleMap, waitingValidator.PubKey())
			assert.True(t, found)
			if newShardId != shardId {
				differentShardCounter++
			}
		}
	}

	assert.True(t, differentShardCounter == 0)
}

func TestIntraShardValidatorDistributor_IsInterfaceNil(t *testing.T) {
	vd := &IntraShardValidatorDistributor{}

	assert.False(t, vd.IsInterfaceNil())
}
