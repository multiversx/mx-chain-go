package trie_test

import (
	"bytes"
	"context"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	errorsCommon "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/mock"
)

var emptyTrieHash = make([]byte, 32)

func emptyTrie() common.Trie {
	tr, _ := trie.NewTrie(getDefaultTrieParameters())

	return tr
}

func emptyTrieWithCustomEnableEpochsHandler(handler common.EnableEpochsHandler) common.Trie {
	storage, marshaller, hasher, _, maxSizeInMem := getDefaultTrieParameters()

	tr, _ := trie.NewTrie(storage, marshaller, hasher, handler, maxSizeInMem)
	return tr
}

func getDefaultTrieParameters() (common.StorageManager, marshal.Marshalizer, hashing.Hasher, common.EnableEpochsHandler, common.TrieCollapseManager) {
	args := trie.GetDefaultTrieStorageManagerParameters()
	trieStorageManager, _ := trie.NewTrieStorageManager(args)

	return trieStorageManager, args.Marshalizer, args.Hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, collapseManager.NewDisabledCollapseManager()
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
	addDefaultDataToTrie(tr)

	return tr
}

func addDefaultDataToTrie(tr common.Trie) {
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))
}

func TestNewTrieWithNilTrieStorage(t *testing.T) {
	t.Parallel()

	_, marshalizer, hasher, enableEpochsHandler, cm := getDefaultTrieParameters()
	tr, err := trie.NewTrie(nil, marshalizer, hasher, enableEpochsHandler, cm)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilTrieStorage, err)
}

func TestNewTrieWithNilMarshalizer(t *testing.T) {
	t.Parallel()

	trieStorage, _, hasher, enableEpochsHandler, cm := getDefaultTrieParameters()
	tr, err := trie.NewTrie(trieStorage, nil, hasher, enableEpochsHandler, cm)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieWithNilHasher(t *testing.T) {
	t.Parallel()

	trieStorage, marshalizer, _, enableEpochsHandler, cm := getDefaultTrieParameters()
	tr, err := trie.NewTrie(trieStorage, marshalizer, nil, enableEpochsHandler, cm)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieWithNilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	trieStorage, marshalizer, hasher, _, cm := getDefaultTrieParameters()
	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher, nil, cm)

	assert.Nil(t, tr)
	assert.Equal(t, errorsCommon.ErrNilEnableEpochsHandler, err)
}

func TestNewTrieWithNilCollapseManager(t *testing.T) {
	t.Parallel()

	trieStorage, marshalizer, hasher, enableEpochsHandler, _ := getDefaultTrieParameters()
	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher, enableEpochsHandler, nil)

	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, trie.ErrNilCollapseManager))
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

func TestPatriciaMerkleTree_DeleteNotPresent(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	err := tr.Commit()
	assert.Nil(t, err)

	err = tr.Delete([]byte("adog"))
	assert.Nil(t, err)
}

func TestPatriciaMerkleTrie_Recreate(t *testing.T) {
	t.Parallel()

	t.Run("nil options", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()

		newTr, err := tr.Recreate(nil)
		assert.Nil(t, newTr)
		assert.Equal(t, trie.ErrNilRootHashHolder, err)
	})

	t.Run("no epoch data", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		rootHash, _ := tr.RootHash()
		_ = tr.Commit()

		rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
		newTr, err := tr.Recreate(rootHashHolder)
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
		newTr, err := tr.Recreate(rootHashHolder)
		assert.Nil(t, err)

		assert.True(t, trie.IsTrieStorageManagerInEpoch(newTr.GetStorageManager()))
	})
}

func TestPatriciaMerkleTrie_RecreateWithInvalidRootHash(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	newTr, err := tr.Recreate(holders.NewDefaultRootHashesHolder([]byte{}))
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
		err := tr.GetAllLeavesOnChannel(nil, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder(), parsers.NewMainTrieLeafParser())
		assert.Equal(t, trie.ErrNilTrieIteratorChannels, err)
	})

	t.Run("nil leaves chan", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		iteratorChannels := &common.TrieIteratorChannels{
			LeavesChan: nil,
			ErrChan:    errChan.NewErrChanWrapper(),
		}
		err := tr.GetAllLeavesOnChannel(iteratorChannels, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder(), parsers.NewMainTrieLeafParser())
		assert.Equal(t, trie.ErrNilTrieIteratorLeavesChannel, err)
	})

	t.Run("nil err chan", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		iteratorChannels := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    nil,
		}
		err := tr.GetAllLeavesOnChannel(iteratorChannels, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder(), parsers.NewMainTrieLeafParser())
		assert.Equal(t, trie.ErrNilTrieIteratorErrChannel, err)
	})

	t.Run("nil keyBuilder", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		iteratorChannels := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    errChan.NewErrChanWrapper(),
		}
		err := tr.GetAllLeavesOnChannel(iteratorChannels, context.Background(), []byte{}, nil, parsers.NewMainTrieLeafParser())
		assert.Equal(t, trie.ErrNilKeyBuilder, err)
	})

	t.Run("nil  trieLeafParser", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		iteratorChannels := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    errChan.NewErrChanWrapper(),
		}
		err := tr.GetAllLeavesOnChannel(iteratorChannels, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder(), nil)
		assert.Equal(t, trie.ErrNilTrieLeafParser, err)
	})

	t.Run("empty trie", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		leavesChannel := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    errChan.NewErrChanWrapper(),
		}
		err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), []byte{}, keyBuilder.NewDisabledKeyBuilder(), parsers.NewMainTrieLeafParser())
		assert.Nil(t, err)
		assert.NotNil(t, leavesChannel)

		_, ok := <-leavesChannel.LeavesChan
		assert.False(t, ok)

		err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
		assert.Nil(t, err)
	})

	t.Run("should fail on getting leaves", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		_ = tr.Commit()
		rootHash, _ := tr.RootHash()

		leavesChannel := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    errChan.NewErrChanWrapper(),
		}

		expectedErr := errors.New("expected error")
		keyBuilderStub := &mock.KeyBuilderStub{}
		keyBuilderStub.GetKeyCalled = func() ([]byte, error) {
			return nil, expectedErr
		}
		keyBuilderStub.ShallowCloneCalled = func() common.KeyBuilder {
			return keyBuilderStub
		}

		err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), rootHash, keyBuilderStub, parsers.NewMainTrieLeafParser())
		assert.Nil(t, err)
		assert.NotNil(t, leavesChannel)

		recovered := make(map[string][]byte)
		for leaf := range leavesChannel.LeavesChan {
			recovered[string(leaf.Key())] = leaf.Value()
		}
		err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
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
			ErrChan:    errChan.NewErrChanWrapper(),
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
		keyBuilderStub.ShallowCloneCalled = func() common.KeyBuilder {
			return keyBuilderStub
		}

		err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), rootHash, keyBuilderStub, parsers.NewMainTrieLeafParser())
		assert.Nil(t, err)
		assert.NotNil(t, leavesChannel)

		recovered := make(map[string][]byte)
		for leaf := range leavesChannel.LeavesChan {
			recovered[string(leaf.Key())] = leaf.Value()
		}
		err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
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
			ErrChan:    errChan.NewErrChanWrapper(),
		}
		err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), rootHash, keyBuilder.NewKeyBuilder(), parsers.NewMainTrieLeafParser())
		assert.Nil(t, err)
		assert.NotNil(t, leavesChannel)

		recovered := make(map[string][]byte)
		for leaf := range leavesChannel.LeavesChan {
			recovered[string(leaf.Key())] = leaf.Value()
		}
		err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
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
			fmt.Printf("error getting proof for %v, err = %s\n", values[randNum], err.Error())
		}
		require.Nil(t, err)
		require.NotEqual(t, 0, len(proof))

		ok, err := tr.VerifyProof(rootHash, values[randNum], proof)
		if err != nil {
			fmt.Printf("error verifying proof for %v, proof = %v, err = %s\n", values[randNum], proof, err.Error())
		}
		require.Nil(t, err)
		require.True(t, ok)
	}
}

