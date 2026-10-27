package headerForBlock_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
	"github.com/stretchr/testify/require"
)

func TestNewLastNotarizedHeaderInfo(t *testing.T) {
	t.Parallel()

	header := &block.Header{Nonce: 1}
	hash := []byte("hash")
	lnhi := headerForBlock.NewLastNotarizedHeaderInfo(header, hash, true, true)
	require.NotNil(t, lnhi)
	require.Equal(t, header, lnhi.GetHeader())
	require.Equal(t, hash, lnhi.GetHash())
	require.True(t, lnhi.NotarizedBasedOnProof())
	require.True(t, lnhi.HasProof())
}

func TestLastNotarizedHeaderInfo_GettersAndSetters(t *testing.T) {
	t.Parallel()

	header1 := &block.Header{Nonce: 1}
	header2 := &block.Header{Nonce: 2}
	hash1 := []byte("hash1")
	hash2 := []byte("hash2")

	lnhi := headerForBlock.NewLastNotarizedHeaderInfo(header1, hash1, false, false)

	// Test Get/Set Header
	lnhi.SetHeader(header2)
	require.Equal(t, header2, lnhi.GetHeader())

	// Test Get/Set Hash
	lnhi.SetHash(hash2)
	require.Equal(t, hash2, lnhi.GetHash())

	// Test NotarizedBasedOnProof
	lnhi.SetNotarizedBasedOnProof(true)
	require.True(t, lnhi.NotarizedBasedOnProof())

	lnhi.SetNotarizedBasedOnProof(false)
	require.False(t, lnhi.NotarizedBasedOnProof())

	// Test HasProof
	lnhi.SetHasProof(true)
	require.True(t, lnhi.HasProof())

	lnhi.SetHasProof(false)
	require.False(t, lnhi.HasProof())
}

func TestLastNotarizedHeaderInfo_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	header := &block.Header{Nonce: 1}
	lnhi := headerForBlock.NewLastNotarizedHeaderInfo(header, []byte("hash"), true, true)
	require.False(t, lnhi.IsInterfaceNil())
}

func TestLastNotarizedHeaderInfo_Concurrency(t *testing.T) {
	require.NotPanics(t, func() {
		t.Parallel()

		lnhi := headerForBlock.NewLastNotarizedHeaderInfo(&block.Header{Nonce: 1}, []byte("hash"), false, false)

		numCalls := 100
		var wg sync.WaitGroup
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(i int) {
				switch i % 8 {
				case 0:
					lnhi.SetHeader(&block.Header{Nonce: uint64(i)})
				case 1:
					lnhi.SetHash([]byte(fmt.Sprintf("hash_%d", i)))
				case 2:
					lnhi.SetNotarizedBasedOnProof(true)
				case 3:
					lnhi.SetHasProof(true)
				case 4:
					lnhi.GetHeader()
				case 5:
					lnhi.GetHash()
				case 6:
					lnhi.NotarizedBasedOnProof()
				case 7:
					lnhi.HasProof()
				default:
					require.Fail(t, "should have not been called")
				}
				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}
