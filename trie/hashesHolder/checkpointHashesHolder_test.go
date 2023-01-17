package hashesHolder

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	chh := NewCheckpointHashesHolder(10, testscommon.HashSize)
	assert.False(t, check.IfNil(chh))
}

type testValues struct {
	rootHashes [][]byte
	values     []common.ModifiedHashes
}

func getTestValues() *testValues {
	hashes1 := make(map[string]struct{})
	hashes1["hash1"] = struct{}{}
	hashes1["hash2"] = struct{}{}
	hashes1["hash3"] = struct{}{}

	hashes2 := make(map[string]struct{})
	hashes2["hash4"] = struct{}{}
	hashes2["hash5"] = struct{}{}
	hashes2["hash6"] = struct{}{}

	hashes3 := make(map[string]struct{})
	hashes3["hash7"] = struct{}{}
	hashes3["hash8"] = struct{}{}
	hashes3["hash9"] = struct{}{}

	rootHash1 := []byte("rootHash1")
	rootHash2 := []byte("rootHash2")
	rootHash3 := []byte("rootHash3")

	testData := &testValues{
		rootHashes: [][]byte{rootHash1, rootHash2, rootHash3},
		values:     []common.ModifiedHashes{hashes1, hashes2, hashes3},
	}

	return testData
}

func TestCheckpointHashesHolder_Put(t *testing.T) {
	t.Parallel()

	chh := NewCheckpointHashesHolder(191, testscommon.HashSize)
	testData := getTestValues()

	shouldCreateCheckpoint := chh.Put(testData.rootHashes[0], testData.values[0])
	assert.False(t, shouldCreateCheckpoint)
	shouldCreateCheckpoint = chh.Put(testData.rootHashes[1], testData.values[1])
	assert.True(t, shouldCreateCheckpoint)
	_ = chh.Put(testData.rootHashes[2], testData.values[2])

	assert.Equal(t, 3, len(chh.hashes))
	assert.Equal(t, 3, len(chh.hashes[0]))
	assert.Equal(t, 3, len(chh.hashes[1]))
	assert.Equal(t, 3, len(chh.hashes[2]))

	assert.Equal(t, testData.rootHashes[0], chh.rootHashes[0])
	assert.Equal(t, testData.values[0], chh.hashes[0])
	assert.Equal(t, testData.rootHashes[1], chh.rootHashes[1])
	assert.Equal(t, testData.values[1], chh.hashes[1])
	assert.Equal(t, testData.rootHashes[2], chh.rootHashes[2])
	assert.Equal(t, testData.values[2], chh.hashes[2])

	assert.Equal(t, uint64(315), chh.currentSize)
}

func TestCheckpointHashesHolder_PutSameRootHashDoesNotAppend(t *testing.T) {
	t.Parallel()

	chh := NewCheckpointHashesHolder(500, testscommon.HashSize)
	testData := getTestValues()

	_ = chh.Put(testData.rootHashes[0], testData.values[0])
	_ = chh.Put(testData.rootHashes[0], testData.values[1])
	_ = chh.Put(testData.rootHashes[0], testData.values[2])

	assert.Equal(t, 1, len(chh.hashes))
	assert.Equal(t, 1, len(chh.rootHashes))

	assert.Equal(t, testData.rootHashes[0], chh.rootHashes[0])
	assert.Equal(t, testData.values[0], chh.hashes[0])
}

func TestCheckpointHashesHolder_ShouldCommit(t *testing.T) {
	t.Parallel()

	chh := NewCheckpointHashesHolder(500, testscommon.HashSize)
	testData := getTestValues()

	_ = chh.Put(testData.rootHashes[0], testData.values[0])
	_ = chh.Put(testData.rootHashes[1], testData.values[1])
	_ = chh.Put(testData.rootHashes[2], testData.values[2])

	assert.True(t, chh.ShouldCommit([]byte("hash3")))
	assert.True(t, chh.ShouldCommit([]byte("hash4")))
	assert.True(t, chh.ShouldCommit([]byte("hash8")))
	assert.False(t, chh.ShouldCommit([]byte("hash10")))
}

func TestCheckpointHashesHolder_RemoveCommitted(t *testing.T) {
	t.Parallel()

	chh := NewCheckpointHashesHolder(500, testscommon.HashSize)
	testData := getTestValues()

	_ = chh.Put(testData.rootHashes[0], testData.values[0])
	_ = chh.Put(testData.rootHashes[1], testData.values[1])
	_ = chh.Put(testData.rootHashes[2], testData.values[2])
	assert.Equal(t, uint64(315), chh.currentSize)

	chh.RemoveCommitted(testData.rootHashes[1])
	assert.Equal(t, 1, len(chh.hashes))
	assert.Equal(t, 3, len(chh.hashes[0]))
	assert.Equal(t, uint64(105), chh.currentSize)

	assert.NotEqual(t, chh.rootHashes[0], testData.rootHashes[0])
	assert.NotEqual(t, chh.hashes[0], testData.values[0])
	assert.NotEqual(t, chh.rootHashes[0], testData.rootHashes[1])
	assert.NotEqual(t, chh.hashes[0], testData.values[1])
	assert.Equal(t, chh.rootHashes[0], testData.rootHashes[2])
	assert.Equal(t, chh.hashes[0], testData.values[2])
}

func TestCheckpointHashesHolder_RemoveCommittedInvalidSizeComputation(t *testing.T) {
	t.Parallel()

	chh := NewCheckpointHashesHolder(500, testscommon.HashSize)
	testData := getTestValues()

	_ = chh.Put(testData.rootHashes[0], testData.values[0])
	_ = chh.Put(testData.rootHashes[1], testData.values[1])
	_ = chh.Put(testData.rootHashes[2], testData.values[2])
	assert.Equal(t, uint64(315), chh.currentSize)
	chh.currentSize = 0

	chh.RemoveCommitted(testData.rootHashes[1])
	assert.Equal(t, 1, len(chh.hashes))
	assert.Equal(t, 3, len(chh.hashes[0]))
	assert.Equal(t, uint64(105), chh.currentSize)
}

func TestCheckpointHashesHolder_Remove(t *testing.T) {
	t.Parallel()

	chh := NewCheckpointHashesHolder(500, testscommon.HashSize)
	testData := getTestValues()

	_ = chh.Put(testData.rootHashes[0], testData.values[0])
	_ = chh.Put(testData.rootHashes[1], testData.values[1])
	_ = chh.Put(testData.rootHashes[2], testData.values[2])
	assert.Equal(t, uint64(315), chh.currentSize)

	chh.Remove([]byte("hash5"))
	assert.Equal(t, 3, len(chh.hashes))
	assert.Equal(t, 2, len(chh.hashes[1]))
	assert.Equal(t, uint64(283), chh.currentSize)
}

func TestCheckpointHashesHolder_RemoveInvalidSizeComputation(t *testing.T) {
	t.Parallel()

	chh := NewCheckpointHashesHolder(500, testscommon.HashSize)
	testData := getTestValues()

	_ = chh.Put(testData.rootHashes[0], testData.values[0])
	_ = chh.Put(testData.rootHashes[1], testData.values[1])
	_ = chh.Put(testData.rootHashes[2], testData.values[2])
	assert.Equal(t, uint64(315), chh.currentSize)
	chh.currentSize = 1

	chh.Remove([]byte("hash5"))
	assert.Equal(t, 3, len(chh.hashes))
	assert.Equal(t, 2, len(chh.hashes[1]))
	assert.Equal(t, uint64(283), chh.currentSize)
}
