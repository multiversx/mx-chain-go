package trie_test

import (
	"context"
	cryptoRand "crypto/rand"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/ElrondNetwork/elrond-go/trie/keyBuilder"
	"github.com/ElrondNetwork/elrond-go/trie/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyTrieHash = make([]byte, 32)

func emptyTrie() common.Trie {
	tr, _ := trie.NewTrie(getDefaultTrieParameters())

	return tr
}

func getDefaultTrieParameters() (common.StorageManager, marshal.Marshalizer, hashing.Hasher, uint) {
	marshalizer := &testscommon.ProtobufMarshalizerMock{}
	hasher := &testscommon.KeccakMock{}

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		SnapshotsGoroutineNum: 1,
	}
	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.NewSnapshotPruningStorerMock(),
		CheckpointsStorer:      testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:            marshalizer,
		Hasher:                 hasher,
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize),
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
	trieStorageManager, _ := trie.NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(5)

	return trieStorageManager, marshalizer, hasher, maxTrieLevelInMemory
}

func initTrieMultipleValues(nr int) (common.Trie, [][]byte) {
	tr := emptyTrie()

	var values [][]byte
	hsh := keccak.NewKeccak()

	for i := 0; i < nr; i++ {
		values = append(values, hsh.Compute(fmt.Sprint(i)))
		_ = tr.Update(values[i], values[i])
	}

	return tr, values
}

func initTrie() common.Trie {
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
		v, _, _ := tr.Get(val[i])
		assert.Equal(t, val[i], v)
	}
}

func TestPatriciaMerkleTree_GetEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	val, _, err := tr.Get([]byte("dog"))
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestPatriciaMerkleTree_Update(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	newVal := []byte("doge")
	_ = tr.Update([]byte("dog"), newVal)

	val, _, _ := tr.Get([]byte("dog"))
	assert.Equal(t, newVal, val)
}

func TestPatriciaMerkleTree_UpdateEmptyVal(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	var empty []byte

	_ = tr.Update([]byte("doe"), []byte{})

	v, _, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_UpdateNotExisting(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	_ = tr.Update([]byte("does"), []byte("this"))

	v, _, _ := tr.Get([]byte("does"))
	assert.Equal(t, []byte("this"), v)
}

func TestPatriciaMerkleTree_Delete(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	var empty []byte

	_ = tr.Delete([]byte("doe"))

	v, _, _ := tr.Get([]byte("doe"))
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

	root, err := tr.RootHash()
	assert.NotNil(t, root)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_NilRoot(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	root, err := tr.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, emptyTrieHash, root)
}

func TestPatriciaMerkleTree_Consistency(t *testing.T) {
	t.Parallel()

	tr := initTrie()
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

	tr := emptyTrie()
	nrInserts := 100
	wg := &sync.WaitGroup{}
	wg.Add(nrInserts)

	for i := 0; i < nrInserts; i++ {
		go func(index int) {
			err := tr.Update([]byte(strconv.Itoa(index)), []byte(strconv.Itoa(index)))
			assert.Nil(t, err)

			val, _, err := tr.Get([]byte(strconv.Itoa(index)))
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

	val, _, err := tr.Get([]byte("dog"))
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

	root1, _ := tr1.RootHash()
	root2, _ := tr2.RootHash()

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

	root1, _ := tr1.RootHash()
	root2, _ := tr2.RootHash()

	assert.Equal(t, root2, root1)
}

func TestPatriciaMerkleTrie_Recreate(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	rootHash, _ := tr.RootHash()
	_ = tr.Commit()

	newTr, err := tr.Recreate(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, newTr)

	root, _ := newTr.RootHash()
	assert.Equal(t, rootHash, root)
}

func TestPatriciaMerkleTrie_RecreateFromEpoch(t *testing.T) {
	t.Parallel()

	t.Run("nil options", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()

		newTr, err := tr.RecreateFromEpoch(nil)
		assert.Nil(t, newTr)
		assert.Equal(t, trie.ErrNilRootHashHolder, err)
	})

	t.Run("no epoch data", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		rootHash, _ := tr.RootHash()
		_ = tr.Commit()

		rootHashHolder := holders.NewRootHashHolder(rootHash, core.OptionalUint32{})
		newTr, err := tr.RecreateFromEpoch(rootHashHolder)
		assert.Nil(t, err)

		assert.True(t, trie.IsBaseTrieStorageManager(newTr.GetStorageManager()))
	})

	t.Run("with epoch data", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		rootHash, _ := tr.RootHash()
		_ = tr.Commit()

		optionalUint32 := core.OptionalUint32{
			Value:    5,
			HasValue: true,
		}
		rootHashHolder := holders.NewRootHashHolder(rootHash, optionalUint32)
		newTr, err := tr.RecreateFromEpoch(rootHashHolder)
		assert.Nil(t, err)

		assert.True(t, trie.IsTrieStorageManagerInEpoch(newTr.GetStorageManager()))
	})
}

