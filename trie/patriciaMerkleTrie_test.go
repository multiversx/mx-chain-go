package trie_test

import (
	cryptoRand "crypto/rand"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyTrieHash = make([]byte, 32)

func emptyTrie(tb testing.TB) common.Trie {
	tr, _ := trie.NewTrie(getDefaultTrieParameters(tb))

	return tr
}

func getDefaultTrieParameters(tb testing.TB) (common.StorageManager, marshal.Marshalizer, hashing.Hasher, uint) {
	db := testscommon.NewMemDbMock()
	marshalizer := &testscommon.ProtobufMarshalizerMock{}
	hasher := &testscommon.KeccakMock{}

	cfg := config.DBConfig{
		FilePath:          tb.TempDir(),
		Type:              string(storageUnit.LvlDBSerial),
		BatchDelaySeconds: 1,
		MaxBatchSize:      1,
		MaxOpenFiles:      10,
	}
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		MaxSnapshots:          2,
		SnapshotsGoroutineNum: 1,
	}
	args := trie.NewTrieStorageManagerArgs{
		DB:                     db,
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            marshalizer,
		Hasher:                 hasher,
		SnapshotDbConfig:       cfg,
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize),
		EpochNotifier:          &epochNotifier.EpochNotifierStub{},
	}
	trieStorageManager, _ := trie.NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(5)

	return trieStorageManager, marshalizer, hasher, maxTrieLevelInMemory
}

func initTrieMultipleValues(t *testing.T, nr int) (common.Trie, [][]byte) {
	tr := emptyTrie(t)

	var values [][]byte
	hsh := keccak.NewKeccak()

	for i := 0; i < nr; i++ {
		values = append(values, hsh.Compute(fmt.Sprint(i)))
		_ = tr.Update(values[i], values[i])
	}

	return tr, values
}

func initTrie(t *testing.T) common.Trie {
	tr := emptyTrie(t)
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))

	return tr
}

func TestNewTrieWithNilTrieStorage(t *testing.T) {
	t.Parallel()

	_, marshalizer, hasher, maxTrieLevelInMemory := getDefaultTrieParameters(t)
	tr, err := trie.NewTrie(nil, marshalizer, hasher, maxTrieLevelInMemory)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilTrieStorage, err)
}

func TestNewTrieWithNilMarshalizer(t *testing.T) {
	t.Parallel()

	trieStorage, _, hasher, maxTrieLevelInMemory := getDefaultTrieParameters(t)
	tr, err := trie.NewTrie(trieStorage, nil, hasher, maxTrieLevelInMemory)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieWithNilHasher(t *testing.T) {
	t.Parallel()

	trieStorage, marshalizer, _, maxTrieLevelInMemory := getDefaultTrieParameters(t)
	tr, err := trie.NewTrie(trieStorage, marshalizer, nil, maxTrieLevelInMemory)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieWithInvalidMaxTrieLevelInMemory(t *testing.T) {
	t.Parallel()

	trieStorage, marshalizer, hasher, _ := getDefaultTrieParameters(t)
	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher, 0)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrInvalidLevelValue, err)
}

func TestPatriciaMerkleTree_Get(t *testing.T) {
	t.Parallel()

	tr, val := initTrieMultipleValues(t, 10000)

	for i := range val {
		v, _ := tr.Get(val[i])
		assert.Equal(t, val[i], v)
	}
}

func TestPatriciaMerkleTree_GetEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	val, err := tr.Get([]byte("dog"))
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestPatriciaMerkleTree_Update(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	newVal := []byte("doge")
	_ = tr.Update([]byte("dog"), newVal)

	val, _ := tr.Get([]byte("dog"))
	assert.Equal(t, newVal, val)
}

