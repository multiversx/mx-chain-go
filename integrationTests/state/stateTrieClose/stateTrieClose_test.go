package stateTrieClose

import (
	"errors"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
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

func TestTrieStorageManager_Close(t *testing.T) {
	closeCalled := false
	args := trie.NewTrieStorageManagerArgs{
		DB: &testscommon.StorerStub{
			CloseCalled: func() error {
				closeCalled = true
				return nil
			},
		},
		Marshalizer:            &testscommon.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10, 32),
	}

	numBefore := runtime.NumGoroutine()
	ts, _ := trie.NewTrieStorageManager(args)
	numGoRoutines := runtime.NumGoroutine()
	assert.Equal(t, 1, numGoRoutines-numBefore)

	err := ts.Close()
	assert.Nil(t, err)
	time.Sleep(time.Second)
	numAfterClose := runtime.NumGoroutine()
	assert.True(t, numAfterClose-numBefore <= 1)
	assert.True(t, closeCalled)
}

func TestTrieStorageManager_CloseErr(t *testing.T) {
	closeCalled := false
	closeErr := errors.New("close error")
	args := trie.NewTrieStorageManagerArgs{
		DB: &testscommon.StorerStub{
			CloseCalled: func() error {
				closeCalled = true
				return closeErr
			},
		},
		Marshalizer:            &testscommon.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10, 32),
	}
	numBefore := runtime.NumGoroutine()
	ts, _ := trie.NewTrieStorageManager(args)
	numGoRoutines := runtime.NumGoroutine()
	assert.Equal(t, 1, numGoRoutines-numBefore)

	err := ts.Close()
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), closeErr.Error()))
	time.Sleep(time.Second)
	numAfterClose := runtime.NumGoroutine()
	assert.True(t, numAfterClose-numBefore <= 1)
	assert.True(t, closeCalled)
}
