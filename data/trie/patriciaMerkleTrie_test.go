package trie_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

var emptyTrieHash = make([]byte, 32)

func emptyTrie() data.Trie {
	tr, _ := trie.NewTrie(getDefaultTrieParameters())

	return tr
}

func getDefaultTrieParameters() (data.StorageManager, marshal.Marshalizer, hashing.Hasher, uint) {
	db := mock.NewMemDbMock()
	marshalizer := &mock.ProtobufMarshalizerMock{}
	hasher := &mock.KeccakMock{}

	tempDir, _ := ioutil.TempDir("", strconv.Itoa(rand.Intn(100000)))

	cfg := config.DBConfig{
		FilePath:          tempDir,
		Type:              string(storageUnit.LvlDBSerial),
		BatchDelaySeconds: 1,
		MaxBatchSize:      1,
		MaxOpenFiles:      10,
	}
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}

	evictionWaitingList, _ := mock.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	trieStorageManager, _ := trie.NewTrieStorageManager(db, marshalizer, hasher, cfg, evictionWaitingList, generalCfg)
	maxTrieLevelInMemory := uint(5)

	return trieStorageManager, marshalizer, hasher, maxTrieLevelInMemory
}

func initTrieMultipleValues(nr int) (data.Trie, [][]byte) {
	tr := emptyTrie()

	var values [][]byte
	hsh := keccak.Keccak{}

	for i := 0; i < nr; i++ {
		values = append(values, hsh.Compute(fmt.Sprint(i)))
		_ = tr.Update(values[i], values[i])
	}

	return tr, values
}

func initTrie() data.Trie {
	tr := emptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))

	return tr
}

func TestNewTrieWithNilTrieStorage(t *testing.T) {
	t.Parallel()

	_, marshalizer, hasher, maxTrieLevelInMemory := getDefaultTrieParameters()
	tr, err := trie.NewTrie(nil, marshalizer, hasher, maxTrieLevelInMemory)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilTrieStorage, err)
}

func TestNewTrieWithNilMarshalizer(t *testing.T) {
	t.Parallel()

	trieStorage, _, hasher, maxTrieLevelInMemory := getDefaultTrieParameters()
	tr, err := trie.NewTrie(trieStorage, nil, hasher, maxTrieLevelInMemory)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieWithNilHasher(t *testing.T) {
	t.Parallel()

	trieStorage, marshalizer, _, maxTrieLevelInMemory := getDefaultTrieParameters()
	tr, err := trie.NewTrie(trieStorage, marshalizer, nil, maxTrieLevelInMemory)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieWithInvalidMaxTrieLevelInMemory(t *testing.T) {
	t.Parallel()

	trieStorage, marshalizer, hasher, _ := getDefaultTrieParameters()
	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher, 0)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrInvalidLevelValue, err)
}

func TestPatriciaMerkleTree_Get(t *testing.T) {
	t.Parallel()

	tr, val := initTrieMultipleValues(10000)

	for i := range val {
		v, _ := tr.Get(val[i])
		assert.Equal(t, val[i], v)
	}
}

func TestPatriciaMerkleTree_GetEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	val, err := tr.Get([]byte("dog"))
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestPatriciaMerkleTree_Update(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	newVal := []byte("doge")
	_ = tr.Update([]byte("dog"), newVal)

	val, _ := tr.Get([]byte("dog"))
	assert.Equal(t, newVal, val)
}

