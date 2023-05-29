package trie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrieNodesHandler(t *testing.T) {
	t.Parallel()

	handler := newTrieNodesHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.existingNodes)
	assert.NotNil(t, handler.missingHashes)
	assert.NotNil(t, handler.hashesOrder)
	assert.Empty(t, handler.existingNodes)
	assert.Empty(t, handler.missingHashes)
	assert.Empty(t, handler.hashesOrder)
}

func TestTrieNodesHandler_addInitialRootHash(t *testing.T) {
	t.Parallel()

	roothash := "roothash"
	handler := newTrieNodesHandler()
	handler.addInitialRootHash(roothash)
	assert.True(t, handler.hashIsMissing(roothash))
	assert.Equal(t, []string{roothash}, handler.hashesOrder)
}

func TestTrieNodesHandler_processMissingHashWasFound(t *testing.T) {
	t.Parallel()

	roothash := "roothash"
	handler := newTrieNodesHandler()
	handler.addInitialRootHash(roothash)

	n := &leafNode{}
	handler.processMissingHashWasFound(n, roothash)
	_, exists := handler.missingHashes[roothash]
	assert.False(t, exists)

	recoveredNode, exists := handler.getExistingNode(roothash)
	assert.True(t, exists)
	assert.True(t, n == recoveredNode) // pointer testing
}

func TestTrieNodesHandler_jobDone(t *testing.T) {
	t.Parallel()

	roothash := "roothash"
	handler := newTrieNodesHandler()
	assert.True(t, handler.jobDone())

	handler.addInitialRootHash(roothash)
	assert.False(t, handler.jobDone())

	handler.processMissingHashWasFound(&leafNode{}, roothash)
	assert.False(t, handler.jobDone())

	handler.replaceParentWithChildren(0, roothash, make([]node, 0), make([][]byte, 0))
	assert.True(t, handler.jobDone())
}

func TestTrieNodesHandler_noMissingHashes(t *testing.T) {
	t.Parallel()

	roothash := "roothash"
	handler := newTrieNodesHandler()
	assert.True(t, handler.noMissingHashes())

	handler.addInitialRootHash(roothash)
	assert.False(t, handler.noMissingHashes())

	handler.processMissingHashWasFound(&leafNode{}, roothash)
	assert.True(t, handler.noMissingHashes())
}

func TestTrieNodesHandler_replaceParentWithChildren(t *testing.T) {
	t.Parallel()

	roothash := "roothash"
	hash1 := "hash1"
	node1 := &leafNode{
		baseNode: &baseNode{},
	}
	node1.setGivenHash([]byte(hash1))

	hash2 := "hash2"
	node2 := &leafNode{
		baseNode: &baseNode{},
	}
	node2.setGivenHash([]byte(hash2))

	hash3 := "hash3"

	handler := newTrieNodesHandler()
	assert.True(t, handler.jobDone())

	handler.addInitialRootHash(roothash)
	handler.processMissingHashWasFound(&leafNode{}, roothash)

	handler.replaceParentWithChildren(0, roothash, []node{node1, node2}, [][]byte{[]byte(hash3)})

	t.Run("test the initial roothash is deleted", func(t *testing.T) {
		assert.False(t, handler.hashIsMissing(roothash))
	})
	t.Run("test that the 2 existing nodes are added", func(t *testing.T) {
		recoveredNode, exists := handler.getExistingNode(hash1)
		assert.True(t, exists)
		assert.True(t, recoveredNode == node1) // pointer testing

		recoveredNode, exists = handler.getExistingNode(hash2)
		assert.True(t, exists)
		assert.True(t, recoveredNode == node2) // pointer testing

		assert.Equal(t, 2, len(handler.existingNodes))
	})
	t.Run("test that the missing node is added", func(t *testing.T) {
		assert.True(t, handler.hashIsMissing(hash3))

		assert.Equal(t, 1, len(handler.missingHashes))
	})
	t.Run("test the order position", func(t *testing.T) {
		expectedOrder := []string{hash1, hash2, hash3}
		assert.Equal(t, expectedOrder, handler.hashesOrder)
	})
}

