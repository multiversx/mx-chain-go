package stateTrieClose

import (
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

func TestPatriciaMerkleTrie_Close(t *testing.T) {
	numLeavesToAdd := 200
	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	tr, _ := trie.NewTrie(trieStorage, integrationTests.TestMarshalizer, integrationTests.TestHasher, 5)

	for i := 0; i < numLeavesToAdd; i++ {
		_ = tr.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}
	_ = tr.Commit()

	numBefore := runtime.NumGoroutine()
	rootHash, _ := tr.RootHash()
	leavesChannel1, _ := tr.GetAllLeavesOnChannel(rootHash)
	numGoRoutines := runtime.NumGoroutine()
	assert.Equal(t, 1, numGoRoutines-numBefore)

	_, _ = tr.GetAllLeavesOnChannel(rootHash)
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 2, numGoRoutines-numBefore)

	_ = tr.Update([]byte("god"), []byte("puppy"))
	_ = tr.Commit()

	rootHash, _ = tr.RootHash()
	_, _ = tr.GetAllLeavesOnChannel(rootHash)
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 3, numGoRoutines-numBefore)

	_ = tr.Update([]byte("eggod"), []byte("cat"))
	_ = tr.Commit()

	rootHash, _ = tr.RootHash()
	leavesChannel2, _ := tr.GetAllLeavesOnChannel(rootHash)
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 4, numGoRoutines-numBefore)

	for range leavesChannel1 {
	}
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 3, numGoRoutines-numBefore)

	for range leavesChannel2 {
	}
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 2, numGoRoutines-numBefore)

	err := tr.Close()
	assert.Nil(t, err)
	time.Sleep(time.Second)
	numGoRoutines = runtime.NumGoroutine()
	assert.True(t, numGoRoutines-numBefore <= 1)
}
