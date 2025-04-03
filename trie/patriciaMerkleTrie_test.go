package trie_test

import (
	"context"
	cryptoRand "crypto/rand"
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
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	errorsCommon "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state/hashesCollector"
	"github.com/multiversx/mx-chain-go/state/parsers"
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
	args := getDefaultTrieParameters()
	args.EnableEpochsHandler = handler
	tr, _ := trie.NewTrie(args)
	return tr
}

func getDefaultTrieParameters() trie.TrieArgs {
	args := trie.GetDefaultTrieStorageManagerParameters()
	trieStorageManager, _ := trie.NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(1)
	th, _ := throttler.NewNumGoRoutinesThrottler(10)

	return trie.TrieArgs{
		TrieStorage:          trieStorageManager,
		Marshalizer:          args.Marshalizer,
		Hasher:               args.Hasher,
		EnableEpochsHandler:  &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		MaxTrieLevelInMemory: maxTrieLevelInMemory,
		Throttler:            th,
	}
}

func initTrieMultipleValues(nr int) (common.Trie, [][]byte) {
	tr := emptyTrie()

	var values [][]byte
	hsh := keccak.NewKeccak()

	for i := 0; i < nr; i++ {
		values = append(values, hsh.Compute(fmt.Sprint(i)))
		tr.Update(values[i], values[i])
	}

	return tr, values
}

func initTrie() common.Trie {
	tr := emptyTrie()
	addDefaultDataToTrie(tr)
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

	return tr
}

func addDefaultDataToTrie(tr common.Trie) {
	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("ddog"), []byte("cat"))
}

func TestNewTrieWithNilTrieStorage(t *testing.T) {
	t.Parallel()

	args := getDefaultTrieParameters()
	args.TrieStorage = nil
	tr, err := trie.NewTrie(args)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilTrieStorage, err)
}

func TestNewTrieWithNilMarshalizer(t *testing.T) {
	t.Parallel()

	args := getDefaultTrieParameters()
	args.Marshalizer = nil
	tr, err := trie.NewTrie(args)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieWithNilHasher(t *testing.T) {
	t.Parallel()

	args := getDefaultTrieParameters()
	args.Hasher = nil
	tr, err := trie.NewTrie(args)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieWithNilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	args := getDefaultTrieParameters()
	args.EnableEpochsHandler = nil
	tr, err := trie.NewTrie(args)

	assert.Nil(t, tr)
	assert.Equal(t, errorsCommon.ErrNilEnableEpochsHandler, err)
}

func TestNewTrieWithInvalidMaxTrieLevelInMemory(t *testing.T) {
	t.Parallel()

	args := getDefaultTrieParameters()
	args.MaxTrieLevelInMemory = 0
	tr, err := trie.NewTrie(args)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrInvalidLevelValue, err)
}

func TestNewTrieWithNilThrottler(t *testing.T) {
	t.Parallel()

	args := getDefaultTrieParameters()
	args.Throttler = nil
	tr, err := trie.NewTrie(args)

	assert.Nil(t, tr)
	assert.Equal(t, trie.ErrNilThrottler, err)
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
	tr.Update([]byte("dog"), newVal)

	val, _, _ := tr.Get([]byte("dog"))
	assert.Equal(t, newVal, val)
}