func TestPatriciaMerkleTrie_GetTrieStats(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	_ = tr.Update([]byte("dog"), []byte("reindeer"))
	_ = tr.Update([]byte("fog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Update([]byte("abch"), []byte("car"))
	_ = tr.Commit()

	rootHash, _ := tr.RootHash()
	address := "address"

	ts, ok := tr.(common.TrieStats)
	assert.True(t, ok)

	stats, err := ts.GetTrieStats(address, rootHash)
	assert.Nil(t, err)

	assert.Equal(t, uint64(3), stats.GetNumBranchNodes())
	assert.Equal(t, uint64(2), stats.GetNumExtensionNodes())
	assert.Equal(t, uint64(4), stats.GetNumLeafNodes())
	assert.Equal(t, uint64(9), stats.GetTotalNumNodes())
	assert.Equal(t, uint32(3), stats.GetMaxTrieDepth())
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

func TestPatriciaMerkleTrie_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	numOperations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numOperations)
	numFunctions := 18

	initialRootHash, _ := tr.RootHash()

	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)

			operation := idx % numFunctions
			switch operation {
			case 0:
				_, _, err := tr.Get([]byte("dog"))
				assert.Nil(t, err)
			case 1:
				err := tr.Update([]byte("doe"), []byte("alt"))
				assert.Nil(t, err)
			case 2:
				err := tr.Delete([]byte("alt"))
				assert.Nil(t, err)
			case 3:
				_, err := tr.RootHash()
				assert.Nil(t, err)
			case 4:
				err := tr.Commit()
				assert.Nil(t, err)
			case 5:
				epoch := core.OptionalUint32{
					Value:    3,
					HasValue: true,
				}
				rootHashHolder := holders.NewRootHashHolder(initialRootHash, epoch)
				_, err := tr.Recreate(rootHashHolder)
				assert.Nil(t, err)
			case 6:
				_ = tr.String()
			case 7:
				_ = tr.GetObsoleteHashes()
			case 8:
				_, err := tr.GetDirtyHashes()
				assert.Nil(t, err)
			case 9:
				_, err := tr.GetSerializedNode(initialRootHash)
				assert.Nil(t, err)
			case 10:
				size1KB := uint64(1024 * 1024)
				_, _, err := tr.GetSerializedNodes(initialRootHash, size1KB)
				assert.Nil(t, err)
			case 11:
				trieIteratorChannels := &common.TrieIteratorChannels{
					LeavesChan: make(chan core.KeyValueHolder, 1000),
					ErrChan:    errChan.NewErrChanWrapper(),
				}

				err := tr.GetAllLeavesOnChannel(
					trieIteratorChannels,
					context.Background(),
					initialRootHash,
					keyBuilder.NewKeyBuilder(),
					parsers.NewMainTrieLeafParser(),
				)
				assert.Nil(t, err)
			case 12:
				_, err := tr.GetAllHashes()
				assert.Nil(t, err)
			case 13:
				_, _, _ = tr.GetProof(initialRootHash) // this might error due to concurrent operations that change the roothash
			case 14:
				// extremely hard to compute an existing hash due to concurrent changes.
				_, _ = tr.VerifyProof([]byte("dog"), []byte("puppy"), [][]byte{[]byte("proof1")}) // this might error due to concurrent operations that change the roothash
			case 15:
				sm := tr.GetStorageManager()
				assert.NotNil(t, sm)
			case 16:
				_ = tr.GetOldRoot()
			case 17:
				trieStatsHandler := tr.(common.TrieStats)
				_, err := trieStatsHandler.GetTrieStats("address", initialRootHash)
				assert.Nil(t, err)
			default:
				assert.Fail(t, fmt.Sprintf("invalid numFunctions value %d, operation: %d", numFunctions, operation))
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestPatriciaMerkleTrie_GetSerializedNodesShouldSerializeTheCalls(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	numConcurrentCalls := int32(0)
	testTrieStorageManager := &storageManager.StorageManagerStub{
		GetCalled: func(bytes []byte) ([]byte, error) {
			newValue := atomic.AddInt32(&numConcurrentCalls, 1)
			defer atomic.AddInt32(&numConcurrentCalls, -1)

			assert.Equal(t, int32(1), newValue)

			// get takes a long time
			time.Sleep(time.Millisecond * 10)

			return bytes, nil
		},
	}

	tr, _ := trie.NewTrie(testTrieStorageManager, args.Marshalizer, args.Hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, collapseManager.NewDisabledCollapseManager())
	numGoRoutines := 100
	wg := sync.WaitGroup{}
	wg.Add(numGoRoutines)

	for i := 0; i < numGoRoutines; i++ {
		if i%2 == 0 {
			go func() {
				time.Sleep(time.Millisecond * 100)
				_, _, _ = tr.GetSerializedNodes([]byte("dog"), 1024)
				wg.Done()
			}()
		} else {
			go func() {
				time.Sleep(time.Millisecond * 100)
				_, _ = tr.GetSerializedNode([]byte("dog"))
				wg.Done()
			}()
		}
	}

	wg.Wait()
	chanClosed := make(chan struct{})
	go func() {
		_ = tr.Close()
		close(chanClosed)
	}()

	timeout := time.Second * 10
	select {
	case <-chanClosed: // ok
	case <-time.After(timeout):
		assert.Fail(t, "timeout waiting for trie to be closed")
	}
}

type dataTrie interface {
	CollectLeavesForMigration(args vmcommon.ArgsMigrateDataTrieLeaves) error
	UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion) error
}

func TestPatriciaMerkleTrie_CollectLeavesForMigration(t *testing.T) {
	t.Parallel()

	t.Run("nil root", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()

		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				assert.Fail(t, "should not have called this function")
				return false
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := tr.(dataTrie).CollectLeavesForMigration(args)
		assert.Nil(t, err)
	})

	t.Run("nil trie migrator", func(t *testing.T) {
		t.Parallel()

		tr := initTrie().(dataTrie)

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: nil,
		}
		err := tr.CollectLeavesForMigration(args)
		assert.Equal(t, errorsCommon.ErrNilTrieMigrator, err)
	})

	t.Run("data trie already migrated", func(t *testing.T) {
		t.Parallel()

		numLoadsCalled := 0
		tr := emptyTrieWithCustomEnableEpochsHandler(
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.AutoBalanceDataTriesFlag
				},
			},
		)
		dtr := tr.(dataTrie)
		_ = dtr.UpdateWithVersion([]byte("dog"), []byte("reindeer"), core.AutoBalanceEnabled)
		_ = dtr.UpdateWithVersion([]byte("ddog"), []byte("puppy"), core.AutoBalanceEnabled)
		_ = dtr.UpdateWithVersion([]byte("doe"), []byte("cat"), core.AutoBalanceEnabled)

		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				numLoadsCalled++
				return true
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.Nil(t, err)
		assert.Equal(t, 1, numLoadsCalled)
	})

	t.Run("trie partially migrated", func(t *testing.T) {
		t.Parallel()

		addLeafToMigrationQueueCalled := 0
		tr := emptyTrieWithCustomEnableEpochsHandler(
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.AutoBalanceDataTriesFlag
				},
			},
		)
		dtr := tr.(dataTrie)
		key := []byte("dog")
		value := []byte("reindeer")
		_ = dtr.UpdateWithVersion(key, value, core.NotSpecified)
		_ = dtr.UpdateWithVersion([]byte("ddog"), []byte("puppy"), core.AutoBalanceEnabled)
		_ = dtr.UpdateWithVersion([]byte("doe"), []byte("cat"), core.AutoBalanceEnabled)

		dtm := &trieMock.DataTrieMigratorStub{
			AddLeafToMigrationQueueCalled: func(leafData core.TrieData, newLeafVersion core.TrieNodeVersion) (bool, error) {
				assert.Equal(t, core.AutoBalanceEnabled, newLeafVersion)
				assert.Equal(t, key, leafData.Key)
				assert.Equal(t, value, leafData.Value)
				assert.Equal(t, core.NotSpecified, leafData.Version)
				addLeafToMigrationQueueCalled++
				return true, nil
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.Nil(t, err)
		assert.Equal(t, 1, addLeafToMigrationQueueCalled)
	})

	t.Run("not enough gas to load the whole trie", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrieWithCustomEnableEpochsHandler(
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.AutoBalanceDataTriesFlag
				},
			},
		)
		addDefaultDataToTrie(tr)

		dtr := tr.(dataTrie)
		numLoads := 0
		numAddLeafToMigrationQueueCalled := 0
		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				if numLoads < 2 {
					numLoads++
					return true
				}

				numLoads++
				return false
			},
			AddLeafToMigrationQueueCalled: func(_ core.TrieData, _ core.TrieNodeVersion) (bool, error) {
				numAddLeafToMigrationQueueCalled++
				return true, nil
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.Nil(t, err)
		assert.Equal(t, 3, numLoads)
		assert.Equal(t, 1, numAddLeafToMigrationQueueCalled)
	})

	t.Run("not enough gas to migrate the whole trie", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrieWithCustomEnableEpochsHandler(
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.AutoBalanceDataTriesFlag
				},
			},
		)
		addDefaultDataToTrie(tr)
		dtr := tr.(dataTrie)
		numLoads := 0
		numAddLeafToMigrationQueueCalled := 0
		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				numLoads++
				return true
			},
			AddLeafToMigrationQueueCalled: func(_ core.TrieData, _ core.TrieNodeVersion) (bool, error) {
				if numAddLeafToMigrationQueueCalled < 1 {
					numAddLeafToMigrationQueueCalled++
					return true, nil
				}

				numAddLeafToMigrationQueueCalled++
				return false, nil
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.Nil(t, err)
		assert.Equal(t, 5, numLoads)
		assert.Equal(t, 2, numAddLeafToMigrationQueueCalled)
	})

	t.Run("migrate to non existent version", func(t *testing.T) {
		t.Parallel()

		numLoadsCalled := 0
		numAddLeafToMigrationQueueCalled := 0
		dtr := initTrie().(dataTrie)
		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				numLoadsCalled++
				return true
			},
			AddLeafToMigrationQueueCalled: func(_ core.TrieData, _ core.TrieNodeVersion) (bool, error) {
				numAddLeafToMigrationQueueCalled++
				return true, nil
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.TrieNodeVersion(100),
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.True(t, strings.Contains(err.Error(), errorsCommon.ErrInvalidTrieNodeVersion.Error()))
		assert.Equal(t, 0, numLoadsCalled)
		assert.Equal(t, 0, numAddLeafToMigrationQueueCalled)
	})

	t.Run("migrate from non existent version", func(t *testing.T) {
		t.Parallel()

		numLoadsCalled := 0
		numAddLeafToMigrationQueueCalled := 0
		dtr := initTrie().(dataTrie)
		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				numLoadsCalled++
				return true
			},
			AddLeafToMigrationQueueCalled: func(_ core.TrieData, _ core.TrieNodeVersion) (bool, error) {
				numAddLeafToMigrationQueueCalled++
				return true, nil
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.TrieNodeVersion(100),
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.True(t, strings.Contains(err.Error(), errorsCommon.ErrInvalidTrieNodeVersion.Error()))
		assert.Equal(t, 0, numLoadsCalled)
		assert.Equal(t, 0, numAddLeafToMigrationQueueCalled)
	})

	t.Run("migrate collapsed trie", func(t *testing.T) {
		t.Parallel()

		numLoadsCalled := 0
		numAddLeafToMigrationQueueCalled := 0

		tr := emptyTrieWithCustomEnableEpochsHandler(
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.AutoBalanceDataTriesFlag
				},
			},
		)
		addDefaultDataToTrie(tr)
		_ = tr.Commit()
		rootHash, _ := tr.RootHash()
		collapsedTrie, _ := tr.Recreate(holders.NewDefaultRootHashesHolder(rootHash))
		dtr := collapsedTrie.(dataTrie)
		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				numLoadsCalled++
				return true
			},
			AddLeafToMigrationQueueCalled: func(_ core.TrieData, _ core.TrieNodeVersion) (bool, error) {
				numAddLeafToMigrationQueueCalled++
				return true, nil
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.Nil(t, err)
		assert.Equal(t, 6, numLoadsCalled)
		assert.Equal(t, 3, numAddLeafToMigrationQueueCalled)
	})

	t.Run("migrate all non migrated leaves", func(t *testing.T) {
		t.Parallel()

		numLoadsCalled := 0
		numAddLeafToMigrationQueueCalled := 0
		tr := emptyTrieWithCustomEnableEpochsHandler(
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.AutoBalanceDataTriesFlag
				},
			},
		)
		dtr := tr.(dataTrie)
		_ = dtr.UpdateWithVersion([]byte("dog"), []byte("reindeer"), core.AutoBalanceEnabled)
		_ = dtr.UpdateWithVersion([]byte("ddog"), []byte("puppy"), core.AutoBalanceEnabled)
		_ = dtr.UpdateWithVersion([]byte("doe"), []byte("cat"), core.NotSpecified)
		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				numLoadsCalled++
				return true
			},
			AddLeafToMigrationQueueCalled: func(_ core.TrieData, _ core.TrieNodeVersion) (bool, error) {
				numAddLeafToMigrationQueueCalled++
				return true, nil
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.Nil(t, err)
		assert.Equal(t, 2, numLoadsCalled)
		assert.Equal(t, 1, numAddLeafToMigrationQueueCalled)
	})

	t.Run("migrate to same version", func(t *testing.T) {
		t.Parallel()

		numLoadsCalled := 0
		numAddLeafToMigrationQueueCalled := 0
		tr := emptyTrieWithCustomEnableEpochsHandler(
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.AutoBalanceDataTriesFlag
				},
			},
		)
		dtr := tr.(dataTrie)
		_ = dtr.UpdateWithVersion([]byte("dog"), []byte("reindeer"), core.AutoBalanceEnabled)
		_ = dtr.UpdateWithVersion([]byte("ddog"), []byte("puppy"), core.AutoBalanceEnabled)
		_ = dtr.UpdateWithVersion([]byte("doe"), []byte("cat"), core.AutoBalanceEnabled)
		dtm := &trieMock.DataTrieMigratorStub{
			ConsumeStorageLoadGasCalled: func() bool {
				numLoadsCalled++
				return true
			},
			AddLeafToMigrationQueueCalled: func(_ core.TrieData, _ core.TrieNodeVersion) (bool, error) {
				numAddLeafToMigrationQueueCalled++
				return true, nil
			},
		}

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: dtm,
		}
		err := dtr.CollectLeavesForMigration(args)
		assert.Nil(t, err)
		assert.Equal(t, 1, numLoadsCalled)
		assert.Equal(t, 0, numAddLeafToMigrationQueueCalled)
	})
}

