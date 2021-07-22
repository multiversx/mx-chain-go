package evictionWaitingList

import (
	"encoding/binary"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/stretchr/testify/require"
)

var testHasher = blake2b.NewBlake2b()
var testHashes map[string]temporary.ModifiedHashes
var roothashes []string
var hashes [][]byte

func initTestHashes() {
	if testHashes == nil {
		testHashes, roothashes, hashes = generateTestHashes(10000, 100)
	}
}

func initEWL() *evictionWaitingList {
	initTestHashes()
	ewl, _ := NewEvictionWaitingList(100000, mock.NewMemDbMock(), &marshal.GogoProtoMarshalizer{})

	for _, roothash := range roothashes {
		_ = ewl.Put([]byte(roothash), testHashes[roothash])
	}

	return ewl
}

func initMemoryEWL() *memoryEvictionWaitingList {
	initTestHashes()
	ewl, _ := NewMemoryEvictionWaitingListV2(100000, &marshal.GogoProtoMarshalizer{})

	for _, roothash := range roothashes {
		_ = ewl.Put([]byte(roothash), testHashes[roothash])
	}

	return ewl
}

func generateTestHashes(numRoothashes int, numHashesOnRoothash int) (map[string]temporary.ModifiedHashes, []string, [][]byte) {
	counter := 0
	results := make(map[string]temporary.ModifiedHashes, numRoothashes)
	resultsRoothashes := make([]string, 0, numRoothashes)
	resultsHashes := make([][]byte, 0, numRoothashes*numHashesOnRoothash)
	for i := 0; i < numRoothashes; i++ {
		rootHash := string(intToHash(counter))
		counter++

		var newHashes temporary.ModifiedHashes
		newHashes, counter = generateHashes(counter, numHashesOnRoothash)
		for h := range newHashes {
			resultsHashes = append(resultsHashes, []byte(h))
		}

		results[rootHash] = newHashes

		resultsRoothashes = append(resultsRoothashes, rootHash)
	}

	return results, resultsRoothashes, resultsHashes
}

func intToHash(value int) []byte {
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, uint64(value))

	return testHasher.Compute(string(buff))
}

func generateHashes(counter int, numHashesOnRoothash int) (temporary.ModifiedHashes, int) {
	result := make(map[string]struct{}, numHashesOnRoothash)

	for i := 0; i < numHashesOnRoothash; i++ {
		hash := intToHash(counter)
		counter++

		result[string(hash)] = struct{}{}
	}

	return result, counter
}

func BenchmarkEvictionWaitingList_Put(b *testing.B) {
	localEwl, err := NewEvictionWaitingList(10000, mock.NewMemDbMock(), &marshal.GogoProtoMarshalizer{})
	require.Nil(b, err)
	initTestHashes()
	b.ResetTimer()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(roothashes)
		roothash := roothashes[idx]
		modifiedHashes := testHashes[roothash]

		b.StartTimer()
		err = localEwl.Put([]byte(roothash), modifiedHashes)
		b.StopTimer()
		require.Nil(b, err)
	}
}

func BenchmarkMemoryEvictionWaitingList_Put(b *testing.B) {
	ewl, err := NewMemoryEvictionWaitingListV2(10000, &marshal.GogoProtoMarshalizer{})
	require.Nil(b, err)
	initTestHashes()
	b.ResetTimer()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(roothashes)
		roothash := roothashes[idx]
		modifiedHashes := testHashes[roothash]

		b.StartTimer()
		err = ewl.Put([]byte(roothash), modifiedHashes)
		b.StopTimer()
		require.Nil(b, err)
	}
}

func BenchmarkEvictionWaitingList_Evict(b *testing.B) {
	ewl := initEWL()
	b.ResetTimer()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(roothashes)
		roothash := roothashes[idx]

		b.StartTimer()
		evicted, err := ewl.Evict([]byte(roothash))
		b.StopTimer()
		require.Nil(b, err)
		require.True(b, len(evicted) > 0)

		_ = ewl.Put([]byte(roothash), testHashes[roothash])
	}
}

func BenchmarkMemoryEvictionWaitingList_Evict(b *testing.B) {
	ewl := initMemoryEWL()
	b.ResetTimer()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(roothashes)
		roothash := roothashes[idx]

		b.StartTimer()
		evicted, err := ewl.Evict([]byte(roothash))
		b.StopTimer()
		require.Nil(b, err)
		require.True(b, len(evicted) > 0)

		_ = ewl.Put([]byte(roothash), testHashes[roothash])
	}
}

func BenchmarkEvictionWaitingList_ShouldKeep(b *testing.B) {
	ewl := initEWL()
	b.ResetTimer()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(hashes)
		hash := hashes[idx]

		b.StartTimer()
		_, err := ewl.ShouldKeepHash(string(hash), temporary.TriePruningIdentifier(i%2))
		b.StopTimer()
		require.Nil(b, err)
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
		_, err := ewl.ShouldKeepHash(string(hash), temporary.TriePruningIdentifier(i%2))
		b.StopTimer()
		require.Nil(b, err)
	}
}
