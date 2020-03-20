package throttle_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	"github.com/stretchr/testify/assert"
)

const testMaxSizeInBytes = uint32(core.MaxSizeInBytes / 1000) * 900
const testMediumSizeInBytes = uint32(core.MaxSizeInBytes / 1000) * 800
const testHalfSizeInBytes = uint32(core.MaxSizeInBytes / 1000) * 450

func TestNewBlockSizeThrottle_ShouldWork(t *testing.T) {
	bst, err := throttle.NewBlockSizeThrottle()

	assert.Nil(t, err)
	assert.NotNil(t, bst)
}

func TestBlockSizeThrottle_MaxSizeToAdd(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSize := uint32(core.MaxSizeInBytes - 1)
	bst.SetMaxSize(maxSize)
	assert.Equal(t, maxSize, bst.GetMaxSize())
}

func TestBlockSizeThrottle_Add(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSize := uint32(testMaxSizeInBytes)
	bst.SetMaxSize(maxSize)

	round := uint64(2)
	size := uint32(testMediumSizeInBytes)
	bst.Add(round, size)

	assert.Equal(t, maxSize, bst.MaxSizeInLastSizeAdded())
	assert.Equal(t, size, bst.SizeInLastSizeAdded())
	assert.Equal(t, round, bst.RoundInLastSizeAdded())
	assert.False(t, bst.SucceedInLastSizeAdded())
}

func TestBlockSizeThrottle_SucceedInLastSizeAdded(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	round := uint64(2)
	size := uint32(testMediumSizeInBytes)
	bst.Add(round, size)
	assert.False(t, bst.SucceedInLastSizeAdded())

	bst.Succeed(round + 1)
	assert.False(t, bst.SucceedInLastSizeAdded())

	bst.Succeed(round)
	assert.True(t, bst.SucceedInLastSizeAdded())
}

func TestBlockSizeThrottle_SucceedInSizeAdded(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	round := uint64(2)
	size := uint32(testMediumSizeInBytes)
	bst.Add(round, size)
	bst.Add(round+1, size)
	assert.False(t, bst.SucceedInSizeAdded(0))
	assert.False(t, bst.SucceedInSizeAdded(1))

	bst.Succeed(round + 1)
	assert.False(t, bst.SucceedInSizeAdded(0))
	assert.True(t, bst.SucceedInSizeAdded(1))

	bst.Succeed(round)
	assert.True(t, bst.SucceedInSizeAdded(0))
	assert.True(t, bst.SucceedInSizeAdded(1))
}

func TestBlockSizeThrottle_ComputeMaxSizeShouldNotSetMaxSizeWhenStatisticListIsEmpty(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	bst.ComputeMaxSize()
	assert.Equal(t, uint32(core.MaxSizeInBytes), bst.GetMaxSize())
}

func TestBlockSizeThrottle_ComputeMaxSizeShouldSetMaxSizeToMaxSizeInBytesWhenLastActionSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	lastActionMaxSize := uint32(core.MaxSizeInBytes)
	bst.SetMaxSize(lastActionMaxSize)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, true)
	bst.ComputeMaxSize()

	assert.Equal(t, uint32(core.MaxSizeInBytes), bst.GetMaxSize())
}

func TestBlockSizeThrottle_ComputeMaxSizeShouldSetMaxSizeToAnIncreasedValueWhenLastActionSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	lastActionMaxSize1 := uint32(core.MaxSizeInBytes)
	bst.SetMaxSize(lastActionMaxSize1)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, true)
	bst.ComputeMaxSize()
	increasedValue := lastActionMaxSize1 + uint32(float32(core.MaxSizeInBytes-lastActionMaxSize1)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.GetMaxSize())

	bst.SetSucceed(2, false)
	lastActionMaxSize2 := uint32(testMediumSizeInBytes)
	bst.SetMaxSize(lastActionMaxSize2)
	bst.Add(3, core.MinSizeInBytes)
	bst.SetSucceed(3, true)
	bst.ComputeMaxSize()
	increasedValue = lastActionMaxSize2 + uint32(float32(lastActionMaxSize1-lastActionMaxSize2)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.GetMaxSize())

	bst.SetSucceed(2, true)
	bst.ComputeMaxSize()
	increasedValue = lastActionMaxSize2 + uint32(float32(core.MaxSizeInBytes-lastActionMaxSize2)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.GetMaxSize())
}