func TestPatriciaMerkleTrie_IsMigrated(t *testing.T) {
	t.Parallel()

	t.Run("nil root", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		isMigrated, err := tr.IsMigratedToLatestVersion()
		assert.True(t, isMigrated)
		assert.Nil(t, err)
	})

	t.Run("not migrated", func(t *testing.T) {
		t.Parallel()

		tsm, marshaller, hasher, _, cm := getDefaultTrieParameters()
		enableEpochs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tr, _ := trie.NewTrie(tsm, marshaller, hasher, enableEpochs, cm)

		_ = tr.Update([]byte("dog"), []byte("reindeer"))
		isMigrated, err := tr.IsMigratedToLatestVersion()
		assert.False(t, isMigrated)
		assert.Nil(t, err)
	})

	t.Run("migrated", func(t *testing.T) {
		t.Parallel()

		tsm, marshaller, hasher, _, cm := getDefaultTrieParameters()
		enableEpochs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tr, _ := trie.NewTrie(tsm, marshaller, hasher, enableEpochs, cm)

		_ = tr.UpdateWithVersion([]byte("dog"), []byte("reindeer"), core.AutoBalanceEnabled)
		isMigrated, err := tr.IsMigratedToLatestVersion()
		assert.True(t, isMigrated)
		assert.Nil(t, err)
	})
}

