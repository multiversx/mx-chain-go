package interceptors

import (
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
)

const defaultSpan = 1 * time.Second

func defaultInterceptedDataVerifier(span time.Duration) *interceptedDataVerifier {
	c, _ := cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: span,
		CacheExpiry: span,
	})

	verifier, _ := NewInterceptedDataVerifier(c)
	return verifier
}

func TestNewInterceptedDataVerifier(t *testing.T) {
	t.Parallel()

	var c storage.Cacher
	verifier, err := NewInterceptedDataVerifier(c)
	require.Equal(t, process.ErrNilInterceptedDataCache, err)
	require.Nil(t, verifier)
}

func TestInterceptedDataVerifier_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var verifier *interceptedDataVerifier
	require.True(t, verifier.IsInterfaceNil())

	verifier = defaultInterceptedDataVerifier(defaultSpan)
	require.False(t, verifier.IsInterfaceNil())
}

func TestInterceptedDataVerifier_EmptyHash(t *testing.T) {
	t.Parallel()

	var checkValidityCounter int
	verifier := defaultInterceptedDataVerifier(defaultSpan)
	interceptedData := &testscommon.InterceptedDataStub{
		CheckValidityCalled: func() error {
			checkValidityCounter++
			return nil
		},
		IsForCurrentShardCalled: func() bool {
			return false
		},
		HashCalled: func() []byte {
			return nil
		},
	}

	err := verifier.Verify(interceptedData)
	require.NoError(t, err)
	require.Equal(t, 1, checkValidityCounter)

	err = verifier.Verify(interceptedData)
	require.NoError(t, err)
	require.Equal(t, 2, checkValidityCounter)
}

func TestInterceptedDataVerifier_CheckValidityShouldWork(t *testing.T) {
	t.Parallel()

	checkValidityCounter := atomic.Counter{}

	interceptedData := &testscommon.InterceptedDataStub{
		CheckValidityCalled: func() error {
			checkValidityCounter.Add(1)
			return nil
		},
		IsForCurrentShardCalled: func() bool {
			return false
		},
		HashCalled: func() []byte {
			return []byte("hash")
		},
	}

	verifier := defaultInterceptedDataVerifier(defaultSpan)

	err := verifier.Verify(interceptedData)
	require.NoError(t, err)

	errCount := atomic.Counter{}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := verifier.Verify(interceptedData)
			if err != nil {
				errCount.Add(1)
			}
		}()
	}
	wg.Wait()

	require.Equal(t, int64(0), errCount.Get())
	require.Equal(t, int64(1), checkValidityCounter.Get())
}

func TestInterceptedDataVerifier_CheckValidityShouldNotWork(t *testing.T) {
	t.Parallel()

	checkValidityCounter := atomic.Counter{}
	interceptedData := &testscommon.InterceptedDataStub{
		CheckValidityCalled: func() error {
			checkValidityCounter.Add(1)
			return process.ErrInvalidInterceptedData
		},
		IsForCurrentShardCalled: func() bool {
			return false
		},
		HashCalled: func() []byte {
			return []byte("hash")
		},
	}

	verifier := defaultInterceptedDataVerifier(defaultSpan)

	err := verifier.Verify(interceptedData)
	require.Equal(t, process.ErrInvalidInterceptedData, err)

	errCount := atomic.Counter{}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := verifier.Verify(interceptedData)
			if err != nil {
				errCount.Add(1)
			}
		}()
	}
	wg.Wait()

	require.Equal(t, int64(100), errCount.Get())
	require.Equal(t, int64(101), checkValidityCounter.Get())
}

func TestInterceptedDataVerifier_CheckExpiryTime(t *testing.T) {
	t.Parallel()

	t.Run("expiry on valid data", func(t *testing.T) {
		expiryTestDuration := defaultSpan * 2

		checkValidityCounter := atomic.Counter{}

		interceptedData := &testscommon.InterceptedDataStub{
			CheckValidityCalled: func() error {
				checkValidityCounter.Add(1)
				return nil
			},
			IsForCurrentShardCalled: func() bool {
				return false
			},
			HashCalled: func() []byte {
				return []byte("hash")
			},
		}

		verifier := defaultInterceptedDataVerifier(expiryTestDuration)

		// First retrieval, check validity is reached.
		err := verifier.Verify(interceptedData)
		require.NoError(t, err)
		require.Equal(t, int64(1), checkValidityCounter.Get())

		// Second retrieval should be from the cache.
		err = verifier.Verify(interceptedData)
		require.NoError(t, err)
		require.Equal(t, int64(1), checkValidityCounter.Get())

		// Wait for the cache expiry
		<-time.After(expiryTestDuration + 100*time.Millisecond)

		// Third retrieval should reach validity check again.
		err = verifier.Verify(interceptedData)
		require.NoError(t, err)
		require.Equal(t, int64(2), checkValidityCounter.Get())
	})

	t.Run("expiry on invalid data", func(t *testing.T) {
		expiryTestDuration := defaultSpan * 2

		checkValidityCounter := atomic.Counter{}

		interceptedData := &testscommon.InterceptedDataStub{
			CheckValidityCalled: func() error {
				checkValidityCounter.Add(1)
				return process.ErrInvalidInterceptedData
			},
			IsForCurrentShardCalled: func() bool {
				return false
			},
			HashCalled: func() []byte {
				return []byte("hash")
			},
		}

		verifier := defaultInterceptedDataVerifier(expiryTestDuration)

		// First retrieval, check validity is reached.
		err := verifier.Verify(interceptedData)
		require.Equal(t, process.ErrInvalidInterceptedData, err)
		require.Equal(t, int64(1), checkValidityCounter.Get())

		// Second retrieval
		err = verifier.Verify(interceptedData)
		require.Equal(t, process.ErrInvalidInterceptedData, err)
		require.Equal(t, int64(2), checkValidityCounter.Get())

		// Wait for the cache expiry
		<-time.After(expiryTestDuration + 100*time.Millisecond)

		// Third retrieval should reach validity check again.
		err = verifier.Verify(interceptedData)
		require.Equal(t, process.ErrInvalidInterceptedData, err)
		require.Equal(t, int64(3), checkValidityCounter.Get())
	})
}