func TestPatriciaMerkleTree_UpdateEmptyVal(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	var empty []byte

	_ = tr.Update([]byte("doe"), []byte{})

	v, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_UpdateNotExisting(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	_ = tr.Update([]byte("does"), []byte("this"))

	v, _ := tr.Get([]byte("does"))
	assert.Equal(t, []byte("this"), v)
}

func TestPatriciaMerkleTree_Delete(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	var empty []byte

	_ = tr.Delete([]byte("doe"))

	v, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_DeleteEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	err := tr.Delete([]byte("dog"))
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_Root(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	root, err := tr.RootHash()
	assert.NotNil(t, root)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_NilRoot(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	root, err := tr.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, emptyTrieHash, root)
}

func TestPatriciaMerkleTree_Consistency(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	root1, _ := tr.RootHash()

	_ = tr.Update([]byte("dodge"), []byte("viper"))
	root2, _ := tr.RootHash()

	_ = tr.Delete([]byte("dodge"))
	root3, _ := tr.RootHash()

	assert.Equal(t, root1, root3)
	assert.NotEqual(t, root1, root2)
}

func TestPatriciaMerkleTrie_UpdateAndGetConcurrently(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)
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

	rootHash, _ := tr.RootHash()
	assert.NotEqual(t, emptyTrieHash, rootHash)

	wg.Add(nrInserts)
	for i := 0; i < nrInserts; i++ {
		go func(index int) {
			assert.Nil(t, tr.Delete([]byte(strconv.Itoa(index))))
			wg.Done()
		}(i)
	}
	wg.Wait()

	rootHash, _ = tr.RootHash()
	assert.Equal(t, emptyTrieHash, rootHash)
}

func TestPatriciaMerkleTree_Commit(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitCollapsesTrieOk(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	_ = tr.Update([]byte("zebra"), []byte("zebra"))
	_ = tr.Update([]byte("doggo"), []byte("doggo"))
	_ = tr.Update([]byte("doggless"), []byte("doggless"))

	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitAfterCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	_ = tr.Commit()
	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitEmptyRoot(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_GetAfterCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	err := tr.Commit()
	assert.Nil(t, err)

	val, err := tr.Get([]byte("dog"))
	assert.Equal(t, []byte("puppy"), val)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_InsertAfterCommit(t *testing.T) {
	t.Parallel()

	tr1 := initTrie(t)
	tr2 := initTrie(t)

	err := tr1.Commit()
	assert.Nil(t, err)

	_ = tr1.Update([]byte("doge"), []byte("coin"))
	_ = tr2.Update([]byte("doge"), []byte("coin"))

	root1, _ := tr1.RootHash()
	root2, _ := tr2.RootHash()

	assert.Equal(t, root2, root1)
}

func TestPatriciaMerkleTree_DeleteAfterCommit(t *testing.T) {
	t.Parallel()

	tr1 := initTrie(t)
	tr2 := initTrie(t)

	err := tr1.Commit()
	assert.Nil(t, err)

	_ = tr1.Delete([]byte("dogglesworth"))
	_ = tr2.Delete([]byte("dogglesworth"))

	root1, _ := tr1.RootHash()
	root2, _ := tr2.RootHash()

	assert.Equal(t, root2, root1)
}

func TestPatriciaMerkleTrie_Recreate(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	rootHash, _ := tr.RootHash()
	_ = tr.Commit()

	newTr, err := tr.Recreate(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, newTr)

	root, _ := newTr.RootHash()
	assert.Equal(t, rootHash, root)
}

func TestPatriciaMerkleTrie_RecreateWithInvalidRootHash(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	newTr, err := tr.Recreate(nil)
	assert.Nil(t, err)
	root, _ := newTr.RootHash()
	assert.Equal(t, emptyTrieHash, root)
}

func TestPatriciaMerkleTrie_GetSerializedNodes(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	maxBuffToSend := uint64(500)
	expectedNodes := 6
	serializedNodes, _, err := tr.GetSerializedNodes(rootHash, maxBuffToSend)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodes, len(serializedNodes))
}

func TestPatriciaMerkleTrie_GetSerializedNodesTinyBufferShouldNotGetAllNodes(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	maxBuffToSend := uint64(150)
	expectedNodes := 2
	serializedNodes, _, err := tr.GetSerializedNodes(rootHash, maxBuffToSend)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodes, len(serializedNodes))
}

func TestPatriciaMerkleTrie_GetSerializedNodesGetFromCheckpoint(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	storageManager := tr.GetStorageManager()
	dirtyHashes := trie.GetDirtyHashes(tr)
	storageManager.AddDirtyCheckpointHashes(rootHash, dirtyHashes)
	storageManager.SetCheckpoint(rootHash, make([]byte, 0), nil, &trieMock.MockStatistics{})
	trie.WaitForOperationToComplete(storageManager)

	err := storageManager.Remove(rootHash)
	assert.Nil(t, err)

	maxBuffToSend := uint64(500)
	expectedNodes := 6
	serializedNodes, _, err := tr.GetSerializedNodes(rootHash, maxBuffToSend)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodes, len(serializedNodes))
}

func TestPatriciaMerkleTrie_String(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	str := tr.String()
	assert.NotEqual(t, 0, len(str))

	tr = emptyTrie(t)
	str = tr.String()
	assert.Equal(t, "*** EMPTY TRIE ***\n", str)
}

func TestPatriciaMerkleTree_reduceBranchNodeReturnsOldHashesCorrectly(t *testing.T) {
	t.Parallel()

	key1 := []byte("ABC")
	key2 := []byte("ABD")
	val1 := []byte("val1")
	val2 := []byte("val2")

	tr := emptyTrie(t)
	_ = tr.Update(key1, val1)
	_ = tr.Update(key2, val2)
	_ = tr.Commit()

	_ = tr.Update(key1, nil)
	_ = tr.Update(key1, val1)

	oldHashes := tr.GetObsoleteHashes()
	newHashes, _ := tr.GetDirtyHashes()

	assert.Equal(t, len(oldHashes), len(newHashes))
}

func TestPatriciaMerkleTrie_GetAllHashesSetsHashes(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	hashes, err := tr.GetAllHashes()
	assert.Nil(t, err)
	assert.Equal(t, 6, len(hashes))
}

func TestPatriciaMerkleTrie_GetAllHashesEmtyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	hashes, err := tr.GetAllHashes()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(hashes))
}

