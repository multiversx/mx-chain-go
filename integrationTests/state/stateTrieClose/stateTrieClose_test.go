package stateTrieClose

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/goroutines"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/hashesHolder"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
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
	time.Sleep(time.Second * 2) // allow the commit go routines to finish completely as to not alter the further counters

	gc := goroutines.NewGoCounter(goroutines.TestsRelevantGoRoutines)
	idxInitial, _ := gc.Snapshot()
	rootHash, _ := tr.RootHash()
	leavesChannel1 := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	_ = tr.GetAllLeavesOnChannel(leavesChannel1, context.Background(), rootHash, keyBuilder.NewDisabledKeyBuilder())
	time.Sleep(time.Second) // allow the go routine to start
	idx, _ := gc.Snapshot()
	diff := gc.DiffGoRoutines(idxInitial, idx)
	assert.True(t, len(diff) <= 1) // can be 0 on a fast running host
	err := leavesChannel1.ErrChan.ReadFromChanNonBlocking()
	assert.Nil(t, err)

	leavesChannel1 = &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	_ = tr.GetAllLeavesOnChannel(leavesChannel1, context.Background(), rootHash, keyBuilder.NewDisabledKeyBuilder())
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.True(t, len(diff) <= 2)
	err = leavesChannel1.ErrChan.ReadFromChanNonBlocking()
	assert.Nil(t, err)

	_ = tr.Update([]byte("god"), []byte("puppy"))
	_ = tr.Commit()

	rootHash, _ = tr.RootHash()
	leavesChannel1 = &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	_ = tr.GetAllLeavesOnChannel(leavesChannel1, context.Background(), rootHash, keyBuilder.NewDisabledKeyBuilder())
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 3, len(diff), fmt.Sprintf("%v", diff))
	err = leavesChannel1.ErrChan.ReadFromChanNonBlocking()
	assert.Nil(t, err)

	_ = tr.Update([]byte("eggod"), []byte("cat"))
	_ = tr.Commit()

	rootHash, _ = tr.RootHash()
	leavesChannel2 := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	_ = tr.GetAllLeavesOnChannel(leavesChannel2, context.Background(), rootHash, keyBuilder.NewDisabledKeyBuilder())
	time.Sleep(time.Second) // allow the go routine to start
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.True(t, len(diff) <= 4)
	err = leavesChannel2.ErrChan.ReadFromChanNonBlocking()
	assert.Nil(t, err)

	for range leavesChannel1.LeavesChan {
	}
	time.Sleep(time.Second) // wait for go routine to finish
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.True(t, len(diff) <= 3)
	err = leavesChannel1.ErrChan.ReadFromChanNonBlocking()
	assert.Nil(t, err)

	for range leavesChannel2.LeavesChan {
	}
	time.Sleep(time.Second) // wait for go routine to finish
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.True(t, len(diff) <= 2)
	err = leavesChannel2.ErrChan.ReadFromChanNonBlocking()
	assert.Nil(t, err)

	err = tr.Close()
	assert.Nil(t, err)
	time.Sleep(time.Second)
	idx, _ = gc.Snapshot()
	diff = gc.DiffGoRoutines(idxInitial, idx)
	assert.Equal(t, 0, len(diff), fmt.Sprintf("%v", diff))
}

func TestTrieStorageManager_Close(t *testing.T) {
	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            &testscommon.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		GeneralConfig:          config.TrieStorageManagerConfig{SnapshotsGoroutineNum: 1},
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10, 32),
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
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
}
