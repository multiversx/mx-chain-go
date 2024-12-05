package common

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModifiedHashes_Clone(t *testing.T) {
	t.Parallel()

	mh := make(ModifiedHashes)
	mh["aaa"] = struct{}{}
	mh["bbb"] = struct{}{}

	cloned := mh.Clone()
	assert.NotEqual(t, fmt.Sprintf("%p", mh), fmt.Sprintf("%p", cloned)) //pointer testing
	assert.Equal(t, len(mh), len(cloned))

	for key := range mh {
		_, found := cloned[key]
		assert.True(t, found)
	}

}

func TestModifiedHashesSlice_Append(t *testing.T) {
	t.Parallel()

	numHashes := 100
	mhs := NewModifiedHashesSlice(numHashes)
	hashes := make([][]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		hashes[i] = []byte{byte(i)}
	}

	wg := sync.WaitGroup{}
	wg.Add(numHashes)
	for i := 0; i < numHashes; i++ {
		go func(index int) {
			mhs.Append([][]byte{hashes[index]})
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < numHashes; i++ {
		hashFound := false
		for _, hash := range mhs.hashes {
			if bytes.Equal(hashes[i], hash) {
				hashFound = true
			}
		}
		assert.True(t, hashFound)
	}
}

func TestModifiedHashesSlice_Get(t *testing.T) {
	t.Parallel()

	hashes := [][]byte{{1}, {2}, {3}}
	mhs := NewModifiedHashesSlice(len(hashes))
	mhs.Append(hashes)
	retrievedHashes := mhs.Get()
	assert.Equal(t, hashes, retrievedHashes)
	retrievedHashes[0] = []byte{4}
	newRetrievedHashes := mhs.Get()
	assert.NotEqual(t, retrievedHashes, newRetrievedHashes)
}