func TestPatriciaMerkleTree_UpdateEmptyVal(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	var empty []byte

	tr.Update([]byte("doe"), []byte{})

	v, _, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_UpdateNotExisting(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	tr.Update([]byte("does"), []byte("this"))

	v, _, _ := tr.Get([]byte("does"))
	assert.Equal(t, []byte("this"), v)
}

func TestPatriciaMerkleTree_Delete(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	var empty []byte

	tr.Delete([]byte("doe"))

	v, _, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_DeleteEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	tr.Delete([]byte("dog"))
	err := tr.Commit(hashesCollector.NewDisabledHashesCollector())
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

	t.Run("root hash consistency for deterministic trie", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		root1, err := tr.RootHash()
		assert.Nil(t, err)

		tr.Update([]byte("dodge"), []byte("viper"))
		root2, err := tr.RootHash()
		assert.Nil(t, err)

		tr.Delete([]byte("dodge"))
		root3, err := tr.RootHash()
		assert.Nil(t, err)

		assert.Equal(t, root1, root3)
		assert.NotEqual(t, root1, root2)
	})
	t.Run("root hash consistency for non-deterministic trie", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		numValues := 1000
		for i := 0; i < numValues; i++ {
			tr.Update(generateRandomByteArray(32), generateRandomByteArray(32))
		}
		originalRootHash, err := tr.RootHash()
		assert.Nil(t, err)
		assert.False(t, common.IsEmptyTrie(originalRootHash))

		newKeys := make([][]byte, numValues)
		for i := 0; i < numValues; i++ {
			newKeys[i] = generateRandomByteArray(32)
			tr.Update(newKeys[i], generateRandomByteArray(32))
		}
		newRootHash, err := tr.RootHash()
		assert.Nil(t, err)
		assert.False(t, common.IsEmptyTrie(newRootHash))
		assert.NotEqual(t, originalRootHash, newRootHash)

		for i := 0; i < numValues; i++ {
			tr.Delete(newKeys[i])
		}
		rootHashAfterDelete, err := tr.RootHash()
		assert.Nil(t, err)
		assert.Equal(t, originalRootHash, rootHashAfterDelete)
	})
}
func generateRandomByteArray(size int) []byte {
	r := make([]byte, size)
	_, _ = cryptoRand.Read(r)
	return r
}

func TestPatriciaMerkleTrie_UpdateAndGetConcurrently(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()
	nrInserts := 100
	wg := &sync.WaitGroup{}
	wg.Add(nrInserts)

	for i := 0; i < nrInserts; i++ {
		go func(index int) {
			tr.Update([]byte(strconv.Itoa(index)), []byte(strconv.Itoa(index)))

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
			tr.Delete([]byte(strconv.Itoa(index)))
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

	err := tr.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitCollapsesTrieOk(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	tr.Update([]byte("zebra"), []byte("zebra"))
	tr.Update([]byte("doggo"), []byte("doggo"))
	tr.Update([]byte("doggless"), []byte("doggless"))

	err := tr.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitAfterCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	err := tr.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitEmptyRoot(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	err := tr.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_GetAfterCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	err := tr.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)

	val, _, err := tr.Get([]byte("dog"))
	assert.Equal(t, []byte("puppy"), val)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_InsertAfterCommit(t *testing.T) {
	t.Parallel()

	tr1 := initTrie()
	tr2 := initTrie()

	err := tr1.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)

	tr1.Update([]byte("doge"), []byte("coin"))
	tr2.Update([]byte("doge"), []byte("coin"))

	root1, _ := tr1.RootHash()
	root2, _ := tr2.RootHash()

	assert.Equal(t, root2, root1)
}

func TestPatriciaMerkleTree_DeleteAfterCommit(t *testing.T) {
	t.Parallel()

	tr1 := initTrie()
	tr2 := initTrie()

	err := tr1.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)

	tr1.Delete([]byte("dogglesworth"))
	tr2.Delete([]byte("dogglesworth"))

	root1, _ := tr1.RootHash()
	root2, _ := tr2.RootHash()

	assert.Equal(t, root2, root1)
}

func TestPatriciaMerkleTree_DeleteNotPresent(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	err := tr.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)

	tr.Delete([]byte("adog"))
	err = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)
}

func TestPatriciaMerkleTrie_Recreate(t *testing.T) {
	t.Parallel()

	t.Run("nil options", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()

		newTr, err := tr.Recreate(nil, "")
		assert.Nil(t, newTr)
		assert.Equal(t, trie.ErrNilRootHashHolder, err)
	})

	t.Run("no epoch data", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		rootHash, _ := tr.RootHash()
		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

		rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
		newTr, err := tr.Recreate(rootHashHolder, "")
		assert.Nil(t, err)

		assert.True(t, trie.IsBaseTrieStorageManager(newTr.GetStorageManager()))
	})

	t.Run("with epoch data", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		rootHash, _ := tr.RootHash()
		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

		optionalUint32 := core.OptionalUint32{
			Value:    5,
			HasValue: true,
		}
		rootHashHolder := holders.NewRootHashHolder(rootHash, optionalUint32)
		newTr, err := tr.Recreate(rootHashHolder, "")
		assert.Nil(t, err)

		assert.True(t, trie.IsTrieStorageManagerInEpoch(newTr.GetStorageManager()))
	})
}