func TestPatriciaMerkleTree_UpdateEmptyVal(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	var empty []byte

	_ = tr.Update([]byte("doe"), []byte{})

	v, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_UpdateNotExisting(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	_ = tr.Update([]byte("does"), []byte("this"))

	v, _ := tr.Get([]byte("does"))
	assert.Equal(t, []byte("this"), v)
}

func TestPatriciaMerkleTree_Delete(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	var empty []byte

	_ = tr.Delete([]byte("doe"))

	v, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_DeleteEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	err := tr.Delete([]byte("dog"))
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_Root(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	root, err := tr.Root()
	assert.NotNil(t, root)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_NilRoot(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	root, err := tr.Root()
	assert.Nil(t, err)
	assert.Equal(t, emptyTrieHash, root)
}

func TestPatriciaMerkleTree_Consistency(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	root1, _ := tr.Root()

	_ = tr.Update([]byte("dodge"), []byte("viper"))
	root2, _ := tr.Root()

	_ = tr.Delete([]byte("dodge"))
	root3, _ := tr.Root()

	assert.Equal(t, root1, root3)
	assert.NotEqual(t, root1, root2)
}

func TestPatriciaMerkleTrie_UpdateAndGetConcurrently(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()
	nrInserts := 100
	wg := &sync.WaitGroup{}
	wg.Add(nrInserts)

	for i := 0; i < nrInserts; i++ {
		go func(index int) {
			err := tr.Update([]byte(strconv.Itoa(index)), []byte(strconv.Itoa(index)))
			assert.Nil(t, err)

			val, err := tr.Get([]byte(strconv.Itoa(index)))
			assert.Nil(t, err)
			assert.Equal(t, []byte(strconv.Itoa(index)), val)

			wg.Done()
		}(i)
	}
	wg.Wait()

	rootHash, _ := tr.Root()
	assert.NotEqual(t, emptyTrieHash, rootHash)

	wg.Add(nrInserts)
	for i := 0; i < nrInserts; i++ {
		go func(index int) {
			assert.Nil(t, tr.Delete([]byte(strconv.Itoa(index))))
			wg.Done()
		}(i)
	}
	wg.Wait()

	rootHash, _ = tr.Root()
	assert.Equal(t, emptyTrieHash, rootHash)
}

func TestPatriciaMerkleTree_Commit(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitCollapsesTrieOk(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	_ = tr.Update([]byte("zebra"), []byte("zebra"))
	_ = tr.Update([]byte("doggo"), []byte("doggo"))
	_ = tr.Update([]byte("doggless"), []byte("doggless"))

	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitAfterCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	_ = tr.Commit()
	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitEmptyRoot(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_GetAfterCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	err := tr.Commit()
	assert.Nil(t, err)

	val, err := tr.Get([]byte("dog"))
	assert.Equal(t, []byte("puppy"), val)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_InsertAfterCommit(t *testing.T) {
	t.Parallel()

	tr1 := initTrie()
	tr2 := initTrie()

	err := tr1.Commit()
	assert.Nil(t, err)

	_ = tr1.Update([]byte("doge"), []byte("coin"))
	_ = tr2.Update([]byte("doge"), []byte("coin"))

	root1, _ := tr1.Root()
	root2, _ := tr2.Root()

	assert.Equal(t, root2, root1)
}

func TestPatriciaMerkleTree_DeleteAfterCommit(t *testing.T) {
	t.Parallel()

	tr1 := initTrie()
	tr2 := initTrie()

	err := tr1.Commit()
	assert.Nil(t, err)

	_ = tr1.Delete([]byte("dogglesworth"))
	_ = tr2.Delete([]byte("dogglesworth"))

	root1, _ := tr1.Root()
	root2, _ := tr2.Root()

	assert.Equal(t, root2, root1)
}

func TestPatriciaMerkleTrie_Recreate(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	rootHash, _ := tr.Root()
	_ = tr.Commit()

	newTr, err := tr.Recreate(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, newTr)

	root, _ := newTr.Root()
	assert.Equal(t, rootHash, root)
}

func TestPatriciaMerkleTrie_RecreateWithInvalidRootHash(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	newTr, err := tr.Recreate(nil)
	assert.Nil(t, err)
	root, _ := newTr.Root()
	assert.Equal(t, emptyTrieHash, root)
}

func TestPatriciaMerkleTrie_PruneAfterCancelPruneShouldFail(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash, _ := tr.Root()

	_ = tr.Update([]byte("dog"), []byte("value of dog"))
	_ = tr.Commit()

	tr.CancelPrune(rootHash, data.OldRoot)
	tr.Prune(rootHash, data.OldRoot)

	newTr, err := tr.Recreate(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, newTr)
}

func TestPatriciaMerkleTrie_Prune(t *testing.T) {
	t.Parallel()

	tr, _ := trie.NewTrie(getDefaultTrieParameters())

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	rootHash, _ := tr.Root()

	_ = tr.Update([]byte("dog"), []byte("value of dog"))
	_ = tr.Commit()

	tr.CancelPrune(rootHash, data.NewRoot)
	tr.Prune(rootHash, data.OldRoot)
	time.Sleep(time.Second)

	expectedErr := fmt.Errorf("key: %s not found", base64.StdEncoding.EncodeToString(rootHash))
	val, err := tr.Database().Get(rootHash)
	assert.Nil(t, val)
	assert.Equal(t, expectedErr, err)
}

func TestPatriciaMerkleTrie_GetSerializedNodes(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash, _ := tr.Root()

	maxBuffToSend := uint64(500)
	expectedNodes := 6
	serializedNodes, _, err := tr.GetSerializedNodes(rootHash, maxBuffToSend)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodes, len(serializedNodes))
}

func TestPatriciaMerkleTrie_GetSerializedNodesTinyBufferShouldNotGetAllNodes(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash, _ := tr.Root()

	maxBuffToSend := uint64(150)
	expectedNodes := 2
	serializedNodes, _, err := tr.GetSerializedNodes(rootHash, maxBuffToSend)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodes, len(serializedNodes))
}

func TestPatriciaMerkleTrie_GetSerializedNodesGetFromSnapshot(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash, _ := tr.Root()

	dirtyHashes, _ := tr.GetDirtyHashes()
	tr.SetNewHashes(dirtyHashes)

	tr.TakeSnapshot(rootHash)
	time.Sleep(time.Second)
	tr.Prune(rootHash, data.NewRoot)

	maxBuffToSend := uint64(500)
	expectedNodes := 6
	serializedNodes, _, err := tr.GetSerializedNodes(rootHash, maxBuffToSend)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodes, len(serializedNodes))
}

func TestPatriciaMerkleTrie_GetAllLeaves(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	leaves, err := tr.GetAllLeaves()

	assert.Nil(t, err)
	assert.Equal(t, 3, len(leaves))
	assert.Equal(t, []byte("reindeer"), leaves["doe"])
	assert.Equal(t, []byte("puppy"), leaves["dog"])
	assert.Equal(t, []byte("cat"), leaves["ddog"])
}

func TestPatriciaMerkleTrie_String(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	str := tr.String()
	assert.NotEqual(t, 0, len(str))

	tr = emptyTrie()
	str = tr.String()
	assert.Equal(t, "*** EMPTY TRIE ***\n", str)
}

func TestPatriciaMerkleTree_reduceBranchNodeReturnsOldHashesCorrectly(t *testing.T) {
	t.Parallel()

	key1 := []byte("ABC")
	key2 := []byte("ABD")
	val1 := []byte("val1")
	val2 := []byte("val2")

	tr := emptyTrie()
	_ = tr.Update(key1, val1)
	_ = tr.Update(key2, val2)
	_ = tr.Commit()

	_ = tr.Update(key1, nil)
	_ = tr.Update(key1, val1)

	oldHashes := tr.ResetOldHashes()
	newHashes, _ := tr.GetDirtyHashes()

	assert.Equal(t, len(oldHashes), len(newHashes))
}

func TestPatriciaMerkleTrie_GetSerializedNodesFromSnapshotShouldNotCommitToMainDB(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	newHashes, _ := tr.GetDirtyHashes()
	tr.SetNewHashes(newHashes)
	_ = tr.Commit()

	rootHash, _ := tr.Root()
	tr.TakeSnapshot(rootHash)
	time.Sleep(time.Second)

	tr.Prune(rootHash, data.NewRoot)

	val, err := tr.Database().Get(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)

	nodes, _, _ := tr.GetSerializedNodes(rootHash, 2000)
	assert.NotEqual(t, 0, len(nodes))

	val, err = tr.Database().Get(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)
}

func TestPatriciaMerkleTrie_GetSerializedNodesShouldCheckFirstInSnapshotsDB(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.ProtobufMarshalizerMock{}
	hasher := &mock.KeccakMock{}

	getDbCalled := false
	getSnapshotCalled := false

	trieStorageManager := &mock.StorageManagerStub{
		GetDbThatContainsHashCalled: func(bytes []byte) data.DBWriteCacher {
			getDbCalled = true
			return nil
		},
		DatabaseCalled: func() data.DBWriteCacher {
			getSnapshotCalled = true
			return mock.NewMemDbMock()
		},
	}
	maxTrieLevelInMemory := uint(5)

	tr, _ := trie.NewTrie(trieStorageManager, marshalizer, hasher, maxTrieLevelInMemory)

	rootHash := []byte("rootHash")
	_, _, _ = tr.GetSerializedNodes(rootHash, 2000)

	assert.False(t, getDbCalled)
	assert.True(t, getSnapshotCalled)
}

func TestPatriciaMerkleTrie_GetAllHashesSetsHashes(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	hashes, err := tr.GetAllHashes()
	assert.Nil(t, err)
	assert.Equal(t, 6, len(hashes))
}

func TestPatriciaMerkleTrie_GetAllHashesEmtyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	hashes, err := tr.GetAllHashes()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(hashes))
}

func TestPatriciaMerkleTrie_GetAllLeavesOnChannelNilTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	leavesChannel := tr.GetAllLeavesOnChannel()
	assert.NotNil(t, leavesChannel)

	_, ok := <-leavesChannel
	assert.False(t, ok)
}