func TestGetNodeDataFromHash(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Update([]byte("111"), []byte("111"))
	_ = tr.Update([]byte("aaa"), []byte("aaa"))
	_ = tr.Commit()

	hashSize := 32
	keySize := 1

	rootHash, _ := tr.RootHash()
	nodeData, err := trie.GetNodeDataFromHash(rootHash, keyBuilder.NewKeyBuilder(), tr.GetStorageManager(), &marshal.GogoProtoMarshalizer{}, &testscommon.KeccakMock{})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(nodeData))

	firstChildData := nodeData[0]
	assert.Equal(t, uint(1), firstChildData.GetKeyBuilder().Size())
	assert.Equal(t, uint64(hashSize+keySize), firstChildData.Size())
	assert.False(t, firstChildData.IsLeaf())

	seconChildData := nodeData[1]
	assert.Equal(t, uint(1), seconChildData.GetKeyBuilder().Size())
	assert.Equal(t, uint64(hashSize+keySize), seconChildData.Size())
	assert.False(t, seconChildData.IsLeaf())

	thirdChildData := nodeData[2]
	assert.Equal(t, uint(1), thirdChildData.GetKeyBuilder().Size())
	assert.Equal(t, uint64(hashSize+keySize), thirdChildData.Size())
	assert.False(t, thirdChildData.IsLeaf())

}