func TestReplaceHashesAtPosition(t *testing.T) {
	t.Parallel()

	empty := make([]string, 0)
	newData := []string{"aaa", "bbb"}

	t.Run("empty initial", func(t *testing.T) {
		result := replaceHashesAtPosition(0, make([]string, 0), make([]string, 0))
		assert.Empty(t, result)
	})
	t.Run("empty initial with index out of bound", func(t *testing.T) {
		result := replaceHashesAtPosition(1, make([]string, 0), make([]string, 0))
		assert.Empty(t, result)
	})
	t.Run("empty initial provided data", func(t *testing.T) {
		result := replaceHashesAtPosition(0, make([]string, 0), newData)
		assert.Empty(t, result)
	})

	t.Run("1 existing and replace on first position", func(t *testing.T) {
		initial := []string{"one"}

		result := replaceHashesAtPosition(0, initial, newData)
		expected := []string{"aaa", "bbb"}
		assert.Equal(t, expected, result)
	})
	t.Run("1 existing and replace on an out of bound position", func(t *testing.T) {
		initial := []string{"one"}

		result := replaceHashesAtPosition(1, initial, newData)
		expected := []string{"one"}
		assert.Equal(t, expected, result)
	})
	t.Run("1 existing and replace with empty", func(t *testing.T) {
		initial := []string{"one"}

		result := replaceHashesAtPosition(0, initial, empty)
		assert.Empty(t, result)
	})
	t.Run("2 existing and replace on first position", func(t *testing.T) {
		initialContainingTwo := []string{"one", "two"}

		result := replaceHashesAtPosition(0, initialContainingTwo, newData)
		expected := []string{"aaa", "bbb", "two"}
		assert.Equal(t, expected, result)
	})
	t.Run("2 existing and replace on second position", func(t *testing.T) {
		initial := []string{"one", "two"}

		result := replaceHashesAtPosition(1, initial, newData)
		expected := []string{"one", "aaa", "bbb"}
		assert.Equal(t, expected, result)
	})
	t.Run("2 existing and replace on an out of bound position", func(t *testing.T) {
		initial := []string{"one", "two"}

		result := replaceHashesAtPosition(2, initial, newData)
		expected := []string{"one", "two"}
		assert.Equal(t, expected, result)
	})
	t.Run("2 existing and replace with empty", func(t *testing.T) {
		initial := []string{"one", "two"}

		result := replaceHashesAtPosition(0, initial, empty)
		expected := []string{"two"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace on first position", func(t *testing.T) {
		initial := []string{"one", "two", "three"}

		result := replaceHashesAtPosition(0, initial, newData)
		expected := []string{"aaa", "bbb", "two", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace on second position", func(t *testing.T) {
		initial := []string{"one", "two", "three"}

		result := replaceHashesAtPosition(1, initial, newData)
		expected := []string{"one", "aaa", "bbb", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace on third position", func(t *testing.T) {
		initial := []string{"one", "two", "three"}

		result := replaceHashesAtPosition(2, initial, newData)
		expected := []string{"one", "two", "aaa", "bbb"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace on an out of bound position", func(t *testing.T) {
		initial := []string{"one", "two", "three"}

		result := replaceHashesAtPosition(3, initial, newData)
		expected := []string{"one", "two", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace with empty", func(t *testing.T) {
		initial := []string{"one", "two", "three"}

		result := replaceHashesAtPosition(0, initial, empty)
		expected := []string{"two", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace with empty on second position", func(t *testing.T) {
		initial := []string{"one", "two", "three"}

		result := replaceHashesAtPosition(1, initial, empty)
		expected := []string{"one", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace with empty on third position", func(t *testing.T) {
		initial := []string{"one", "two", "three"}

		result := replaceHashesAtPosition(2, initial, empty)
		expected := []string{"one", "two"}
		assert.Equal(t, expected, result)
	})
}
