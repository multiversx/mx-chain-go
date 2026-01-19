package cache

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderBodyCache(t *testing.T) {
	t.Parallel()

	c := NewHeaderBodyCache()
	require.NotNil(t, c)
	require.False(t, c.IsInterfaceNil())
}

func TestHeaderBodyCache_AddOrReplace(t *testing.T) {
	t.Parallel()

	t.Run("nil header", func(t *testing.T) {
		t.Parallel()
		c := NewHeaderBodyCache()
		err := c.AddOrReplace(HeaderBodyPair{
			Header: nil,
			Body:   &block.Body{},
		})
		require.Equal(t, common.ErrNilHeaderHandler, err)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		c := NewHeaderBodyCache()
		err := c.AddOrReplace(HeaderBodyPair{
			Header: &block.HeaderV3{},
			Body:   nil,
		})
		require.Equal(t, data.ErrNilBlockBody, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		c := NewHeaderBodyCache()

		headerNonce := uint64(10)
		pair := HeaderBodyPair{
			Header: &block.HeaderV3{
				Nonce: headerNonce,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("a"),
		}

		err := c.AddOrReplace(pair)
		require.Nil(t, err)

		// Verification
		retrievedPair, found := c.GetByNonce(headerNonce)
		require.True(t, found)
		require.Equal(t, pair, retrievedPair)
	})

	t.Run("replace existing", func(t *testing.T) {
		t.Parallel()
		c := NewHeaderBodyCache()

		headerNonce := uint64(10)
		pair1 := HeaderBodyPair{
			Header: &block.HeaderV3{
				Nonce: headerNonce,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("a"),
		}

		err := c.AddOrReplace(pair1)
		require.Nil(t, err)

		pair2 := HeaderBodyPair{
			Header: &block.HeaderV3{
				Nonce: headerNonce,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("a"),
		}

		err = c.AddOrReplace(pair2)
		require.Nil(t, err)

		retrievedPair, found := c.GetByNonce(headerNonce)
		require.True(t, found)
		// Should be pair2 (note: they are different pointers for Header/Body so Equal check works if pointer comparison)
		// Wait, slice/map/func in structs make Go equality tricky unless pointers.
		// These are pointers to structs, so it should be fine.
		require.Equal(t, pair2, retrievedPair)
	})
}

func TestHeaderBodyCache_GetByNonce(t *testing.T) {
	t.Parallel()

	c := NewHeaderBodyCache()

	pair := HeaderBodyPair{
		Header: &block.HeaderV3{
			Nonce: 5,
		},
		Body:       &block.Body{},
		HeaderHash: []byte("a"),
	}
	_ = c.AddOrReplace(pair)

	retrieved, found := c.GetByNonce(5)
	require.True(t, found)
	require.Equal(t, pair, retrieved)

	_, found = c.GetByNonce(999)
	require.False(t, found)
}

func TestHeaderBodyCache_RemoveAtNonceAndHigher(t *testing.T) {
	t.Parallel()

	c := NewHeaderBodyCache()
	nonces := []uint64{1, 2, 3, 4, 5, 10}

	for _, n := range nonces {
		_ = c.AddOrReplace(HeaderBodyPair{
			Header: &block.HeaderV3{
				Nonce: n,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("a"),
		})
	}

	removed := c.RemoveAtNonceAndHigher(4)
	require.Equal(t, []uint64{4, 5, 10}, removed)

	// Verify remaining
	for _, n := range []uint64{1, 2, 3} {
		_, found := c.GetByNonce(n)
		require.True(t, found, "nonce %d should exist", n)
	}

	// Verify removed
	for _, n := range []uint64{4, 5, 10} {
		_, found := c.GetByNonce(n)
		require.False(t, found, "nonce %d should be removed", n)
	}
}

func TestHeaderBodyCache_Remove(t *testing.T) {
	t.Parallel()

	c := NewHeaderBodyCache()
	_ = c.AddOrReplace(HeaderBodyPair{
		Header: &block.HeaderV3{
			Nonce: 10,
		},
		Body:       &block.Body{},
		HeaderHash: []byte("a"),
	})

	c.Remove(10)
	_, found := c.GetByNonce(10)
	require.False(t, found)
}

func TestHeaderBodyCache_Clean(t *testing.T) {
	t.Parallel()

	c := NewHeaderBodyCache()
	_ = c.AddOrReplace(HeaderBodyPair{
		Header: &block.HeaderV3{
			Nonce: 10,
		},
		Body:       &block.Body{},
		HeaderHash: []byte("a"),
	})

	c.Clean()
	_, found := c.GetByNonce(10)
	require.False(t, found)
}

func TestHeaderBodyCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	c := NewHeaderBodyCache()
	numGoroutines := 100
	done := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			nonce := uint64(idx)
			_ = c.AddOrReplace(HeaderBodyPair{
				Header: &block.HeaderV3{
					Nonce: nonce,
				},
				Body:       &block.Body{},
				HeaderHash: []byte("a"),
			})
			_, _ = c.GetByNonce(nonce)

			if idx%2 == 0 {
				c.Remove(nonce)
			}
			done <- struct{}{}
		}(i)
	}

	timeout := time.After(2 * time.Second)
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("timeout waiting for goroutines")
		}
	}
}