func TestPatriciaMerkleTree_SizeInMemory(t *testing.T) {
	t.Parallel()

	db, msh, hsh, epochs, _ := getDefaultTrieParameters()
	cm, err := collapseManager.NewCollapseManager(common.TenMbSize, common.NumLeavesToCollapseSingleRun)
	assert.Nil(t, err)
	tr, _ := trie.NewTrie(db, msh, hsh, epochs, cm)

	assert.Equal(t, 0, tr.SizeInMemory())
	addDefaultDataToTrie(tr)

	assert.Equal(t, 779, tr.SizeInMemory()) // 3 leaves + 2 branch nodes + 1 extension node
	err = tr.Commit()
	assert.Nil(t, err)

	err = tr.Delete([]byte("dog"))
	assert.Nil(t, err)
	assert.Equal(t, 380, tr.SizeInMemory()) // 1 branch node + 2 leaves

	err = tr.Commit()
	assert.Nil(t, err)

	err = tr.Update([]byte("dog"), []byte("puppy"))
	assert.Nil(t, err)
	assert.Equal(t, 779, tr.SizeInMemory())

	rootHash, err := tr.RootHash()
	assert.Nil(t, err)
	newTrie, err := tr.Recreate(holders.NewDefaultRootHashesHolder(rootHash))
	assert.Nil(t, err)
	assert.Equal(t, 249, newTrie.SizeInMemory()) // only root node is in memory

	val, depth, err := newTrie.Get([]byte("dog"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("puppy"), val)
	assert.Equal(t, uint32(3), depth)
	assert.Equal(t, 654, newTrie.SizeInMemory())

	val, depth, err = newTrie.Get([]byte("dog"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("puppy"), val)
	assert.Equal(t, uint32(3), depth)
	assert.Equal(t, 654, newTrie.SizeInMemory())

	err = tr.Delete([]byte("doe")) // delete collapsed node
	assert.Nil(t, err)
	assert.Equal(t, 464, tr.SizeInMemory())
}

func TestPatriciaMerkleTree_CollapseTrie(t *testing.T) {
	t.Parallel()

	db, msh, hsh, epochs, _ := getDefaultTrieParameters()
	oneMbSize := uint64(1048576)
	cm, _ := collapseManager.NewCollapseManager(oneMbSize, common.NumLeavesToCollapseSingleRun)
	tr, _ := trie.NewTrie(db, msh, hsh, epochs, cm)

	for uint64(tr.SizeInMemory()) < oneMbSize {
		randomKey := make([]byte, 32)
		_, _ = cryptoRand.Read(randomKey)
		_ = tr.Update(randomKey, randomKey)
	}

	sizeBeforeCollapse := tr.SizeInMemory()
	numCollapsed, err := tr.GetNumCollapsedNodes()
	assert.Nil(t, err)
	assert.Equal(t, 0, numCollapsed)
	err = tr.Commit()
	assert.Nil(t, err)
	sizeAfterCollapse := tr.SizeInMemory()
	assert.Less(t, sizeAfterCollapse, sizeBeforeCollapse)
	assert.Less(t, sizeAfterCollapse, int(oneMbSize))
	numCollapsed, err = tr.GetNumCollapsedNodes()
	assert.Nil(t, err)
	assert.Equal(t, 100, numCollapsed)
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

func BenchmarkPatriciaMerkleTrie_Update(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.NewKeccak()

	nrValuesInTrie := 2000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		key := hsh.Compute(strconv.Itoa(i))
		value := append(key, []byte(strconv.Itoa(i))...)

		_ = tr.Update(key, value)
		values[i] = key
	}
	_ = tr.Commit()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < nrValuesInTrie; j++ {
			_ = tr.Update(values[j], values[j])
		}
	}
}

const commitMemoryBudget = 1 << 20

var errCommitBudgetWrite = errors.New("commit budget test: storage write failed")

type commitBudgetStore struct {
	common.BaseStorer
	writes atomic.Int64
	fail   atomic.Bool
}

func (store *commitBudgetStore) Put(key, value []byte) error {
	store.writes.Add(1)
	if store.fail.Load() {
		return errCommitBudgetWrite
	}
	return store.BaseStorer.Put(key, value)
}

type commitBudgetFixture struct {
	tr    *trie.PatriciaMerkleTrieForTests
	store *commitBudgetStore
	keys  [][]byte
	root  []byte
}

func commitBudgetValue(index, block int) []byte {
	value := bytes.Repeat([]byte{byte(block + 1)}, 128)
	binary.BigEndian.PutUint64(value, uint64(index))
	return value
}

func newCommitBudgetFixture(t *testing.T, version core.TrieNodeVersion, count int, batch uint32, largeFirst bool) *commitBudgetFixture {
	t.Helper()
	args := trie.GetDefaultTrieStorageManagerParameters()
	store := &commitBudgetStore{BaseStorer: args.MainStorer}
	args.MainStorer = store
	storage, err := trie.NewTrieStorageManager(args)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	if version == core.AutoBalanceEnabled {
		flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
	}
	// Populate persisted state independently of the memory policy being tested.
	seed, err := trie.NewTrie(storage, args.Marshalizer, args.Hasher, flags, collapseManager.NewDisabledCollapseManager())
	require.NoError(t, err)
	keys := make([][]byte, count)
	for i := range keys {
		hash := sha256.Sum256([]byte(fmt.Sprintf("account-%d", i)))
		keys[i] = hash[:]
		value := commitBudgetValue(i, 0)
		if i == 0 && largeFirst {
			value = bytes.Repeat([]byte{1}, 900*1024)
		}
		require.NoError(t, seed.UpdateWithVersion(keys[i], value, version))
	}
	require.NoError(t, seed.Commit())
	root, err := seed.RootHash()
	require.NoError(t, err)
	manager, err := collapseManager.NewCollapseManager(commitMemoryBudget, batch)
	require.NoError(t, err)
	template, err := trie.NewTrie(storage, args.Marshalizer, args.Hasher, flags, manager)
	require.NoError(t, err)
	recreated, err := template.Recreate(holders.NewDefaultRootHashesHolder(root))
	require.NoError(t, err)
	return &commitBudgetFixture{tr: recreated.(*trie.PatriciaMerkleTrieForTests), store: store, keys: keys, root: root}
}

func requireCommitBudgetValue(t *testing.T, tr common.Trie, key, expected []byte, version core.TrieNodeVersion) {
	t.Helper()
	value, _, actualVersion, err := tr.GetWithVersion(key)
	require.NoError(t, err)
	require.Equal(t, expected, value)
	require.Equal(t, uint32(version), actualVersion)
}

func TestTrieCommitMemoryBudgetAcrossBlocks(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		for _, readOnly := range []bool{false, true} {
			t.Run(fmt.Sprintf("version-%d/read-only-%t", version, readOnly), func(t *testing.T) {
				fixture := newCommitBudgetFixture(t, version, 12000, 100, false)
				tr := fixture.tr
				writes := fixture.store.writes.Load()
				for block := 1; block <= 12; block++ {
					for i := (block - 1) * 300; i < block*300; i++ {
						if readOnly {
							requireCommitBudgetValue(t, tr, fixture.keys[i], commitBudgetValue(i, 0), version)
						} else {
							require.NoError(t, tr.UpdateWithVersion(fixture.keys[i], commitBudgetValue(i, block), version))
						}
					}
					root, err := tr.RootHash()
					require.NoError(t, err)
					require.NoError(t, tr.Commit())
					after, err := tr.RootHash()
					require.NoError(t, err)
					require.Equal(t, root, after, "eviction changed the root at block %d", block)
					require.LessOrEqual(t, tr.SizeInMemory(), commitMemoryBudget, "block %d", block)
					require.LessOrEqual(t, tr.ResidentSizeForTests(), commitMemoryBudget, "resident nodes at block %d", block)
					if readOnly {
						require.Equal(t, fixture.root, after)
						require.Equal(t, writes, fixture.store.writes.Load(), "clean commit wrote nodes")
					}
				}
				root, err := tr.RootHash()
				require.NoError(t, err)
				cold, err := tr.Recreate(holders.NewDefaultRootHashesHolder(root))
				require.NoError(t, err)
				for _, i := range []int{0, 1799, 3599, 11999} {
					block := 0
					if !readOnly && i < 3600 {
						block = i/300 + 1
					}
					requireCommitBudgetValue(t, cold, fixture.keys[i], commitBudgetValue(i, block), version)
				}
			})
		}
	}
}