func TestPatriciaMerkleTrie_RecreateWithInvalidRootHash(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	newTr, err := tr.Recreate(holders.NewDefaultRootHashesHolder([]byte{}), "")
	assert.Nil(t, err)
	root, _ := newTr.RootHash()
	assert.Equal(t, emptyTrieHash, root)
}

func TestPatriciaMerkleTrie_GetSerializedNodes(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
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
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()

	maxBuffToSend := uint64(150)
	expectedNodes := 2
	serializedNodes, _, err := tr.GetSerializedNodes(rootHash, maxBuffToSend)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodes, len(serializedNodes))
}

type trieWithToString interface {
	common.Trie
	ToString() string
}

func TestPatriciaMerkleTrie_String(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	str := tr.(trieWithToString).ToString()
	assert.NotEqual(t, 0, len(str))

	tr = emptyTrie()
	str = tr.(trieWithToString).ToString()
	assert.Equal(t, "*** EMPTY TRIE ***\n", str)
}

func TestPatriciaMerkleTree_reduceBranchNodeReturnsOldHashesCorrectly(t *testing.T) {
	t.Parallel()

	key1 := []byte("ABC")
	key2 := []byte("ABD")
	val1 := []byte("val1")
	val2 := []byte("val2")

	tr := emptyTrie()
	tr.Update(key1, val1)
	tr.Update(key2, val2)
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

	tr.Update(key1, nil)
	tr.Update(key1, val1)
	hc := hashesCollector.NewDataTrieHashesCollector()
	_ = tr.Commit(hc)

	_, oldHashes, newHashes := hc.GetCollectedData()
	assert.Equal(t, len(oldHashes), len(newHashes))
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
		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
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
		keyBuilderStub.CloneCalled = func() common.KeyBuilder {
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
		tr.Update([]byte("doe"), []byte("reindeer"))
		tr.Update([]byte("dog"), []byte("puppy"))
		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
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
		keyBuilderStub.CloneCalled = func() common.KeyBuilder {
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
		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
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
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()

	proof, value, err := tr.GetProof([]byte("dog"), rootHash)
	assert.Nil(t, err)
	assert.Equal(t, []byte("puppy"), value)
	ok, _ := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
}

func TestPatriciaMerkleTree_ProveCollapsedTrie(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()

	proof, _, err := tr.GetProof([]byte("dog"), rootHash)
	assert.Nil(t, err)
	ok, _ := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
}

func TestPatriciaMerkleTree_ProveOnEmptyTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	proof, _, err := tr.GetProof([]byte("dog"), emptyTrieHash)
	assert.Nil(t, proof)
	assert.Equal(t, trie.ErrNilNode, err)
}

func TestPatriciaMerkleTree_VerifyProof(t *testing.T) {
	t.Parallel()

	tr, val := initTrieMultipleValues(50)
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()

	for i := range val {
		proof, _, _ := tr.GetProof(val[i], rootHash)

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

	tr.Update([]byte("dog"), []byte("cat"))
	tr.Update([]byte("zebra"), []byte("horse"))
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()

	proof, _, _ := tr.GetProof([]byte("dog"), rootHash)
	ok, err := tr.VerifyProof(rootHash, []byte("dog"), proof)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTrie_VerifyProofExtensionNodeWantHashShouldWork(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	tr.Update([]byte("dog"), []byte("cat"))
	tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()

	proof, _, _ := tr.GetProof([]byte("dog"), rootHash)
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

	tr1.Update([]byte("doe"), []byte("reindeer"))
	tr1.Update([]byte("dog"), []byte("puppy"))
	tr1.Update([]byte("dogglesworth"), []byte("cat"))

	tr2.Update([]byte("doe"), []byte("reindeer"))
	tr2.Update([]byte("dog"), []byte("puppy"))
	tr2.Update([]byte("dogglesworth"), []byte("caterpillar"))
	_ = tr2.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash2, _ := tr2.RootHash()
	rootHash, _ := tr1.RootHash()

	proof, _, _ := tr2.GetProof([]byte("dogglesworth"), rootHash2)
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
		tr.Update(values[i], values[i])
	}

	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()
	for i := 0; i < numRuns; i++ {
		randNum := rand.Intn(nrLeaves)
		proof, _, err := tr.GetProof(values[randNum], rootHash)
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
	fmt.Println(tr.(trieWithToString).ToString())
	for _, val := range values {
		fmt.Println(val)
	}
}

func TestPatriciaMerkleTrie_GetTrieStats(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()

	tr.Update([]byte("dog"), []byte("reindeer"))
	tr.Update([]byte("fog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

	rootHash, _ := tr.RootHash()
	address := "address"

	ts, ok := tr.(common.TrieStats)
	assert.True(t, ok)

	stats, err := ts.GetTrieStats(address, rootHash)
	assert.Nil(t, err)

	assert.Equal(t, uint64(2), stats.GetNumBranchNodes())
	assert.Equal(t, uint64(1), stats.GetNumExtensionNodes())
	assert.Equal(t, uint64(3), stats.GetNumLeafNodes())
	assert.Equal(t, uint64(6), stats.GetTotalNumNodes())
	assert.Equal(t, uint32(3), stats.GetMaxTrieDepth())
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
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	numOperations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numOperations)
	numFunctions := 13

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
				tr.Update([]byte("doe"), []byte("alt"))
			case 2:
				tr.Delete([]byte("alt"))
			case 3:
				_, err := tr.RootHash()
				assert.Nil(t, err)
			case 4:
				err := tr.Commit(hashesCollector.NewDisabledHashesCollector())
				assert.Nil(t, err)
			case 5:
				epoch := core.OptionalUint32{
					Value:    3,
					HasValue: true,
				}
				rootHashHolder := holders.NewRootHashHolder(initialRootHash, epoch)
				_, err := tr.Recreate(rootHashHolder, "")
				assert.Nil(t, err)
			case 6:
				_, err := tr.GetSerializedNode(initialRootHash)
				assert.Nil(t, err)
			case 7:
				size1KB := uint64(1024 * 1024)
				_, _, err := tr.GetSerializedNodes(initialRootHash, size1KB)
				assert.Nil(t, err)
			case 8:
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
			case 9:
				_, _, _ = tr.GetProof(initialRootHash, initialRootHash) // this might error due to concurrent operations that change the roothash
			case 10:
				// extremely hard to compute an existing hash due to concurrent changes.
				_, _ = tr.VerifyProof([]byte("dog"), []byte("puppy"), [][]byte{[]byte("proof1")}) // this might error due to concurrent operations that change the roothash
			case 11:
				sm := tr.GetStorageManager()
				assert.NotNil(t, sm)
			case 12:
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

	trieArgs := getDefaultTrieParameters()
	trieArgs.TrieStorage = testTrieStorageManager
	trieArgs.MaxTrieLevelInMemory = 5

	tr, _ := trie.NewTrie(trieArgs)
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
	UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion)
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
		dtr.UpdateWithVersion([]byte("dog"), []byte("reindeer"), core.AutoBalanceEnabled)
		dtr.UpdateWithVersion([]byte("ddog"), []byte("puppy"), core.AutoBalanceEnabled)
		dtr.UpdateWithVersion([]byte("doe"), []byte("cat"), core.AutoBalanceEnabled)
		trie.ExecuteUpdatesFromBatch(tr)

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
		dtr.UpdateWithVersion(key, value, core.NotSpecified)
		dtr.UpdateWithVersion([]byte("ddog"), []byte("puppy"), core.AutoBalanceEnabled)
		dtr.UpdateWithVersion([]byte("doe"), []byte("cat"), core.AutoBalanceEnabled)
		trie.ExecuteUpdatesFromBatch(tr)

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
		trie.ExecuteUpdatesFromBatch(tr)

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
		trie.ExecuteUpdatesFromBatch(tr)

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
		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
		rootHash, _ := tr.RootHash()
		collapsedTrie, _ := tr.Recreate(holders.NewDefaultRootHashesHolder(rootHash), "")
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
		dtr.UpdateWithVersion([]byte("dog"), []byte("reindeer"), core.AutoBalanceEnabled)
		dtr.UpdateWithVersion([]byte("ddog"), []byte("puppy"), core.AutoBalanceEnabled)
		dtr.UpdateWithVersion([]byte("doe"), []byte("cat"), core.NotSpecified)
		trie.ExecuteUpdatesFromBatch(tr)

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
		dtr.UpdateWithVersion([]byte("dog"), []byte("reindeer"), core.AutoBalanceEnabled)
		dtr.UpdateWithVersion([]byte("ddog"), []byte("puppy"), core.AutoBalanceEnabled)
		dtr.UpdateWithVersion([]byte("doe"), []byte("cat"), core.AutoBalanceEnabled)
		trie.ExecuteUpdatesFromBatch(tr)

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

		trieArgs := getDefaultTrieParameters()
		enableEpochs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		trieArgs.EnableEpochsHandler = enableEpochs
		tr, _ := trie.NewTrie(trieArgs)

		tr.Update([]byte("dog"), []byte("reindeer"))
		trie.ExecuteUpdatesFromBatch(tr)
		isMigrated, err := tr.IsMigratedToLatestVersion()
		assert.False(t, isMigrated)
		assert.Nil(t, err)
	})

	t.Run("migrated", func(t *testing.T) {
		t.Parallel()

		trieArgs := getDefaultTrieParameters()
		enableEpochs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		trieArgs.EnableEpochsHandler = enableEpochs
		tr, _ := trie.NewTrie(trieArgs)

		tr.UpdateWithVersion([]byte("dog"), []byte("reindeer"), core.AutoBalanceEnabled)
		isMigrated, err := tr.IsMigratedToLatestVersion()
		assert.True(t, isMigrated)
		assert.Nil(t, err)
	})
}

func TestPatriciaMerkleTrie_InsertOneValInNilTrie(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()
	key := []byte("dog")
	value := []byte("cat")
	tr.Update(key, value)
	trie.ExecuteUpdatesFromBatch(tr)

	val, depth, err := tr.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, val)
	assert.Equal(t, uint32(0), depth)
}