func TestPatriciaMerkleTrie_GetAllLeavesOnChannelEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	leavesChannel, err := tr.GetAllLeavesOnChannel([]byte{})
	assert.Nil(t, err)
	assert.NotNil(t, leavesChannel)

	_, ok := <-leavesChannel
	assert.False(t, ok)
}

func TestPatriciaMerkleTrie_GetAllLeavesOnChannel(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	leaves := map[string][]byte{
		"doe":  []byte("reindeer"),
		"dog":  []byte("puppy"),
		"ddog": []byte("cat"),
	}
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	leavesChannel, err := tr.GetAllLeavesOnChannel(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, leavesChannel)

	recovered := make(map[string][]byte)
	for leaf := range leavesChannel {
		recovered[string(leaf.Key())] = leaf.Value()
	}
	assert.Equal(t, leaves, recovered)
}

func TestPatriciaMerkleTree_Prove(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	rootHash, _ := tr.RootHash()

	proof, value, err := tr.GetProof([]byte("dog"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("puppy"), value)
	ok, _ := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
}

func TestPatriciaMerkleTree_ProveCollapsedTrie(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	proof, _, err := tr.GetProof([]byte("dog"))
	assert.Nil(t, err)
	ok, _ := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
}

func TestPatriciaMerkleTree_ProveOnEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	proof, _, err := tr.GetProof([]byte("dog"))
	assert.Nil(t, proof)
	assert.Equal(t, trie.ErrNilNode, err)
}

func TestPatriciaMerkleTree_VerifyProof(t *testing.T) {
	t.Parallel()

	tr, val := initTrieMultipleValues(t, 50)
	rootHash, _ := tr.RootHash()

	for i := range val {
		proof, _, _ := tr.GetProof(val[i])

		ok, err := tr.VerifyProof(rootHash, val[i], proof)
		assert.Nil(t, err)
		assert.True(t, ok)

		ok, err = tr.VerifyProof(rootHash, []byte("dog"+strconv.Itoa(i)), proof)
		assert.Nil(t, err)
		assert.False(t, ok)
	}
}

