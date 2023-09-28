package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedundancyHandler(t *testing.T) {
	t.Parallel()

	redundancy := NewRedundancyHandler()

	assert.NotNil(t, redundancy)
	assert.Zero(t, redundancy.RoundsOfInactivity())
}

func TestCheckMaxRoundsOfInactivity(t *testing.T) {
	t.Parallel()

	t.Run("with negative value should error", func(t *testing.T) {
		t.Parallel()

		err := CheckMaxRoundsOfInactivity(-1)
		assert.ErrorIs(t, err, errInvalidValue)
		assert.Contains(t, err.Error(), "for maxRoundsOfInactivity, minimum 2 (or 0), got -1")
	})
	t.Run("with value 0 should work", func(t *testing.T) {
		t.Parallel()

		err := CheckMaxRoundsOfInactivity(0)
		assert.Nil(t, err)
	})
	t.Run("with value of 1 should error", func(t *testing.T) {
		t.Parallel()

		err := CheckMaxRoundsOfInactivity(1)
		assert.ErrorIs(t, err, errInvalidValue)
		assert.Contains(t, err.Error(), "for maxRoundsOfInactivity, minimum 2 (or 0), got 1")
	})
	t.Run("with positive values should work", func(t *testing.T) {
		t.Parallel()

		for i := 2; i < 10; i++ {
			err := CheckMaxRoundsOfInactivity(i)
			assert.Nil(t, err)
		}
	})
}

func TestRedundancyHandler_IncrementRoundsOfInactivity(t *testing.T) {
	t.Parallel()

	redundancy := NewRedundancyHandler()

	for i := 0; i < 10; i++ {
		assert.Equal(t, i, redundancy.RoundsOfInactivity())
		redundancy.IncrementRoundsOfInactivity()
	}
}

func TestRedundancyHandler_ResetRoundsOfInactivity(t *testing.T) {
	t.Parallel()

	redundancy := NewRedundancyHandler()

	redundancy.IncrementRoundsOfInactivity()
	assert.Equal(t, 1, redundancy.RoundsOfInactivity())

	redundancy.ResetRoundsOfInactivity()
	assert.Equal(t, 0, redundancy.RoundsOfInactivity())

	for i := 0; i < 10; i++ {
		redundancy.IncrementRoundsOfInactivity()
	}
	assert.Equal(t, 10, redundancy.RoundsOfInactivity())

	redundancy.ResetRoundsOfInactivity()
	assert.Equal(t, 0, redundancy.RoundsOfInactivity())
}

func TestRedundancyHandler_IsRedundancyNode(t *testing.T) {
	t.Parallel()

	redundancy := NewRedundancyHandler()
	assert.False(t, redundancy.IsRedundancyNode(0)) // main machine
	assert.True(t, redundancy.IsRedundancyNode(-1)) // invalid setup
	assert.True(t, redundancy.IsRedundancyNode(1))  // invalid setup
	for i := 2; i < 10; i++ {
		assert.True(t, redundancy.IsRedundancyNode(i)) // backup machine
	}
}

func TestRedundancyHandler_IsMainMachineActive(t *testing.T) {
	t.Parallel()

	redundancy := NewRedundancyHandler()
	t.Run("running on the backup machine, the main machine is active", func(t *testing.T) {
		assert.True(t, redundancy.IsMainMachineActive(2))
	})
	t.Run("running on the backup machine, the main machine lost one round", func(t *testing.T) {
		redundancy.IncrementRoundsOfInactivity()
		assert.True(t, redundancy.IsMainMachineActive(2))
	})
	t.Run("running on the backup machine, the main machine lost the second round", func(t *testing.T) {
		redundancy.IncrementRoundsOfInactivity()
		assert.False(t, redundancy.IsMainMachineActive(2))
	})
	t.Run("running on the backup machine, the main machine lost the third round", func(t *testing.T) {
		redundancy.IncrementRoundsOfInactivity()
		assert.False(t, redundancy.IsMainMachineActive(2))
	})
	t.Run("running on the backup machine, the main machine recovered", func(t *testing.T) {
		redundancy.IncrementRoundsOfInactivity()
		redundancy.ResetRoundsOfInactivity()
		assert.True(t, redundancy.IsMainMachineActive(2))
	})
}