func TestPatriciaMerkleTrie_AddBatchedDataToTrie(t *testing.T) {
	t.Parallel()

	t.Run("add different data to batch concurrently", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		numOperations := 1000

		wg := sync.WaitGroup{}
		wg.Add(numOperations)
		for i := 0; i < numOperations; i++ {

			if i%2 != 0 {
				go func(idx int) {
					tr.Update([]byte("dog"+strconv.Itoa(idx)), []byte("reindeer"))
					wg.Done()
				}(i)
			} else {
				go func(idx int) {
					tr.Delete([]byte("dog" + strconv.Itoa(idx)))
					wg.Done()
				}(i)
			}
		}
		wg.Wait()

		bm := trie.GetBatchManager(tr)
		for i := 0; i < numOperations; i++ {
			key := []byte("dog" + strconv.Itoa(i))
			val, present := bm.Get(key)
			assert.True(t, present)
			if i%2 == 0 {
				assert.Nil(t, val)
				continue
			}
			assert.Equal(t, []byte("reindeer"), val)
		}
	})
	t.Run("add data to batch while commiting another batch to trie", func(t *testing.T) {
		tr := emptyTrie()
		waitForSignal := atomic.Bool{}
		waitForSignal.Store(true)
		startedProcessing := atomic.Bool{}
		grm := &mock.GoroutinesManagerStub{
			CanStartGoRoutineCalled: func() bool {
				startedProcessing.Store(true)
				return true
			},
			EndGoRoutineProcessingCalled: func() {
				for waitForSignal.Load() {
					time.Sleep(time.Millisecond * 100)
				}
			},
			SetErrorCalled: func(err error) {
				assert.Fail(t, "should not have called this function")
			},
		}
		trie.SetGoRoutinesManager(tr, grm)

		firstBatchOperations := 1000
		secondBatchOperations := 500
		for i := 0; i < firstBatchOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			trie.ExecuteUpdatesFromBatch(tr)
			wg.Done()
		}()

		// wait for start of processing of the first batch
		for !startedProcessing.Load() {
			time.Sleep(time.Millisecond * 100)
		}

		for i := firstBatchOperations; i < firstBatchOperations+secondBatchOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}

		waitForSignal.Store(false)
		// wait for end of processing of the first batch
		wg.Wait()

		bm := trie.GetBatchManager(tr)
		for i := 0; i < firstBatchOperations; i++ {
			val, _, err := tr.Get([]byte("dog" + strconv.Itoa(i)))
			assert.Nil(t, err)
			assert.Equal(t, []byte("reindeer"), val)

			val, found := bm.Get([]byte("dog" + strconv.Itoa(i)))
			assert.False(t, found)
			assert.Nil(t, val)
		}

		for i := firstBatchOperations; i < firstBatchOperations+secondBatchOperations; i++ {
			val, _, err := tr.Get([]byte("dog" + strconv.Itoa(i)))
			assert.Nil(t, err)
			assert.Equal(t, []byte("reindeer"), val)

			val, found := bm.Get([]byte("dog" + strconv.Itoa(i)))
			assert.True(t, found)
			assert.Equal(t, []byte("reindeer"), val)
		}
	})
	t.Run("commit batch to trie while commiting another batch to trie", func(t *testing.T) {
		tr := emptyTrie()
		waitForSignal := atomic.Bool{}
		waitForSignal.Store(true)
		startedProcessing := atomic.Bool{}
		grm := &mock.GoroutinesManagerStub{
			CanStartGoRoutineCalled: func() bool {
				startedProcessing.Store(true)
				return true
			},
			EndGoRoutineProcessingCalled: func() {
				for waitForSignal.Load() {
					time.Sleep(time.Millisecond * 100)
				}
			},
			SetErrorCalled: func(err error) {
				assert.Fail(t, "should not have called this function")
			},
		}
		trie.SetGoRoutinesManager(tr, grm)

		firstBatchOperations := 1000
		secondBatchOperations := 500
		for i := 0; i < firstBatchOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			trie.ExecuteUpdatesFromBatch(tr)
			wg.Done()
		}()

		// wait for start of processing of the first batch
		for !startedProcessing.Load() {
			time.Sleep(time.Millisecond * 100)
		}

		for i := firstBatchOperations; i < firstBatchOperations+secondBatchOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}

		wg.Add(1)
		go func() {
			trie.ExecuteUpdatesFromBatch(tr)
			wg.Done()
		}()

		waitForSignal.Store(false)
		// wait for end of processing of both batches
		wg.Wait()

		bm := trie.GetBatchManager(tr)
		for i := 0; i < firstBatchOperations+secondBatchOperations; i++ {
			val, _, err := tr.Get([]byte("dog" + strconv.Itoa(i)))
			assert.Nil(t, err)
			assert.Equal(t, []byte("reindeer"), val)

			val, found := bm.Get(keyBuilder.KeyBytesToHex([]byte("dog" + strconv.Itoa(i))))
			assert.False(t, found)
			assert.Nil(t, val)
		}
	})
}

