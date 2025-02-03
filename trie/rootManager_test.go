package trie

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootManager(t *testing.T) {
	t.Parallel()

	rm := NewRootManager()
	assert.Nil(t, rm.root)
	assert.Empty(t, rm.oldHashes)
	assert.Empty(t, rm.oldRootHash)
}

func TestRootManager_GetRootNode(t *testing.T) {
	t.Parallel()

	bn := &branchNode{
		baseNode: &baseNode{
			hash: []byte{1, 2, 3},
		},
	}
	rm := NewRootManager()
	rm.root = bn
	assert.Equal(t, bn, rm.GetRootNode())
}

func TestRootManager_SetNewRootNode(t *testing.T) {
	t.Parallel()

	bn := &branchNode{
		baseNode: &baseNode{
			hash: []byte{1, 2, 3},
		},
	}
	rm := NewRootManager()
	rm.SetNewRootNode(bn)
	assert.Equal(t, bn, rm.root)
}

func TestRootManager_SetDataForRootChange(t *testing.T) {
	t.Parallel()

	bn := &branchNode{
		baseNode: &baseNode{
			hash: []byte{1, 2, 3},
		},
	}
	oldRootHash := []byte{4, 5, 6}
	oldHashes := [][]byte{{7, 8, 9}, {10, 11, 12}}
	rm := NewRootManager()

	rm.SetDataForRootChange(bn, oldRootHash, oldHashes)
	assert.Equal(t, bn, rm.root)
	assert.Equal(t, oldRootHash, rm.oldRootHash)
	assert.Equal(t, oldHashes, rm.oldHashes)

	var newHash []byte
	rm.SetDataForRootChange(bn, newHash, oldHashes)
	assert.Equal(t, bn, rm.root)
	assert.Equal(t, oldRootHash, rm.oldRootHash)
	assert.Equal(t, append(oldHashes, oldHashes...), rm.oldHashes)
}

func TestRootManager_ResetCollectedHashes(t *testing.T) {
	t.Parallel()

	oldRootHash := []byte{4, 5, 6}
	oldHashes := [][]byte{{7, 8, 9}, {10, 11, 12}}
	rm := NewRootManager()
	rm.oldRootHash = oldRootHash
	rm.oldHashes = oldHashes

	rm.ResetCollectedHashes()
	assert.Empty(t, rm.oldRootHash)
	assert.Empty(t, rm.oldHashes)
}

func TestRootManager_GetOldHashes(t *testing.T) {
	t.Parallel()

	oldHashes := [][]byte{{7, 8, 9}, {10, 11, 12}}
	rm := NewRootManager()
	rm.oldHashes = oldHashes
	assert.Equal(t, oldHashes, rm.GetOldHashes())
}

func TestRootManager_GetOldRootHash(t *testing.T) {
	t.Parallel()

	oldRootHash := []byte{4, 5, 6}
	rm := NewRootManager()
	rm.oldRootHash = oldRootHash
	assert.Equal(t, oldRootHash, rm.GetOldRootHash())
}

func TestRootManager_Concurrency(t *testing.T) {
	t.Parallel()

	numOperations := 1000
	numMethods := 6
	rm := NewRootManager()
	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	for i := 0; i < numOperations; i++ {
		go func(index int) {
			defer wg.Done()
			switch index % numMethods {
			case 0:
				rm.GetRootNode()
			case 1:
				rm.SetNewRootNode(&extensionNode{baseNode: &baseNode{dirty: true}})
			case 2:
				rm.SetDataForRootChange(&branchNode{baseNode: &baseNode{dirty: true}}, []byte{1, 2, 3}, [][]byte{{4, 5, 6}})
			case 3:
				rm.ResetCollectedHashes()
			case 4:
				rm.GetOldHashes()
			case 5:
				rm.GetOldRootHash()
			}
		}(i)
	}
	wg.Wait()
}