func TestTrieCommitMemoryBudgetKeepsCacheWhenLeafEvictionIsEnough(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		for _, readOnly := range []bool{false, true} {
			t.Run(fmt.Sprintf("version-%d/read-only-%t", version, readOnly), func(t *testing.T) {
				fixture := newCommitBudgetFixture(t, version, 2200, 1, true)
				tr := fixture.tr
				// The large leaf is least recently used. A single eviction suffices,
				// so the remaining decoded paths must survive this commit.
				requireCommitBudgetValue(t, tr, fixture.keys[0], bytes.Repeat([]byte{1}, 900*1024), version)
				for i := 1; i < len(fixture.keys); i++ {
					requireCommitBudgetValue(t, tr, fixture.keys[i], commitBudgetValue(i, 0), version)
				}
				if !readOnly {
					require.NoError(t, tr.UpdateWithVersion(fixture.keys[1], commitBudgetValue(1, 1), version))
				}
				require.Greater(t, tr.SizeInMemory(), commitMemoryBudget)
				rootNode := tr.RootNodeForTests()
				root, err := tr.RootHash()
				require.NoError(t, err)
				require.NoError(t, tr.Commit())
				require.LessOrEqual(t, tr.SizeInMemory(), commitMemoryBudget)
				require.Same(t, rootNode, tr.RootNodeForTests(), "leaf eviction unnecessarily collapsed the whole trie")
				residentSize := tr.SizeInMemory()
				requireCommitBudgetValue(t, tr, fixture.keys[2199], commitBudgetValue(2199, 0), version)
				require.Equal(t, residentSize, tr.SizeInMemory(), "hot leaf had to be reloaded")
				after, err := tr.RootHash()
				require.NoError(t, err)
				require.Equal(t, root, after)
				requireCommitBudgetValue(t, tr, fixture.keys[0], bytes.Repeat([]byte{1}, 900*1024), version)
			})
		}
	}
}

func TestTrieCommitMemoryBudgetKeepsCacheBelowLimit(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		for _, readOnly := range []bool{false, true} {
			t.Run(fmt.Sprintf("version-%d/read-only-%t", version, readOnly), func(t *testing.T) {
				fixture := newCommitBudgetFixture(t, version, 128, 100, false)
				tr := fixture.tr
				for i, key := range fixture.keys {
					requireCommitBudgetValue(t, tr, key, commitBudgetValue(i, 0), version)
				}
				if !readOnly {
					require.NoError(t, tr.UpdateWithVersion(fixture.keys[0], commitBudgetValue(0, 1), version))
				}
				rootNode := tr.RootNodeForTests()
				size := tr.SizeInMemory()
				resident := tr.ResidentSizeForTests()
				writes := fixture.store.writes.Load()
				require.Less(t, size, commitMemoryBudget)
				require.NoError(t, tr.Commit())
				require.Same(t, rootNode, tr.RootNodeForTests())
				require.Equal(t, size, tr.SizeInMemory())
				require.Equal(t, resident, tr.ResidentSizeForTests())
				if readOnly {
					require.Equal(t, writes, fixture.store.writes.Load())
				}
			})
		}
	}
}

func TestTrieCommitMemoryBudgetWaitsForSuccessfulPersistence(t *testing.T) {
	fixture := newCommitBudgetFixture(t, core.NotSpecified, 12000, 100, false)
	tr := fixture.tr
	for i := 0; i < 6000; i++ {
		require.NoError(t, tr.Update(fixture.keys[i], commitBudgetValue(i, 1)))
	}
	require.Greater(t, tr.SizeInMemory(), commitMemoryBudget)
	rootNode := tr.RootNodeForTests()
	fixture.store.fail.Store(true)
	require.ErrorIs(t, tr.Commit(), errCommitBudgetWrite)
	require.Same(t, rootNode, tr.RootNodeForTests())
	require.Greater(t, tr.SizeInMemory(), commitMemoryBudget)
	fixture.store.fail.Store(false)
	require.NoError(t, tr.Commit())
	require.LessOrEqual(t, tr.SizeInMemory(), commitMemoryBudget)
	root, err := tr.RootHash()
	require.NoError(t, err)
	cold, err := tr.Recreate(holders.NewDefaultRootHashesHolder(root))
	require.NoError(t, err)
	requireCommitBudgetValue(t, cold, fixture.keys[5999], commitBudgetValue(5999, 1), core.NotSpecified)
}

func TestTrieCommitMemoryBudgetRetainsOversizedRootLeaf(t *testing.T) {
	fixture := newCommitBudgetFixture(t, core.NotSpecified, 1, 100, false)
	value := bytes.Repeat([]byte{7}, commitMemoryBudget+1)
	require.NoError(t, fixture.tr.Update(fixture.keys[0], value))
	root, err := fixture.tr.RootHash()
	require.NoError(t, err)
	// A root leaf cannot be collapsed further. One bounded pass must return
	// successfully without discarding its value or retrying until under budget.
	for i := 0; i < 2; i++ {
		require.NoError(t, fixture.tr.Commit())
		after, rootErr := fixture.tr.RootHash()
		require.NoError(t, rootErr)
		require.Equal(t, root, after)
		requireCommitBudgetValue(t, fixture.tr, fixture.keys[0], value, core.NotSpecified)
	}
}

func newTrieForWholeCollapse(t *testing.T) (*trie.PatriciaMerkleTrieForTests, common.TrieCollapseManager) {
	t.Helper()
	args := trie.GetDefaultTrieStorageManagerParameters()
	storage, err := trie.NewTrieStorageManager(args)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	manager, err := collapseManager.NewCollapseManager(commitMemoryBudget, 1)
	require.NoError(t, err)
	tr, err := trie.NewTrie(storage, args.Marshalizer, args.Hasher, enableEpochsHandlerMock.NewEnableEpochsHandlerStub(), manager)
	require.NoError(t, err)
	return tr, manager
}