func TestPatriciaMerkleTrie_RecreateWithInvalidRootHash(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	newTr, err := tr.Recreate(nil)
	assert.Nil(t, err)
	root, _ := newTr.RootHash()
	assert.Equal(t, emptyTrieHash, root)
}

func TestPatriciaMerkleTrie_GetSerializedNodes(t *testing.T) {
	t.Parallel()

	tr := initTrie()
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

	tr := initTrie()
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

	tr := initTrie()
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	storageManager := tr.GetStorageManager()
	dirtyHashes := trie.GetDirtyHashes(tr)
	storageManager.AddDirtyCheckpointHashes(rootHash, dirtyHashes)
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: nil,
		ErrChan:    make(chan error, 1),
	}
	storageManager.SetCheckpoint(rootHash, make([]byte, 0), iteratorChannels, nil, &trieMock.MockStatistics{})
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

	oldHashes := tr.GetObsoleteHashes()
	newHashes, _ := tr.GetDirtyHashes()

	assert.Equal(t, len(oldHashes), len(newHashes))
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

func TestPatriciaMerkleTrie_GetAllLeavesOnChannel(t *testing.T) {
	t.Parallel()

	t.Run("nil trie iterator channels", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		err := tr.GetAllLeavesOnChannel(nil, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder())
		assert.Equal(t, trie.ErrNilTrieIteratorChannels, err)
	})

	t.Run("nil leaves chan", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		iteratorChannels := &common.TrieIteratorChannels{
			LeavesChan: nil,
			ErrChan:    make(chan error, 1),
		}
		err := tr.GetAllLeavesOnChannel(iteratorChannels, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder())
		assert.Equal(t, trie.ErrNilTrieIteratorLeavesChannel, err)
	})

	t.Run("nil err chan", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		iteratorChannels := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    nil,
		}
		err := tr.GetAllLeavesOnChannel(iteratorChannels, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder())
		assert.Equal(t, trie.ErrNilTrieIteratorErrChannel, err)
	})

	t.Run("empty trie", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		leavesChannel := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    make(chan error, 1),
		}
		err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder())
		assert.Nil(t, err)
		assert.NotNil(t, leavesChannel)

		_, ok := <-leavesChannel.LeavesChan
		assert.False(t, ok)

		err = common.GetErrorFromChanNonBlocking(leavesChannel.ErrChan)
		assert.Nil(t, err)
	})

	t.Run("should fail on getting leaves", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		_ = tr.Commit()
		rootHash, _ := tr.RootHash()

		leavesChannel := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    make(chan error, 1),
		}

		expectedErr := errors.New("expected error")
		keyBuilderStub := &mock.KeyBuilderStub{}
		keyBuilderStub.GetKeyCalled = func() ([]byte, error) {
			return nil, expectedErr
		}
		keyBuilderStub.CloneCalled = func() common.KeyBuilder {
			return keyBuilderStub
		}

		err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), rootHash, keyBuilderStub)
		assert.Nil(t, err)
		assert.NotNil(t, leavesChannel)

		recovered := make(map[string][]byte)
		for leaf := range leavesChannel.LeavesChan {
			recovered[string(leaf.Key())] = leaf.Value()
		}
		err = common.GetErrorFromChanNonBlocking(leavesChannel.ErrChan)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 0, len(recovered))
	})

	t.Run("should work for first leaf but fail at second one", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		_ = tr.Update([]byte("doe"), []byte("reindeer"))
		_ = tr.Update([]byte("dog"), []byte("puppy"))
		_ = tr.Commit()
		rootHash, _ := tr.RootHash()

		leavesChannel := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    make(chan error, 1),
		}

		expectedErr := errors.New("expected error")

		keyBuilderStub := &mock.KeyBuilderStub{}
		firstRun := true
		keyBuilderStub.GetKeyCalled = func() ([]byte, error) {
			if firstRun {
				firstRun = false
				return []byte("doe"), nil
			}
			return nil, expectedErr
		}
		keyBuilderStub.CloneCalled = func() common.KeyBuilder {
			return keyBuilderStub
		}

		err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), rootHash, keyBuilderStub)
		assert.Nil(t, err)
		assert.NotNil(t, leavesChannel)

		recovered := make(map[string][]byte)
		for leaf := range leavesChannel.LeavesChan {
			recovered[string(leaf.Key())] = leaf.Value()
		}
		err = common.GetErrorFromChanNonBlocking(leavesChannel.ErrChan)
		assert.Equal(t, expectedErr, err)

		expectedLeaves := map[string][]byte{
			"doe": []byte("reindeer"),
		}
		assert.Equal(t, expectedLeaves, recovered)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		leaves := map[string][]byte{
			"doe":  []byte("reindeer"),
			"dog":  []byte("puppy"),
			"ddog": []byte("cat"),
		}
		_ = tr.Commit()
		rootHash, _ := tr.RootHash()

		leavesChannel := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    make(chan error, 1),
		}
		err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), rootHash, keyBuilder.NewKeyBuilder())
		assert.Nil(t, err)
		assert.NotNil(t, leavesChannel)

		recovered := make(map[string][]byte)
		for leaf := range leavesChannel.LeavesChan {
			recovered[string(leaf.Key())] = leaf.Value()
		}
		err = common.GetErrorFromChanNonBlocking(leavesChannel.ErrChan)
		assert.Nil(t, err)
		assert.Equal(t, leaves, recovered)
	})
}

