package state

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStateChangesCollector(t *testing.T) {
	t.Parallel()

	stateChangesCollector := NewStateChangesCollector()
	require.False(t, stateChangesCollector.IsInterfaceNil())
}

func TestStateChangesCollector_AddStateChange(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(StateChangeDTO{})
	}
	assert.Equal(t, numStateChanges, len(scc.stateChanges))
}

func TestStateChangesCollector_GetStateChanges(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(StateChangeDTO{
			MainTrieKey: []byte(strconv.Itoa(i)),
		})
	}
	assert.Equal(t, numStateChanges, len(scc.stateChanges))

	stateChanges := scc.GetStateChanges()
	assert.Equal(t, numStateChanges, len(stateChanges))
	for i := 0; i < numStateChanges; i++ {
		assert.Equal(t, []byte(strconv.Itoa(i)), stateChanges[i].MainTrieKey)
	}
}

func TestStateChangesCollector_Reset(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(StateChangeDTO{})
	}
	assert.Equal(t, numStateChanges, len(scc.stateChanges))
	stateChanges := scc.GetStateChanges()

	scc.Reset()
	assert.Equal(t, 0, len(scc.stateChanges))
	assert.Equal(t, numStateChanges, len(stateChanges))
}
