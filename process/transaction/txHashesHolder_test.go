package transaction

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTxHashesHolder(t *testing.T) {
	t.Parallel()

	holder := NewTxHashesHolder()
	require.NotNil(t, holder)
}

func TestTxHashesHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var holder *txHashesHolder
	require.True(t, holder.IsInterfaceNil())

	holder = NewTxHashesHolder()
	require.False(t, holder.IsInterfaceNil())
}

func TestTxHashesHolder(t *testing.T) {
	t.Parallel()

	t.Run("synced operations", func(t *testing.T) {
		t.Parallel()

		holder := NewTxHashesHolder()

		providedHash := []byte("hash")
		holder.Append(providedHash)
		hashes := holder.GetAllHashes()
		require.Equal(t, 1, len(hashes))
		require.True(t, bytes.Equal(providedHash, hashes[0]))

		holder.Reset()
		hashes = holder.GetAllHashes()
		require.Zero(t, len(hashes))
	})
	t.Run("asynced operations", func(t *testing.T) {
		t.Parallel()

		holder := NewTxHashesHolder()

		numCalls := 1000
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				switch idx % 5 {
				case 0, 1, 2:
					holder.Append([]byte(fmt.Sprintf("hash_%d", idx)))
				case 3:
					holder.GetAllHashes()
				case 4:
					holder.Reset()
				}

				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}