func TestPatriciaMerkleTree_Prove(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	rootHash, _ := tr.RootHash()

	proof, value, err := tr.GetProof([]byte("dog"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("puppy"), value)
	ok, _ := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
}

func TestPatriciaMerkleTree_ProveCollapsedTrie(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	proof, _, err := tr.GetProof([]byte("dog"))
	assert.Nil(t, err)
	ok, _ := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
}

func TestPatriciaMerkleTree_ProveOnEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	proof, _, err := tr.GetProof([]byte("dog"))
	assert.Nil(t, proof)
	assert.Equal(t, trie.ErrNilNode, err)
}

func TestPatriciaMerkleTree_VerifyProof(t *testing.T) {
	t.Parallel()

	tr, val := initTrieMultipleValues(50)
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

	tr := emptyTrie()

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

	tr := emptyTrie()

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

	tr := initTrie()
	rootHash, _ := tr.RootHash()

	ok, err := tr.VerifyProof(rootHash, []byte("dog"), nil)
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_VerifyProofEmptyProofs(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	rootHash, _ := tr.RootHash()

	ok, err := tr.VerifyProof(rootHash, []byte("dog"), [][]byte{})
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTrie_VerifyProofFromDifferentTrieShouldNotWork(t *testing.T) {
	t.Parallel()

	tr1 := emptyTrie()
	tr2 := emptyTrie()

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

	tr := emptyTrie()

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

	tr := emptyTrie()

	numNodes := tr.GetNumNodes()
	assert.Equal(t, common.NumNodesDTO{}, numNodes)
}

func TestPatriciaMerkleTrie_GetTrieStats(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	_ = tr.Update([]byte("dog"), []byte("reindeer"))
	_ = tr.Update([]byte("fog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()

	rootHash, _ := tr.RootHash()
	address := "address"

	ts, ok := tr.(common.TrieStats)
	assert.True(t, ok)

	stats, err := ts.GetTrieStats(address, rootHash)
	assert.Nil(t, err)

	assert.Equal(t, rootHash, stats.RootHash)
	assert.Equal(t, address, stats.Address)

	assert.Equal(t, uint64(2), stats.NumBranchNodes)
	assert.Equal(t, uint64(1), stats.NumExtensionNodes)
	assert.Equal(t, uint64(3), stats.NumLeafNodes)
	assert.Equal(t, uint64(6), stats.TotalNumNodes)
	assert.Equal(t, uint32(3), stats.MaxTrieDepth)
}

func TestPatriciaMerkleTrie_GetNumNodes(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()
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

	tr := emptyTrie()
	_ = tr.Update([]byte("eod"), []byte("reindeer"))
	_ = tr.Update([]byte("god"), []byte("puppy"))
	_ = tr.Commit()
	expecterOldRoot, _ := tr.RootHash()

	_ = tr.Update([]byte("eggod"), []byte("cat"))
	assert.Equal(t, expecterOldRoot, tr.GetOldRoot())
}

func TestPatriciaMerkleTree_GetValueReturnsTrieDepth(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_, depth, err := tr.Get([]byte("doe"))
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), depth)
	_, depth, err = tr.Get([]byte("dog"))
	assert.Nil(t, err)
	assert.Equal(t, uint32(3), depth)
	_, depth, err = tr.Get([]byte("ddog"))
	assert.Nil(t, err)
	assert.Equal(t, uint32(3), depth)
}

func BenchmarkPatriciaMerkleTree_Insert(b *testing.B) {
	tr := emptyTrie()
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
	tr := emptyTrie()
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
	tr := emptyTrie()
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
	tr := emptyTrie()
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
	tr := emptyTrie()
	hsh := keccak.NewKeccak()

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		_ = tr.Update(values[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = tr.Get(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_GetCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
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
		_, _, _ = tr.Get(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Commit(b *testing.B) {
	nrValuesInTrie := 1000000
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hsh := keccak.NewKeccak()
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
	tr := emptyTrie()
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
