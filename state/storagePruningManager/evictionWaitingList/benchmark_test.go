package evictionWaitingList

import (
	"encoding/binary"
	"testing"

	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/stretchr/testify/require"
)

var testHasher = blake2b.NewBlake2b()
var testHashes map[string]*rootHashData
var rootHashes []string
var hashes [][]byte

func initTestHashes() {
	if testHashes == nil {
		testHashes, rootHashes, hashes = generateTestHashes(10000, 100)
	}
}

func initMemoryEWL() *memoryEvictionWaitingList {
	initTestHashes()
	args := MemoryEvictionWaitingListArgs{
		RootHashesSize: 10000000,
		HashesSize:     10000000,
	}
	ewl, _ := NewMemoryEvictionWaitingList(args)

	for _, roothash := range rootHashes {
		_ = ewl.Put([]byte(roothash), testHashes[roothash].hashes)
	}

	return ewl
}

func generateTestHashes(numRoothashes int, numHashesOnRoothash int) (map[string]*rootHashData, []string, [][]byte) {
	counter := 0
	results := make(map[string]*rootHashData, numRoothashes)
	resultsRoothashes := make([]string, 0, numRoothashes)
	resultsHashes := make([][]byte, 0, numRoothashes*numHashesOnRoothash)
	for i := 0; i < numRoothashes; i++ {
		rootHash := string(intToHash(counter))
		counter++

		var newHashes common.ModifiedHashes
		newHashes, counter = generateHashes(counter, numHashesOnRoothash)
		for h := range newHashes {
			resultsHashes = append(resultsHashes, []byte(h))
		}

		results[rootHash] = &rootHashData{
			numReferences: 1,
			hashes:        newHashes,
		}

		resultsRoothashes = append(resultsRoothashes, rootHash)
	}

	return results, resultsRoothashes, resultsHashes
}

func intToHash(value int) []byte {
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, uint64(value))

	return testHasher.Compute(string(buff))
}

func generateHashes(counter int, numHashesOnRoothash int) (common.ModifiedHashes, int) {
	result := make(map[string]struct{}, numHashesOnRoothash)

	for i := 0; i < numHashesOnRoothash; i++ {
		hash := intToHash(counter)
		counter++

		result[string(hash)] = struct{}{}
	}

	return result, counter
}

func BenchmarkMemoryEvictionWaitingList_Put(b *testing.B) {
	args := MemoryEvictionWaitingListArgs{
		RootHashesSize: 10000000,
		HashesSize:     10000000,
	}
	ewl, err := NewMemoryEvictionWaitingList(args)
	require.Nil(b, err)
	initTestHashes()
	b.ResetTimer()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(rootHashes)
		roothash := rootHashes[idx]
		modifiedHashes := testHashes[roothash].hashes

		b.StartTimer()
		err = ewl.Put([]byte(roothash), modifiedHashes)
		b.StopTimer()
		require.Nil(b, err)
	}
}

func BenchmarkMemoryEvictionWaitingList_Evict(b *testing.B) {
	ewl := initMemoryEWL()
	b.ResetTimer()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(rootHashes)
		roothash := rootHashes[idx]

		b.StartTimer()
		evicted, err := ewl.Evict([]byte(roothash))
		b.StopTimer()
		require.Nil(b, err)
		require.True(b, len(evicted) > 0)

		_ = ewl.Put([]byte(roothash), testHashes[roothash].hashes)
	}
}

func BenchmarkMemoryEvictionWaitingList_ShouldKeep(b *testing.B) {
	ewl := initMemoryEWL()
	b.ResetTimer()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(hashes)
		hash := hashes[idx]

		b.StartTimer()
		_, err := ewl.ShouldKeepHash(string(hash), state.TriePruningIdentifier(i%2))
		b.StopTimer()
		require.Nil(b, err)
	}
}