func TestPatriciaMerkleTrie_VerifyProofBranchNodeWantHashShouldWork(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	_ = tr.Update([]byte("dog"), []byte("cat"))
	_ = tr.Update([]byte("zebra"), []byte("horse"))
	rootHash, _ := tr.RootHash()

	proof, _, _ := tr.GetProof([]byte("dog"))
	ok, err := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTrie_VerifyProofExtensionNodeWantHashShouldWork(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	_ = tr.Update([]byte("dog"), []byte("cat"))
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	rootHash, _ := tr.RootHash()

	proof, _, _ := tr.GetProof([]byte("dog"))
	ok, err := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_VerifyProofNilProofs(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	rootHash, _ := tr.RootHash()

	ok, err := tr.VerifyProof(rootHash, []byte("dog"), nil)
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_VerifyProofEmptyProofs(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	rootHash, _ := tr.RootHash()

	ok, err := tr.VerifyProof(rootHash, []byte("dog"), [][]byte{})
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTrie_VerifyProofFromDifferentTrieShouldNotWork(t *testing.T) {
	t.Parallel()

	tr1 := emptyTrie(t)
	tr2 := emptyTrie(t)

	_ = tr1.Update([]byte("doe"), []byte("reindeer"))
	_ = tr1.Update([]byte("dog"), []byte("puppy"))
	_ = tr1.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr2.Update([]byte("doe"), []byte("reindeer"))
	_ = tr2.Update([]byte("dog"), []byte("puppy"))
	_ = tr2.Update([]byte("dogglesworth"), []byte("caterpillar"))
	rootHash, _ := tr1.RootHash()

	proof, _, _ := tr2.GetProof([]byte("dogglesworth"))
	ok, _ := tr1.VerifyProof(rootHash, []byte("dogglesworth"), proof)
	assert.False(t, ok)
}

func TestPatriciaMerkleTrie_GetAndVerifyProof(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	nrLeaves := 10000
	values := make([][]byte, nrLeaves)
	numRuns := 5000

	for i := 0; i < nrLeaves; i++ {
		buff := make([]byte, 32)
		_, _ = cryptoRand.Read(buff)

		values[i] = buff
		_ = tr.Update(values[i], values[i])
	}

	rootHash, _ := tr.RootHash()
	for i := 0; i < numRuns; i++ {
		randNum := rand.Intn(nrLeaves)
		proof, _, err := tr.GetProof(values[randNum])
		if err != nil {
			dumpTrieContents(tr, values)
			fmt.Printf("error getting proof for %v, err = %s\n", values[randNum], err.Error())
		}
		require.Nil(t, err)
		require.NotEqual(t, 0, len(proof))

		ok, err := tr.VerifyProof(rootHash, values[randNum], proof)
		if err != nil {
			dumpTrieContents(tr, values)
			fmt.Printf("error verifying proof for %v, proof = %v, err = %s\n", values[randNum], proof, err.Error())
		}
		require.Nil(t, err)
		require.True(t, ok)
	}
}

func dumpTrieContents(tr common.Trie, values [][]byte) {
	fmt.Println(tr.String())
	for _, val := range values {
		fmt.Println(val)
	}
}

func TestPatriciaMerkleTrie_GetNumNodesNilRootShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)

	numNodes := tr.GetNumNodes()
	assert.Equal(t, common.NumNodesDTO{}, numNodes)
}

func TestPatriciaMerkleTrie_GetNumNodes(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)
	_ = tr.Update([]byte("eod"), []byte("reindeer"))
	_ = tr.Update([]byte("god"), []byte("puppy"))
	_ = tr.Update([]byte("eggod"), []byte("cat"))

	numNodes := tr.GetNumNodes()
	assert.Equal(t, 5, numNodes.MaxLevel)
	assert.Equal(t, 3, numNodes.Leaves)
	assert.Equal(t, 2, numNodes.Extensions)
	assert.Equal(t, 2, numNodes.Branches)
}

func TestPatriciaMerkleTrie_GetOldRoot(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)
	_ = tr.Update([]byte("eod"), []byte("reindeer"))
	_ = tr.Update([]byte("god"), []byte("puppy"))
	_ = tr.Commit()
	expecterOldRoot, _ := tr.RootHash()

	_ = tr.Update([]byte("eggod"), []byte("cat"))
	assert.Equal(t, expecterOldRoot, tr.GetOldRoot())
}

func BenchmarkPatriciaMerkleTree_Insert(b *testing.B) {
	tr := emptyTrie(b)
	hsh := keccak.NewKeccak()

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
	tr := emptyTrie(b)
	hsh := keccak.NewKeccak()

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
	tr := emptyTrie(b)
	hsh := keccak.NewKeccak()

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
	tr := emptyTrie(b)
	hsh := keccak.NewKeccak()

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
	tr := emptyTrie(b)
	hsh := keccak.NewKeccak()

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
	tr := emptyTrie(b)
	hsh := keccak.NewKeccak()

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
		hsh := keccak.NewKeccak()
		tr := emptyTrie(b)
		for j := 0; j < nrValuesInTrie; j++ {
			hash := hsh.Compute(strconv.Itoa(j))
			_ = tr.Update(hash, hash)
		}
		b.StartTimer()

		_ = tr.Commit()
	}
}

func BenchmarkPatriciaMerkleTrie_RootHashAfterChanging30000Nodes(b *testing.B) {
	tr := emptyTrie(b)
	hsh := keccak.NewKeccak()

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
		_, _ = tr.RootHash()
	}
}

func BenchmarkPatriciaMerkleTrie_RootHashAfterChanging30000NodesInBatchesOf200(b *testing.B) {
	tr := emptyTrie(b)
	hsh := keccak.NewKeccak()

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
				_, _ = tr.RootHash()
			}
		}
	}
}
