package stateTrieClose

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/goroutines"
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

	gc := goroutines.NewGoCounter(goroutines.TestsRelevantGoRoutines)
	idxInitial, _ := gc.Snapshot()
	rootHash, _ := tr.RootHash()
	leavesChannel1, _ := tr.GetAllLeavesOnChannel(rootHash)
	idx, _ := gc.Snapshot()
	diff := gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 1, len(diff), fmt.Sprintf("%v", diff))

	_, _ = tr.GetAllLeavesOnChannel(rootHash)
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 2, len(diff), fmt.Sprintf("%v", diff))

	_ = tr.Update([]byte("god"), []byte("puppy"))
	_ = tr.Commit()

	rootHash, _ = tr.RootHash()
	_, _ = tr.GetAllLeavesOnChannel(rootHash)
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 3, len(diff), fmt.Sprintf("%v", diff))

	_ = tr.Update([]byte("eggod"), []byte("cat"))
	_ = tr.Commit()

	rootHash, _ = tr.RootHash()
	leavesChannel2, _ := tr.GetAllLeavesOnChannel(rootHash)
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 4, len(diff), fmt.Sprintf("%v", diff))

	for range leavesChannel1 {
	}
	time.Sleep(time.Second) //wait for go routine to finish
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 3, len(diff), fmt.Sprintf("%v", diff))

	for range leavesChannel2 {
	}
	time.Sleep(time.Second) //wait for go routine to finish
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 2, len(diff), fmt.Sprintf("%v", diff))

	err := tr.Close()
	assert.Nil(t, err)
	time.Sleep(time.Second)
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 0, len(diff), fmt.Sprintf("%v", diff))
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
		Hasher:                 &testscommon.HasherMock{},
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10, 32),
	}

	gc := goroutines.NewGoCounter(goroutines.TestsRelevantGoRoutines)
	idxInitial, _ := gc.Snapshot()
	ts, _ := trie.NewTrieStorageManager(args)
	idx, _ := gc.Snapshot()
	diff := gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 1, len(diff), fmt.Sprintf("%v", diff))

	err := ts.Close()
	assert.Nil(t, err)
	time.Sleep(time.Second)
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 0, len(diff), fmt.Sprintf("%v", diff))
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
		Hasher:                 &testscommon.HasherMock{},
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10, 32),
	}
	gc := goroutines.NewGoCounter(goroutines.TestsRelevantGoRoutines)
	idxInitial, _ := gc.Snapshot()
	ts, _ := trie.NewTrieStorageManager(args)
	idx, _ := gc.Snapshot()
	diff := gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 1, len(diff), fmt.Sprintf("%v", diff))

	err := ts.Close()
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), closeErr.Error()))
	time.Sleep(time.Second)
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 0, len(diff), fmt.Sprintf("%v", diff))
	assert.True(t, closeCalled)
}
