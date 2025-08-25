package headerForBlock_test

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderInfo(t *testing.T) {
	t.Parallel()

	header := &block.Header{Nonce: 1}
	hi := headerForBlock.NewHeaderInfo(header, true, true, false)
	require.NotNil(t, hi)
	require.Equal(t, header, hi.GetHeader())
	require.True(t, hi.UsedInBlock())
	require.True(t, hi.HasProof())
	require.False(t, hi.HasProofRequested())
}

func TestNewEmptyHeaderInfo(t *testing.T) {
	t.Parallel()

	hi := headerForBlock.NewEmptyHeaderInfo()
	require.NotNil(t, hi)
}

func TestHeaderInfo_GettersAndSetters(t *testing.T) {
	t.Parallel()

	header1 := &block.Header{Nonce: 1}
	hi := headerForBlock.NewEmptyHeaderInfo()

	// Test SetHeader and GetHeader
	hi.SetHeader(nil) // coverage only
	hi.SetHeader(header1)
	require.Equal(t, header1, hi.GetHeader())

	// Test SetUsedInBlock and UsedInBlock
	hi.SetUsedInBlock(true)
	require.True(t, hi.UsedInBlock())

	// Test SetHasProof and HasProof
	hi.SetHasProof(true)
	require.True(t, hi.HasProof())

	// Test SetHasProofRequested and HasProofRequested
	hi.SetHasProofRequested(true)
	require.True(t, hi.HasProofRequested())
}

func TestHeaderInfo_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	hi := headerForBlock.NewEmptyHeaderInfo()
	require.False(t, hi.IsInterfaceNil())
}

func TestHeaderInfo_Concurrency(t *testing.T) {
	require.NotPanics(t, func() {
		t.Parallel()

		hi := headerForBlock.NewEmptyHeaderInfo()

		numCalls := 100
		var wg sync.WaitGroup
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(i int) {
				switch i % 8 {
				case 0:
					hi.SetHeader(&block.Header{Nonce: uint64(i)})
				case 1:
					hi.SetUsedInBlock(true)
				case 2:
					hi.SetHasProof(true)
				case 3:
					hi.SetHasProofRequested(true)
				case 4:
					hi.GetHeader()
				case 5:
					hi.UsedInBlock()
				case 6:
					hi.HasProof()
				case 7:
					hi.HasProofRequested()
				default:
					require.Fail(t, "should have not been called")
				}
				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}