func growTrieForWholeCollapse(tr *trie.PatriciaMerkleTrieForTests, round int) error {
	// Evicting a single 600 KiB leaf still leaves the trie above its 1 MiB budget.
	for j := 1; j <= 3; j++ {
		payload := bytes.Repeat([]byte{byte(round*3 + j)}, 600<<10)
		if err := tr.Update([]byte{byte(j)}, payload); err != nil {
			return err
		}
	}
	return nil
}

func TestWholeTrieCollapseKeepsManager(t *testing.T) {
	tr, manager := newTrieForWholeCollapse(t)
	for round := range 2 {
		require.NoError(t, growTrieForWholeCollapse(tr, round))
		require.Greater(t, tr.SizeInMemory(), commitMemoryBudget)
		root, err := tr.RootHash()
		require.NoError(t, err)
		require.NoError(t, tr.Commit())
		require.Same(t, manager, tr.CollapseManagerForTests())
		require.True(t, tr.RootIsCollapsedForTests())
		require.Equal(t, tr.RootSizeForTests(), tr.SizeInMemory())
		committed, err := tr.RootHash()
		require.NoError(t, err)
		require.Equal(t, root, committed)
	}

	recreated, err := tr.Recreate(holders.NewDefaultRootHashesHolder(nil))
	require.NoError(t, err)
	clone := recreated.(*trie.PatriciaMerkleTrieForTests).CollapseManagerForTests()
	require.NotSame(t, manager, clone)
	require.Zero(t, clone.GetSizeInMemory())
	clone.AddSizeInMemory(commitMemoryBudget)
	require.False(t, clone.ShouldCollapseTrie())
	clone.AddSizeInMemory(1)
	require.True(t, clone.ShouldCollapseTrie())
	require.Equal(t, tr.RootSizeForTests(), tr.SizeInMemory())
}

func TestRecreateConcurrentlyWithWholeTrieCollapse(t *testing.T) {
	for _, mode := range []string{"empty", "stored root", "leaves", "stats", "serialized nodes"} {
		t.Run(mode, func(t *testing.T) {
			tr, _ := newTrieForWholeCollapse(t)
			key, value := []byte{9}, []byte("stable value")
			require.NoError(t, tr.Update(key, value))
			require.NoError(t, tr.Commit())
			root, err := tr.RootHash()
			require.NoError(t, err)

			var wg sync.WaitGroup
			start, done := make(chan struct{}), make(chan struct{})
			failure := make(chan error, 2)
			wg.Go(func() {
				defer close(done)
				<-start
				for round := range 4 {
					if errGrow := growTrieForWholeCollapse(tr, round); errGrow != nil {
						failure <- errGrow
						return
					}
					if errCommit := tr.Commit(); errCommit != nil {
						failure <- errCommit
						return
					}
					if !tr.RootIsCollapsedForTests() || tr.SizeInMemory() != tr.RootSizeForTests() {
						failure <- fmt.Errorf("commit %d did not leave only the collapsed root", round)
						return
					}
				}
			})
			wg.Go(func() {
				<-start
				for {
					if errRead := readTrieDuringCollapse(tr, mode, root, key, value); errRead != nil {
						failure <- errRead
						return
					}
					select {
					case <-done:
						return
					default:
					}
				}
			})
			close(start)
			wg.Wait()
			close(failure)
			for errOperation := range failure {
				require.NoError(t, errOperation)
			}
		})
	}
}

func readTrieDuringCollapse(tr *trie.PatriciaMerkleTrieForTests, mode string, root, key, value []byte) error {
	switch mode {
	case "empty", "stored root":
		selected, expected := root, value
		if mode == "empty" {
			selected, expected = nil, nil
		}
		recreated, err := tr.Recreate(holders.NewDefaultRootHashesHolder(selected))
		if err != nil {
			return err
		}
		got, _, err := recreated.Get(key)
		if err != nil {
			return err
		}
		if !bytes.Equal(got, expected) {
			return fmt.Errorf("recreated trie returned %q, expected %q", got, expected)
		}
		return nil
	case "leaves":
		channels := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, 1),
			ErrChan:    errChan.NewErrChanWrapper(),
		}
		err := tr.GetAllLeavesOnChannel(channels, context.Background(), root, keyBuilder.NewKeyBuilder(), parsers.NewMainTrieLeafParser())
		if err != nil {
			return err
		}
		var matches int
		for leaf := range channels.LeavesChan {
			if leaf != nil && bytes.Equal(leaf.Key(), key) && bytes.Equal(leaf.Value(), value) {
				matches++
			}
		}
		if matches != 1 {
			return fmt.Errorf("leaf iteration returned %d matches, expected one", matches)
		}
		return channels.ErrChan.ReadFromChanNonBlocking()
	case "stats":
		_, err := tr.GetTrieStats("account", root)
		return err
	default:
		_, _, err := tr.GetSerializedNodes(root, 1024)
		return err
	}
}

func TestTrie_GetWithVersionReportsGrowthOnlyWhenNodesLoad(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		t.Run(fmt.Sprintf("version-%d", version), func(t *testing.T) {
			args := trie.GetDefaultTrieStorageManagerParameters()
			store := &trie.TrieReadFailureStore{BaseStorer: args.MainStorer}
			args.MainStorer = store
			storage, err := trie.NewTrieStorageManager(args)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, storage.Close()) })
			manager, err := collapseManager.NewCollapseManager(1<<20, 100)
			require.NoError(t, err)
			seed, err := trie.NewTrie(storage, args.Marshalizer, args.Hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, manager)
			require.NoError(t, err)

			key, sibling := []byte{0x01, 0x11}, []byte{0x02, 0x11}
			value := []byte("leaf value")
			require.NoError(t, seed.UpdateWithVersion(key, value, version))
			require.NoError(t, seed.UpdateWithVersion(sibling, []byte("sibling value"), version))
			require.NoError(t, seed.Commit())
			root, err := seed.RootHash()
			require.NoError(t, err)
			// The root extension can load its branch even when the leaf read fails.
			proof, _, err := seed.GetProof(key)
			require.NoError(t, err)
			require.NotEmpty(t, proof)
			leafHash := args.Hasher.Compute(string(proof[len(proof)-1]))
			missing := []byte{0x21, 0x11}

			for _, result := range []string{"hit", "miss", "failure"} {
				t.Run(result, func(t *testing.T) {
					recreated, recreateErr := seed.Recreate(holders.NewDefaultRootHashesHolder(root))
					require.NoError(t, recreateErr)
					cold := recreated.(*trie.PatriciaMerkleTrieForTests)
					initialSize := cold.SizeInMemory()
					lookupKey := key
					switch result {
					case "miss":
						lookupKey = missing
					case "failure":
						store.FailKey = leafHash
					}
					actual, depth, storedVersion, size, readErr := cold.GetWithVersionAndSizeOnGrowth(lookupKey)
					switch result {
					case "failure":
						cause := core.UnwrapGetNodeFromDBErr(readErr)
						require.NotNil(t, cause)
						require.Equal(t, leafHash, cause.GetKey())
						require.Empty(t, actual)
					case "miss":
						require.NoError(t, readErr)
						require.Empty(t, actual)
					default:
						require.NoError(t, readErr)
						require.Equal(t, value, actual)
						require.EqualValues(t, version, storedVersion)
					}
					require.Greater(t, size, initialSize)
					require.Equal(t, cold.ResidentSizeForTests(), size)

					warmValue, warmDepth, warmVersion, warmSize, warmErr := cold.GetWithVersionAndSizeOnGrowth(lookupKey)
					require.NoError(t, warmErr)
					if result == "failure" {
						require.Equal(t, value, warmValue, "retry should load the repaired leaf")
						require.EqualValues(t, version, warmVersion)
						require.Greater(t, warmSize, size)
					} else {
						require.Equal(t, actual, warmValue)
						require.Equal(t, depth, warmDepth)
						require.Equal(t, storedVersion, warmVersion)
						require.Zero(t, warmSize)
					}
					_, _, _, warmSize, warmErr = cold.GetWithVersionAndSizeOnGrowth(lookupKey)
					require.NoError(t, warmErr)
					require.Zero(t, warmSize)
				})
			}
		})
	}
}