func TestBlockSizeThrottle_ComputeMaxSizeShouldSetMaxSizeToMinSizeInBlockWhenLastActionNotSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	lastActionMaxSize := uint32(core.MinSizeInBytes)
	bst.SetMaxSize(lastActionMaxSize)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, false)
	bst.ComputeMaxSize()

	assert.Equal(t, uint32(core.MinSizeInBytes), bst.GetMaxSize())
}

func TestBlockSizeThrottle_ComputeMaxSizeShouldSetMaxSizeToADecreasedValueWhenLastActionNotSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	lastActionMaxSize1 := core.MaxUint32(testMediumSizeInBytes, core.MinSizeInBytes)
	bst.SetMaxSize(lastActionMaxSize1)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, false)
	bst.ComputeMaxSize()
	decreasedValue := lastActionMaxSize1 - uint32(float32(lastActionMaxSize1-core.MinSizeInBytes)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.GetMaxSize())

	bst.SetSucceed(2, true)
	lastActionMaxSize2 := core.MaxUint32(testMaxSizeInBytes, core.MinSizeInBytes)
	bst.SetMaxSize(lastActionMaxSize2)
	bst.Add(3, core.MinSizeInBytes)
	bst.SetSucceed(3, false)
	bst.ComputeMaxSize()
	decreasedValue = lastActionMaxSize2 - uint32(float32(lastActionMaxSize2-lastActionMaxSize1)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.GetMaxSize())

	bst.SetSucceed(2, false)
	bst.ComputeMaxSize()
	decreasedValue = lastActionMaxSize2 - uint32(float32(lastActionMaxSize2-core.MinSizeInBytes)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.GetMaxSize())
}