func TestPatriciaMerkleTrie_Get(t *testing.T) {
	t.Parallel()

	t.Run("get multiple values from trie in parallel", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		numOperations := 1000
		for i := 0; i < numOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"+strconv.Itoa(i)))
		}
		trie.ExecuteUpdatesFromBatch(tr)

		wg := sync.WaitGroup{}
		wg.Add(numOperations)
		for i := 0; i < numOperations; i++ {
			go func(idx int) {
				val, _, err := tr.Get([]byte("dog" + strconv.Itoa(idx)))
				assert.Nil(t, err)
				assert.Equal(t, []byte("reindeer"+strconv.Itoa(idx)), val)
				wg.Done()
			}(i)
		}
		wg.Wait()
	})
	t.Run("get multiple values from trie and batch in parallel", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		numTrieValues := 1000
		numBatchValues := 500
		for i := 0; i < numTrieValues; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"+strconv.Itoa(i)))
		}
		trie.ExecuteUpdatesFromBatch(tr)

		for i := numTrieValues; i < numTrieValues+numBatchValues; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"+strconv.Itoa(i)))
		}

		// check some values are in the batch
		bm := trie.GetBatchManager(tr)
		for i := numTrieValues; i < numTrieValues+numBatchValues; i++ {
			val, found := bm.Get([]byte("dog" + strconv.Itoa(i)))
			assert.True(t, found)
			assert.Equal(t, []byte("reindeer"+strconv.Itoa(i)), val)
		}

		wg := sync.WaitGroup{}
		wg.Add(numTrieValues + numBatchValues)
		for i := 0; i < numTrieValues+numBatchValues; i++ {
			go func(idx int) {
				val, _, err := tr.Get([]byte("dog" + strconv.Itoa(idx)))
				assert.Nil(t, err)
				assert.Equal(t, []byte("reindeer"+strconv.Itoa(idx)), val)
				wg.Done()
			}(i)
		}
		wg.Wait()
	})
	t.Run("get multiple values from collapsed trie", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		numTrieValues := 1000
		numBatchValues := 500
		for i := 0; i < numTrieValues; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"+strconv.Itoa(i)))
		}
		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

		// collapse the trie
		rootHash, _ := tr.RootHash()
		tr, _ = tr.Recreate(holders.NewDefaultRootHashesHolder(rootHash), "")

		for i := numTrieValues; i < numTrieValues+numBatchValues; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"+strconv.Itoa(i)))
		}

		// check some values are in the batch
		bm := trie.GetBatchManager(tr)
		for i := numTrieValues; i < numTrieValues+numBatchValues; i++ {
			val, found := bm.Get([]byte("dog" + strconv.Itoa(i)))
			assert.True(t, found)
			assert.Equal(t, []byte("reindeer"+strconv.Itoa(i)), val)
		}

		// get from collapsed trie
		wg := sync.WaitGroup{}
		wg.Add(numTrieValues + numBatchValues)
		for i := 0; i < numTrieValues+numBatchValues; i++ {
			go func(idx int) {
				val, _, err := tr.Get([]byte("dog" + strconv.Itoa(idx)))
				assert.Nil(t, err)
				assert.Equal(t, []byte("reindeer"+strconv.Itoa(idx)), val)
				wg.Done()
			}(i)
		}
		wg.Wait()
	})
	t.Run("get value from batch that is currently being committed", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		waitForSignal := atomic.Bool{}
		waitForSignal.Store(true)
		startedProcessing := atomic.Bool{}
		grm := &mock.GoroutinesManagerStub{
			CanStartGoRoutineCalled: func() bool {
				startedProcessing.Store(true)
				return true
			},
			EndGoRoutineProcessingCalled: func() {
				for waitForSignal.Load() {
					time.Sleep(time.Millisecond * 100)
				}
			},
			SetErrorCalled: func(err error) {
				assert.Fail(t, "should not have called this function")
			},
		}
		trie.SetGoRoutinesManager(tr, grm)

		numOperations := 1000
		for i := 0; i < numOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			trie.ExecuteUpdatesFromBatch(tr)
			wg.Done()
		}()

		// wait for start of processing of the first batch
		for !startedProcessing.Load() {
			time.Sleep(time.Millisecond * 100)
		}

		for i := 0; i < numOperations; i++ {
			val, _, err := tr.Get([]byte("dog" + strconv.Itoa(i)))
			assert.Nil(t, err)
			assert.Equal(t, []byte("reindeer"), val)
		}

		waitForSignal.Store(false)
		// wait for end of processing of both batches
		wg.Wait()
	})
}

