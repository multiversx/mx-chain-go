package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

func TestWriteContinuously(t *testing.T) {
	t.Skip("this is a long test")

	nbTxsWrite := 1000000
	testStorage := integrationTests.NewTestStorage()
	store := testStorage.CreateStorageLevelDB()

	defer func() {
		_ = store.DestroyUnit()
	}()

	startTime := time.Now()
	written := 10000
	for i := 1; i <= nbTxsWrite; i++ {
		if i%written == 0 {
			endTime := time.Now()
			diff := endTime.Sub(startTime)
			fmt.Printf("Written %d, total %d in %f s\n", written, i, diff.Seconds())
			startTime = time.Now()
		}

		key, val := testStorage.CreateStoredData(uint64(i))
		err := store.Put(key, val)

		assert.Nil(t, err)
	}
}

func TestWriteReadDeleteLevelDB(t *testing.T) {
	t.Skip("this is a long test")

	maxWritten := uint64(0)
	mapRemovedKeys := &sync.Map{}

	wg := &sync.WaitGroup{}
	errors := int32(0)
	testStorage := integrationTests.NewTestStorage()
	store := testStorage.CreateStorageLevelDB()
	nbTxsWrite := 1000000
	wg.Add(3)
	chWriteDone := make(chan struct{})

	testStorage.InitAdditionalFieldsForStorageOperations(nbTxsWrite, mapRemovedKeys, &maxWritten)

	defer func() {
		_ = store.DestroyUnit()
	}()

	go testStorage.WriteMultipleWithNotif(store, wg, chWriteDone, 2, &errors)
	go testStorage.RemoveMultiple(store, wg, chWriteDone, &errors)
	go testStorage.ReadMultiple(store, wg, chWriteDone, &errors)
	wg.Wait()

	assert.Equal(t, int32(0), errors)
}

func TestWriteReadDeleteLevelDBSerial(t *testing.T) {
	t.Skip("this is a long test")

	maxWritten := uint64(0)
	mapRemovedKeys := &sync.Map{}

	wg := &sync.WaitGroup{}
	errors := int32(0)

	testStorage := integrationTests.NewTestStorage()
	store := testStorage.CreateStorageLevelDBSerial()
	nbTxsWrite := 1000000

	testStorage.InitAdditionalFieldsForStorageOperations(nbTxsWrite, mapRemovedKeys, &maxWritten)
	wg.Add(3)
	chWriteDone := make(chan struct{})

	defer func() {
		_ = store.DestroyUnit()
	}()

	go testStorage.WriteMultipleWithNotif(store, wg, chWriteDone, 2, &errors)
	go testStorage.RemoveMultiple(store, wg, chWriteDone, &errors)
	go testStorage.ReadMultiple(store, wg, chWriteDone, &errors)
	wg.Wait()

	assert.Equal(t, int32(0), errors)
}

func TestWriteContinuouslyInTree(t *testing.T) {
	t.Skip("this is a long test")

	nbTxsWrite := 1000000
	testStorage := integrationTests.NewTestStorage()
	store := testStorage.CreateStorageLevelDB()

	trieStorage, _ := trie.NewTrieStorageManagerWithoutPruning(store)

	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(trieStorage, &marshal.JsonMarshalizer{}, blake2b.NewBlake2b(), maxTrieLevelInMemory)

	defer func() {
		_ = store.DestroyUnit()
	}()

	startTime := time.Now()
	written := 10000

	for i := 1; i <= nbTxsWrite; i++ {
		if i%written == 0 {
			endTime := time.Now()
			diff := endTime.Sub(startTime)
			_ = tr.Commit()
			fmt.Printf("Written %d, total %d in %f s\n", written, i, diff.Seconds())
			startTime = time.Now()
		}

		key, val := testStorage.CreateStoredData(uint64(i))
		err := tr.Update(key, val)

		assert.Nil(t, err)
	}
}
