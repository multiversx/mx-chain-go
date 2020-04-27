package peer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorCounters(t *testing.T) {
	t.Parallel()
	nrOfIncrements := 10
	testKey := []byte("testKey")
	vc := make(validatorRoundCounters)

	for i := 0; i < nrOfIncrements; i++ {
		vc.increaseLeader(testKey)
		vc.decreaseLeader(testKey)
		vc.increaseValidator(testKey)
		vc.decreaseValidator(testKey)
	}

	assert.Equal(t, uint32(nrOfIncrements), vc.get(testKey).leaderIncreaseCount)
	assert.Equal(t, uint32(nrOfIncrements), vc.get(testKey).leaderDecreaseCount)
	assert.Equal(t, uint32(nrOfIncrements), vc.get(testKey).validatorIncreaseCount)
	assert.Equal(t, uint32(nrOfIncrements), vc.get(testKey).validatorDecreaseCount)
}

func TestValidatorCountersReset(t *testing.T) {
	t.Parallel()
	nrOfIncrements := 10
	testKey := []byte("testKey")
	vc := make(validatorRoundCounters)

	for i := 0; i < nrOfIncrements; i++ {
		vc.increaseLeader(testKey)
		vc.decreaseLeader(testKey)
		vc.increaseValidator(testKey)
		vc.decreaseValidator(testKey)
	}

	vc.reset()

	assert.Equal(t, uint32(0), vc.get(testKey).leaderIncreaseCount)
	assert.Equal(t, uint32(0), vc.get(testKey).leaderDecreaseCount)
	assert.Equal(t, uint32(0), vc.get(testKey).validatorIncreaseCount)
	assert.Equal(t, uint32(0), vc.get(testKey).validatorDecreaseCount)
}