func TestPatriciaMerkleTrie_RootHash(t *testing.T) {
	t.Parallel()

	t.Run("set root hash with batched data commits batch", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		numOperations := 1000
		for i := 0; i < numOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}

		rootHash, err := tr.RootHash()
		assert.Nil(t, err)
		assert.NotEqual(t, emptyTrieHash, rootHash)
	})
	t.Run("set root hash and update trie concurrently should serialize operations", func(t *testing.T) {
		t.Parallel()

		// create trie with some data
		tr := emptyTrie()
		numOperations := 1000
		for i := 0; i < numOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}
		trie.ExecuteUpdatesFromBatch(tr)

		waitForSignal := atomic.Bool{}
		waitForSignal.Store(true)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for waitForSignal.Load() {
				rootHash1, err := tr.RootHash()
				assert.Nil(t, err)
				assert.NotEqual(t, emptyTrieHash, rootHash1)
			}
			wg.Done()
		}()

		for i := numOperations; i < numOperations*10; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}

		waitForSignal.Store(false)
		wg.Wait()

		// check root hash is set
		rootHash1, err := tr.RootHash()
		assert.Nil(t, err)
		assert.NotEqual(t, emptyTrieHash, rootHash1)

		// check no vals are in the batch
		batchManager := trie.GetBatchManager(tr)
		batch, err := batchManager.MarkTrieUpdateInProgress()
		assert.Nil(t, err)
		insertData := batch.GetSortedDataForInsertion()
		assert.Equal(t, 0, len(insertData))
		batchManager.MarkTrieUpdateCompleted()
	})
	t.Run("set root hash and get from trie concurrently", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		numOperations := 100000
		for i := 0; i < numOperations; i++ {
			tr.Update([]byte("dog"+strconv.Itoa(i)), []byte("reindeer"))
		}
		trie.ExecuteUpdatesFromBatch(tr)

		wg := sync.WaitGroup{}
		wg.Add(1)
		setRootHashFinished := atomic.Bool{}
		go func() {
			for !setRootHashFinished.Load() {
				index := rand.Intn(numOperations)
				val, _, err := tr.Get([]byte("dog" + strconv.Itoa(index)))
				assert.Nil(t, err)
				assert.Equal(t, []byte("reindeer"), val)
			}
			wg.Done()
		}()

		rootHash, err := tr.RootHash()
		assert.Nil(t, err)
		assert.NotEqual(t, emptyTrieHash, rootHash)

		setRootHashFinished.Store(true)
		wg.Wait()
	})
}