func TestTrieMutationSizeAccounting_CommitUsesModeledSize(t *testing.T) {
	const margin = 8
	t.Run("new branch slot crosses budget", func(t *testing.T) {
		tr := trie.NewTrieForSizeAccounting(t, commitMemoryBudget)
		keys := [][]byte{{0x01}, {0x02}, {0x03}}
		value := []byte("value")
		for _, key := range keys {
			require.NoError(t, tr.Update(key, value))
		}
		size := len(value) + commitMemoryBudget + margin - tr.ResidentSizeForTests()
		large := bytes.Repeat([]byte{3}, size)
		require.NoError(t, tr.Update(keys[2], large))
		require.Equal(t, commitMemoryBudget+margin, tr.ResidentSizeForTests())
		root, err := tr.RootHash()
		require.NoError(t, err)
		require.NoError(t, tr.Commit())
		assert.LessOrEqual(t, tr.ResidentSizeForTests(), commitMemoryBudget)
		assert.Equal(t, tr.ResidentSizeForTests(), tr.SizeInMemory())
		actualRoot, err := tr.RootHash()
		require.NoError(t, err)
		require.Equal(t, root, actualRoot)
		reopened, err := tr.Recreate(holders.NewDefaultRootHashesHolder(root))
		require.NoError(t, err)
		requireCommitBudgetValue(t, reopened, keys[0], value, core.NotSpecified)
		requireCommitBudgetValue(t, reopened, keys[1], value, core.NotSpecified)
		requireCommitBudgetValue(t, reopened, keys[2], large, core.NotSpecified)
	})

	t.Run("extension splits and merges remain below budget", func(t *testing.T) {
		tr := trie.NewTrieForSizeAccounting(t, commitMemoryBudget)
		// The first two keys share three nibbles. The third shares only
		// the first nibble, so the split retains a shorter extension.
		first, second, temporary := []byte{0x43, 0x21}, []byte{0x53, 0x21}, []byte{0x43, 0x51}
		const leafSize = 500 << 10
		secondValue := bytes.Repeat([]byte{2}, leafSize)
		require.NoError(t, tr.Update(first, []byte("a")))
		require.NoError(t, tr.Update(second, secondValue))
		size := 1 + commitMemoryBudget - margin - tr.ResidentSizeForTests()
		firstValue := bytes.Repeat([]byte{1}, size)
		require.NoError(t, tr.Update(first, firstValue))
		require.NoError(t, tr.Commit())
		root, err := tr.RootHash()
		require.NoError(t, err)
		fullSize := tr.ResidentSizeForTests()
		require.Equal(t, commitMemoryBudget-margin, fullSize)

		for range margin + 2 {
			require.NoError(t, tr.Update(temporary, []byte("temporary")))
			require.NoError(t, tr.Delete(temporary))
		}
		require.Equal(t, fullSize, tr.ResidentSizeForTests())
		require.NoError(t, tr.Commit())
		assert.Equal(t, fullSize, tr.ResidentSizeForTests())
		assert.Equal(t, fullSize, tr.SizeInMemory())
		actualRoot, err := tr.RootHash()
		require.NoError(t, err)
		require.Equal(t, root, actualRoot)
		requireCommitBudgetValue(t, tr, first, firstValue, core.NotSpecified)
		requireCommitBudgetValue(t, tr, second, secondValue, core.NotSpecified)
	})
}

func TestTrieMutationSizeAccounting_MixedOperations(t *testing.T) {
	for _, phase := range []string{"fresh", "recreated"} {
		t.Run(phase, func(t *testing.T) {
			tr := trie.NewTrieForSizeAccounting(t, commitMemoryBudget)
			random := rand.New(rand.NewSource(51))
			if phase == "recreated" {
				runTrieSizeAccountingOperations(t, tr, random, 1000, false)
				require.NoError(t, tr.Commit())
				root, err := tr.RootHash()
				require.NoError(t, err)
				reopened, err := tr.Recreate(holders.NewDefaultRootHashesHolder(root))
				require.NoError(t, err)
				tr = reopened.(*trie.PatriciaMerkleTrieForTests)
				require.Equal(t, tr.RootSizeForTests(), tr.SizeInMemory())
			}
			runTrieSizeAccountingOperations(t, tr, random, 2000, true)
		})
	}
}

func runTrieSizeAccountingOperations(t *testing.T, tr *trie.PatriciaMerkleTrieForTests, random *rand.Rand, count int, compareSize bool) {
	t.Helper()
	for i := range count {
		key := binary.LittleEndian.AppendUint16(nil, uint16(random.Intn(400)))
		switch op := random.Intn(100); {
		case op < 45:
			version := core.NotSpecified
			if random.Intn(2) == 1 {
				version = core.AutoBalanceEnabled
			}
			value := bytes.Repeat([]byte{byte(i)}, 100+random.Intn(100))
			require.NoError(t, tr.UpdateWithVersion(key, value, version))
		case op < 80:
			require.NoError(t, tr.Delete(key))
		case op < 95:
			_, _, err := tr.Get(key)
			require.NoError(t, err)
		default:
			require.NoError(t, tr.Commit())
		}
		if compareSize {
			require.Equal(t, tr.ResidentSizeForTests(), tr.SizeInMemory(), "operation %d for key %x", i, key)
		}
	}
}
