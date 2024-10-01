package interceptors

import (
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
)

const defaultSpan = 1 * time.Second

func defaultInterceptedDataVerifier(span time.Duration) process.InterceptedDataVerifier {
	c, _ := cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: span,
		CacheExpiry: span,
	})

	return NewInterceptedDataVerifier(c)
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
	for i := 0; i < 3; i++ {
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
			return ErrInvalidInterceptedData
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
	require.Equal(t, ErrInvalidInterceptedData, err)

	errCount := atomic.Counter{}
	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
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

	require.Equal(t, int64(3), errCount.Get())
	require.Equal(t, int64(1), checkValidityCounter.Get())
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
				return ErrInvalidInterceptedData
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
		require.Equal(t, ErrInvalidInterceptedData, err)
		require.Equal(t, int64(1), checkValidityCounter.Get())

		// Second retrieval should be from the cache.
		err = verifier.Verify(interceptedData)
		require.Equal(t, ErrInvalidInterceptedData, err)
		require.Equal(t, int64(1), checkValidityCounter.Get())

		// Wait for the cache expiry
		<-time.After(expiryTestDuration + 100*time.Millisecond)

		// Third retrieval should reach validity check again.
		err = verifier.Verify(interceptedData)
		require.Equal(t, ErrInvalidInterceptedData, err)
		require.Equal(t, int64(2), checkValidityCounter.Get())
	})
}