func TestPatricianMerkleTrie_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	t.Run("rootNode changes while retrieving from the trie, check consistency", func(t *testing.T) {
		t.Parallel()

		tr := emptyTrie()
		numValsPerBatch := 100
		numBatches := 1000

		keys := make([][]byte, numValsPerBatch*numBatches)

		for i := 0; i < numValsPerBatch; i++ {
			key := []byte("dog" + strconv.Itoa(i))
			tr.Update(key, []byte("reindeer"+strconv.Itoa(i)))
			keys[i] = key
		}
		trie.ExecuteUpdatesFromBatch(tr)

		latestIndex := atomic.Int32{}
		latestIndex.Store(int32(numValsPerBatch))
		finishedProcessing := atomic.Bool{}
		go func() {
			for !finishedProcessing.Load() {
				getIndex := rand.Intn(int(latestIndex.Load()))
				val, _, err := tr.Get(keys[getIndex])
				assert.Nil(t, err)
				assert.Equal(t, []byte("reindeer"+strconv.Itoa(getIndex)), val)
			}
		}()

		for i := 1; i < numBatches; i++ {
			for j := 0; j < numValsPerBatch; j++ {
				keyIndex := i*numValsPerBatch + j
				key := []byte("dog" + strconv.Itoa(keyIndex))
				tr.Update(key, []byte("reindeer"+strconv.Itoa(keyIndex)))
				keys[keyIndex] = key
			}
			trie.ExecuteUpdatesFromBatch(tr)

			latestIndex.Store(int32((i + 1) * numValsPerBatch))
		}
		finishedProcessing.Store(true)
		rootHash1, err := tr.RootHash()
		assert.Nil(t, err)

		tr2 := emptyTrie()
		for i, key := range keys {
			tr2.Update(key, []byte("reindeer"+strconv.Itoa(i)))
		}
		trie.ExecuteUpdatesFromBatch(tr2)
		rootHash2, err := tr2.RootHash()
		assert.Nil(t, err)
		assert.Equal(t, rootHash1, rootHash2)
	})
}