func TestPatriciaMerkleTrie_GetAllLeavesOnChannel(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	leaves := map[string][]byte{
		"doe":  []byte("reindeer"),
		"dog":  []byte("puppy"),
		"ddog": []byte("cat"),
	}

	leavesChannel := tr.GetAllLeavesOnChannel()
	assert.NotNil(t, leavesChannel)

	recovered := make(map[string][]byte)
	for leaf := range leavesChannel {
		recovered[string(leaf.Key())] = leaf.Value()
	}
	assert.Equal(t, leaves, recovered)
}

func BenchmarkPatriciaMerkleTree_Insert(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 1000000
	nrValuesNotInTrie := 9000000
	values := make([][]byte, nrValuesNotInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		val := hsh.Compute(strconv.Itoa(i))
		_ = tr.Update(val, val)
	}
	for i := 0; i < nrValuesNotInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i + nrValuesInTrie))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Update(values[i%nrValuesNotInTrie], values[i%nrValuesNotInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_InsertCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 1000000
	nrValuesNotInTrie := 9000000
	values := make([][]byte, nrValuesNotInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		val := hsh.Compute(strconv.Itoa(i))
		_ = tr.Update(val, val)
	}
	for i := 0; i < nrValuesNotInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i + nrValuesInTrie))
	}
	_ = tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Update(values[i%nrValuesNotInTrie], values[i%nrValuesNotInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Delete(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		_ = tr.Update(values[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Delete(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_DeleteCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 1000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		_ = tr.Update(values[i], values[i])
	}

	_ = tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Delete(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Get(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		_ = tr.Update(values[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tr.Get(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_GetCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 1000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		_ = tr.Update(values[i], values[i])
	}
	_ = tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tr.Get(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Commit(b *testing.B) {
	nrValuesInTrie := 1000000
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hsh := keccak.Keccak{}
		tr := emptyTrie()
		for j := 0; j < nrValuesInTrie; j++ {
			hash := hsh.Compute(strconv.Itoa(j))
			_ = tr.Update(hash, hash)
		}
		b.StartTimer()

		_ = tr.Commit()
	}
}

func BenchmarkPatriciaMerkleTrie_RootHashAfterChanging30000Nodes(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 2000000
	values := make([][]byte, nrValuesInTrie)
	nrOfValuesToModify := 30000

	for i := 0; i < nrValuesInTrie; i++ {
		key := hsh.Compute(strconv.Itoa(i))
		value := append(key, []byte(strconv.Itoa(i))...)

		_ = tr.Update(key, value)
		values[i] = key
	}
	_ = tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < nrOfValuesToModify; j++ {
			_ = tr.Update(values[j], values[j])
		}
		b.StartTimer()
		_, _ = tr.Root()
	}
}

func BenchmarkPatriciaMerkleTrie_RootHashAfterChanging30000NodesInBatchesOf200(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 2000000
	values := make([][]byte, nrValuesInTrie)
	nrOfValuesToModify := 30000
	nrOfValuesToCommit := 200

	for i := 0; i < nrValuesInTrie; i++ {
		key := hsh.Compute(strconv.Itoa(i))
		value := append(key, []byte(strconv.Itoa(i))...)

		_ = tr.Update(key, value)
		values[i] = key
	}
	_ = tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < nrOfValuesToModify; j++ {
			b.StopTimer()
			_ = tr.Update(values[j], values[j])
			if j%nrOfValuesToCommit == 0 {
				b.StartTimer()
				_, _ = tr.Root()
			}
		}
	}
}