func TestBlockSizeThrottle_GetMaxSizeWhenSucceedShouldReturnMaxSizeInBytes(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeWhenSucceed := bst.GetMaxSizeWhenSucceed(core.MaxSizeInBytes)
	assert.Equal(t, uint32(core.MaxSizeInBytes), maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenSucceedShouldReturnMaxSizeUsedWithoutSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeUsedWithoutSucceed := uint32(testMaxSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, core.MinSizeInBytes)
	jumpAbovePercent := float32(bst.JumpAbovePercent()+1) / float32(100)
	maxSizeWhenSucceed := bst.GetMaxSizeWhenSucceed(uint32(float32(maxSizeUsedWithoutSucceed) * jumpAbovePercent))

	assert.Equal(t, maxSizeUsedWithoutSucceed, maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenSucceedShouldIncreaseMaxSizeWithAtLeastOneUnit(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeUsedWithoutSucceed := core.MinUint32(core.MinSizeInBytes+1, core.MaxSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, core.MinSizeInBytes)
	maxSizeWhenSucceed := bst.GetMaxSizeWhenSucceed(core.MinSizeInBytes)

	assert.Equal(t, core.MinUint32(core.MinSizeInBytes+1, core.MaxSizeInBytes), maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenSucceedShouldIncreaseMaxSize(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeUsedWithoutSucceed := uint32(testMaxSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, core.MinSizeInBytes)
	lastActionMaxSize := uint32(float32(maxSizeUsedWithoutSucceed) * 0.5)
	maxSizeWhenSucceed := bst.GetMaxSizeWhenSucceed(lastActionMaxSize)
	increasedValue := lastActionMaxSize + uint32(float32(maxSizeUsedWithoutSucceed-lastActionMaxSize)*0.5)

	assert.Equal(t, increasedValue, maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetCloserAboveMaxSizeUsedWithoutSucceedShouldReturnMaxSizeInBytes(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	assert.Equal(t, uint32(core.MaxSizeInBytes), bst.GetCloserAboveMaxSizeUsedWithoutSucceed(testMaxSizeInBytes - 1))

	maxSizeUsedWithoutSucceed := uint32(testMaxSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, core.MinSizeInBytes)
	assert.Equal(t, uint32(core.MaxSizeInBytes), bst.GetCloserAboveMaxSizeUsedWithoutSucceed(testMaxSizeInBytes))

	bst.SetSucceed(2, true)
	assert.Equal(t, uint32(core.MaxSizeInBytes), bst.GetCloserAboveMaxSizeUsedWithoutSucceed(testMaxSizeInBytes - 1))
}

func TestBlockSizeThrottle_GetCloserAboveMaxSizeUsedWithoutSucceedShouldReturnMaxSizeUsedWithoutSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeUsedWithoutSucceed := uint32(testMaxSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, core.MinSizeInBytes)
	assert.Equal(t, maxSizeUsedWithoutSucceed, bst.GetCloserAboveMaxSizeUsedWithoutSucceed(testMaxSizeInBytes - 1))
}

func TestBlockSizeThrottle_GetMaxSizeWhenNotSucceedShouldReturnMinSizeInBytes(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeWhenSucceed := bst.GetMaxSizeWhenNotSucceed(core.MinSizeInBytes)
	assert.Equal(t, uint32(core.MinSizeInBytes), maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenNotSucceedShouldReturnMaxSizeUsedWithSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeUsedWithSucceed := uint32(testMaxSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, true)
	jumpBelowPercent := float32(bst.JumpBelowPercent()+1) / float32(100)
	maxSizeWhenNotSucceed := bst.GetMaxSizeWhenNotSucceed(uint32(float32(maxSizeUsedWithSucceed) / jumpBelowPercent))

	assert.Equal(t, maxSizeUsedWithSucceed, maxSizeWhenNotSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenNotSucceedShouldDecreaseMaxSizeWithAtLeastOneUnit(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeUsedWithSucceed := uint32(core.MinSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, true)
	maxSizeWhenNotSucceed := bst.GetMaxSizeWhenNotSucceed(core.MinSizeInBytes + 1)

	assert.Equal(t, uint32(core.MinSizeInBytes), maxSizeWhenNotSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenNotSucceedShouldDecreaseMaxSize(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeUsedWithSucceed := core.MaxUint32(testHalfSizeInBytes , core.MinSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, true)
	lastActionMaxSize := uint32(float32(maxSizeUsedWithSucceed) / 0.5)
	maxSizeWhenNotSucceed := bst.GetMaxSizeWhenNotSucceed(lastActionMaxSize)
	decreasedValue := lastActionMaxSize - uint32(float32(lastActionMaxSize-maxSizeUsedWithSucceed)*0.5)

	assert.Equal(t, decreasedValue, maxSizeWhenNotSucceed)
}

func TestBlockSizeThrottle_GetCloserBelowMaxSizeUsedWithSucceedShouldReturnMinSizeInBytes(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	assert.Equal(t, uint32(core.MinSizeInBytes), bst.GetCloserBelowMaxSizeUsedWithSucceed(testMaxSizeInBytes + 1))

	maxSizeUsedWithSucceed := uint32(testMaxSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, true)
	assert.Equal(t, uint32(core.MinSizeInBytes), bst.GetCloserBelowMaxSizeUsedWithSucceed(testMaxSizeInBytes))

	bst.SetSucceed(2, false)
	assert.Equal(t, uint32(core.MinSizeInBytes), bst.GetCloserBelowMaxSizeUsedWithSucceed(testMaxSizeInBytes + 1))
}

func TestBlockSizeThrottle_GetCloserBelowMaxSizeUsedWithSucceedShouldReturnMaxSizeUsedWithoutSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxSizeUsedWithSucceed := uint32(testMaxSizeInBytes)
	bst.SetMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, core.MinSizeInBytes)
	bst.SetSucceed(2, true)
	assert.Equal(t, maxSizeUsedWithSucceed, bst.GetCloserBelowMaxSizeUsedWithSucceed(testMaxSizeInBytes + 1))
}