func BenchmarkPatriciaMerkleTree_Insert(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.NewKeccak()

	nrValuesInTrie := 1000000
	nrValuesNotInTrie := 9000000
	values := make([][]byte, nrValuesNotInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		val := hsh.Compute(strconv.Itoa(i))
		tr.Update(val, val)
	}
	for i := 0; i < nrValuesNotInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i + nrValuesInTrie))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Update(values[i%nrValuesNotInTrie], values[i%nrValuesNotInTrie])
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
		tr.Update(val, val)
	}
	for i := 0; i < nrValuesNotInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i + nrValuesInTrie))
	}
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Update(values[i%nrValuesNotInTrie], values[i%nrValuesNotInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Delete(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.NewKeccak()

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Delete(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_DeleteCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.NewKeccak()

	nrValuesInTrie := 1000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}

	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Delete(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Get(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.NewKeccak()

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
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
		tr.Update(values[i], values[i])
	}
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

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
			tr.Update(hash, hash)
		}
		b.StartTimer()

		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
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

		tr.Update(key, value)
		values[i] = key
	}
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < nrOfValuesToModify; j++ {
			tr.Update(values[j], values[j])
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

		tr.Update(key, value)
		values[i] = key
	}
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < nrOfValuesToModify; j++ {
			b.StopTimer()
			tr.Update(values[j], values[j])
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

		tr.Update(key, value)
		values[i] = key
	}
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < nrValuesInTrie; j++ {
			tr.Update(values[j], values[j])
		}
	}
}
