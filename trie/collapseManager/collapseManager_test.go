package collapseManager

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewCollapseManager(t *testing.T) {
	t.Parallel()

	t.Run("invalid maxSizeInMem should error", func(t *testing.T) {
		t.Parallel()

		cm, err := NewCollapseManager(0)
		assert.Nil(t, cm)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cm, err := NewCollapseManager(2 * minSizeInMemory)
		assert.False(t, check.IfNil(cm))
		assert.Nil(t, err)
		assert.Equal(t, 0, cm.sizeInMemory)
		assert.Equal(t, 2*minSizeInMemory, int(cm.maxSizeInMem))
		assert.Equal(t, 0, len(cm.accessedKeys))
		assert.Equal(t, 0, cm.orderAccess.Len())
	})
}

func TestCollapseManager_MarkKeyAsAccessed(t *testing.T) {
	t.Parallel()

	cm, _ := NewCollapseManager(2 * minSizeInMemory)
	cm.MarkKeyAsAccessed([]byte("key1"), 500)
	assert.Equal(t, 500, cm.sizeInMemory)
	assert.Equal(t, 1, len(cm.accessedKeys))
	assert.Equal(t, 1, cm.orderAccess.Len())

	cm.MarkKeyAsAccessed([]byte("key2"), 600)
	assert.Equal(t, 1100, cm.sizeInMemory)
	assert.Equal(t, 2, len(cm.accessedKeys))
	assert.Equal(t, 2, cm.orderAccess.Len())
	oldestVal := cm.orderAccess.Back().Value.([]byte)
	assert.Equal(t, "key1", string(oldestVal))

	cm.MarkKeyAsAccessed([]byte("key1"), 0)
	assert.Equal(t, 1100, cm.sizeInMemory)
	assert.Equal(t, 2, len(cm.accessedKeys))
	assert.Equal(t, 2, cm.orderAccess.Len())
	oldestVal = cm.orderAccess.Back().Value.([]byte)
	assert.Equal(t, "key2", string(oldestVal))
}

func TestCollapseManager_RemoveKey(t *testing.T) {
	t.Parallel()

	cm, _ := NewCollapseManager(2 * minSizeInMemory)
	cm.MarkKeyAsAccessed([]byte("key1"), 500)
	cm.MarkKeyAsAccessed([]byte("key2"), 600)
	assert.Equal(t, 1100, cm.sizeInMemory)
	assert.Equal(t, 2, len(cm.accessedKeys))
	assert.Equal(t, 2, cm.orderAccess.Len())

	cm.RemoveKey([]byte("key1"), -500)
	assert.Equal(t, 600, cm.sizeInMemory)
	assert.Equal(t, 1, len(cm.accessedKeys))
	assert.Equal(t, 1, cm.orderAccess.Len())

	// removing non existing key should do nothing
	cm.RemoveKey([]byte("key3"), 0)
	assert.Equal(t, 600, cm.sizeInMemory)
	assert.Equal(t, 1, len(cm.accessedKeys))
	assert.Equal(t, 1, cm.orderAccess.Len())
}

func TestCollapseManager_AddSizeInMemory(t *testing.T) {
	t.Parallel()

	cm, _ := NewCollapseManager(2 * minSizeInMemory)
	cm.AddSizeInMemory(700)
	assert.Equal(t, 700, cm.GetSizeInMemory())

	cm.AddSizeInMemory(-200)
	assert.Equal(t, 500, cm.GetSizeInMemory())

	cm.AddSizeInMemory(0)
	assert.Equal(t, 500, cm.GetSizeInMemory())

	// sizeInMemory should not go below 0
	cm.AddSizeInMemory(-800)
	assert.Equal(t, 0, cm.GetSizeInMemory())
}

func TestCollapseManager_ShouldCollapseTrie(t *testing.T) {
	t.Parallel()

	cm, _ := NewCollapseManager(minSizeInMemory)
	assert.False(t, cm.ShouldCollapseTrie())

	cm.AddSizeInMemory(minSizeInMemory + 1)
	assert.True(t, cm.ShouldCollapseTrie())
}

func TestCollapseManager_GetCollapsibleLeaves(t *testing.T) {
	t.Parallel()

	t.Run("sizeInMemory below limit should return nil", func(t *testing.T) {
		t.Parallel()

		cm, _ := NewCollapseManager(minSizeInMemory)
		cm.MarkKeyAsAccessed([]byte("key1"), 500)
		leaves, err := cm.GetCollapsibleLeaves()
		assert.Nil(t, err)
		assert.Nil(t, leaves)
	})
	t.Run("should return evicted keys until sizeInMemory is below limit", func(t *testing.T) {
		t.Parallel()

		cm, _ := NewCollapseManager(minSizeInMemory)
		cm.MarkKeyAsAccessed([]byte("key1"), 700)
		cm.MarkKeyAsAccessed([]byte("key2"), 600)
		cm.MarkKeyAsAccessed([]byte("key3"), 500)

		leaves, err := cm.GetCollapsibleLeaves()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(leaves))

		cm.MarkKeyAsAccessed([]byte("key4"), minSizeInMemory-1)
		leaves, err = cm.GetCollapsibleLeaves()
		assert.Nil(t, err)
		assert.Equal(t, 4, len(leaves))
		assert.Equal(t, "key1", string(leaves[0]))
		assert.Equal(t, "key2", string(leaves[1]))
		assert.Equal(t, "key3", string(leaves[2]))
		assert.Equal(t, "key4", string(leaves[3]))
		assert.Equal(t, 0, len(cm.accessedKeys))
		assert.Equal(t, 0, cm.orderAccess.Len())
	})

	t.Run("should return up to numLeavesToCollapseSingleRun evicted keys", func(t *testing.T) {
		t.Parallel()

		cm, _ := NewCollapseManager(minSizeInMemory)
		for i := 0; i < numLeavesToCollapseSingleRun+2; i++ {
			key := []byte("key" + string(rune(i)))
			cm.MarkKeyAsAccessed(key, 500)
		}
		cm.AddSizeInMemory(minSizeInMemory + 1)

		leaves, err := cm.GetCollapsibleLeaves()
		assert.Nil(t, err)
		assert.Equal(t, numLeavesToCollapseSingleRun, len(leaves))
		for i := 0; i < numLeavesToCollapseSingleRun; i++ {
			expectedKey := "key" + string(rune(i))
			assert.Equal(t, expectedKey, string(leaves[i]))
		}
		assert.Equal(t, 2, len(cm.accessedKeys))
		assert.Equal(t, 2, cm.orderAccess.Len())
	})
}

func TestCollapseManager_CloneWithoutState(t *testing.T) {
	t.Parallel()

	cm, _ := NewCollapseManager(2 * minSizeInMemory)
	cm.MarkKeyAsAccessed([]byte("key1"), 500)
	assert.Equal(t, 500, cm.sizeInMemory)
	assert.Equal(t, 1, len(cm.accessedKeys))
	assert.Equal(t, 1, cm.orderAccess.Len())
	clone := cm.CloneWithoutState()

	assert.False(t, check.IfNil(clone))
	clonedCM := clone.(*collapseManager)
	assert.Equal(t, 2*minSizeInMemory, int(clonedCM.maxSizeInMem))
	assert.Equal(t, 0, len(clonedCM.accessedKeys))
	assert.Equal(t, 0, clonedCM.orderAccess.Len())
	assert.Equal(t, 0, clone.GetSizeInMemory())
}
