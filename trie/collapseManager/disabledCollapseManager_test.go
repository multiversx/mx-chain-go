package collapseManager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDisabledCollapseManager(t *testing.T) {
	t.Parallel()

	dcm := NewDisabledCollapseManager()
	assert.False(t, dcm.IsInterfaceNil())
}

func TestDisabledCollapseManager_MarkKeyAsAccessed(t *testing.T) {
	t.Parallel()

	dcm := NewDisabledCollapseManager()
	dcm.MarkKeyAsAccessed([]byte("key1"), 500)

	assert.Equal(t, 0, dcm.GetSizeInMemory())
}

func TestDisabledCollapseManager_RemoveKey(t *testing.T) {
	t.Parallel()

	dcm := NewDisabledCollapseManager()
	dcm.RemoveKey([]byte("key1"), 500)

	assert.Equal(t, 0, dcm.GetSizeInMemory())
}

func TestDisabledCollapseManager_ShouldCollapseTrie(t *testing.T) {
	t.Parallel()

	dcm := NewDisabledCollapseManager()
	shouldCollapse := dcm.ShouldCollapseTrie()

	assert.False(t, shouldCollapse)
}

func TestDisabledCollapseManager_GetCollapsibleLeaves(t *testing.T) {
	t.Parallel()

	dcm := NewDisabledCollapseManager()
	leaves, err := dcm.GetCollapsibleLeaves()

	assert.Nil(t, err)
	assert.Nil(t, leaves)
}

func TestDisabledCollapseManager_AddSizeInMemory(t *testing.T) {
	t.Parallel()

	dcm := NewDisabledCollapseManager()
	dcm.AddSizeInMemory(500)
	assert.Equal(t, 0, dcm.GetSizeInMemory())
}

func TestDisabledCollapseManager_CloneWithoutState(t *testing.T) {
	t.Parallel()

	dcm := NewDisabledCollapseManager()
	clone := dcm.CloneWithoutState()

	assert.False(t, clone.IsInterfaceNil())
	assert.False(t, clone.IsCollapseEnabled())
}

func TestDisabledCollapseManager_IsCollapseEnabled(t *testing.T) {
	t.Parallel()

	dcm := NewDisabledCollapseManager()
	assert.False(t, dcm.IsCollapseEnabled())
}
