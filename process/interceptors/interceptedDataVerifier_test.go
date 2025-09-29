package interceptors

import (
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-go/common"
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

	err := verifier.Verify(interceptedData, "topic")
	require.NoError(t, err)
	require.Equal(t, 1, checkValidityCounter)

	err = verifier.Verify(interceptedData, "topic")
	require.NoError(t, err)
	require.Equal(t, 2, checkValidityCounter)
}

func TestInterceptedDataVerifier_checkCachedData(t *testing.T) {
	t.Parallel()

	t.Run("already cached invalid data should error", func(t *testing.T) {
		t.Parallel()

		interceptedData := &testscommon.InterceptedDataStub{
			CheckValidityCalled: func() error {
				require.Fail(t, "should have not been called")
				return nil
			},
			HashCalled: func() []byte {
				return []byte("hash")
			},
		}

		verifier := defaultInterceptedDataVerifier(defaultSpan)

		verifier.PutInCache(interceptedData, interceptedDataStatus(1)) // not validInterceptedData

		err := verifier.Verify(interceptedData, "topic")
		require.Equal(t, process.ErrInvalidInterceptedData, err)
	})
	t.Run("already cached intra shard data should work", func(t *testing.T) {
		t.Parallel()

		interceptedData := &testscommon.InterceptedDataStub{
			CheckValidityCalled: func() error {
				require.Fail(t, "should have not been called")
				return nil
			},
			HashCalled: func() []byte {
				return []byte("hash")
			},
			ShouldAllowDuplicatesCalled: func() bool {
				require.Fail(t, "should have not been called")
				return true
			},
		}

		verifier := defaultInterceptedDataVerifier(defaultSpan)

		verifier.PutInCache(interceptedData, validInterceptedData)

		err := verifier.Verify(interceptedData, "topic_1_REQUEST") // intra shard
		require.NoError(t, err)
	})
	t.Run("already cached data that does not allow duplicates should error", func(t *testing.T) {
		t.Parallel()

		interceptedData := &testscommon.InterceptedDataStub{
			CheckValidityCalled: func() error {
				require.Fail(t, "should have not been called")
				return nil
			},
			HashCalled: func() []byte {
				return []byte("hash")
			},
			ShouldAllowDuplicatesCalled: func() bool {
				return false
			},
		}

		verifier := defaultInterceptedDataVerifier(defaultSpan)

		verifier.PutInCache(interceptedData, validInterceptedData)

		err := verifier.Verify(interceptedData, "topic_0_1")
		require.Equal(t, process.ErrDuplicatedInterceptedDataNotAllowed, err)
	})
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
	interceptedDataWithNoHash := &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("") // peerAuthentication
		},
	}

	verifier := defaultInterceptedDataVerifier(defaultSpan)

	err := verifier.Verify(interceptedData, "topic")
	require.NoError(t, err)

	verifier.MarkVerified(interceptedData, "topic_1") // intra shard, for coverage only

	verifier.MarkVerified(interceptedDataWithNoHash, "topic") // no hash, for coverage only

	verifier.MarkVerified(interceptedData, "topic")

	errCount := atomic.Counter{}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := verifier.Verify(interceptedData, "topic")
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

	err := verifier.Verify(interceptedData, "topic")
	require.Equal(t, process.ErrInvalidInterceptedData, err)

	errCount := atomic.Counter{}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := verifier.Verify(interceptedData, "topic")
			if err != nil {
				errCount.Add(1)
			}
		}()
	}
	wg.Wait()

	require.Equal(t, int64(100), errCount.Get())
	require.Equal(t, int64(101), checkValidityCounter.Get())
}

func TestInterceptedDataVerifier_CheckValidityInterceptedProof(t *testing.T) {
	t.Parallel()

	interceptedData := &testscommon.InterceptedDataStub{
		CheckValidityCalled: func() error {
			return common.ErrAlreadyExistingEquivalentProof // for coverage only
		},
		HashCalled: func() []byte {
			return []byte("hash")
		},
	}

	verifier := defaultInterceptedDataVerifier(defaultSpan)

	err := verifier.Verify(interceptedData, "topic")
	require.Equal(t, process.ErrInvalidInterceptedData, err)
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
		err := verifier.Verify(interceptedData, "topic")
		require.NoError(t, err)
		require.Equal(t, int64(1), checkValidityCounter.Get())

		verifier.MarkVerified(interceptedData, "topic")

		// Second retrieval should be from the cache.
		err = verifier.Verify(interceptedData, "topic")
		require.NoError(t, err)
		require.Equal(t, int64(1), checkValidityCounter.Get())

		// Wait for the cache expiry
		<-time.After(expiryTestDuration + 100*time.Millisecond)

		// Third retrieval should reach validity check again.
		err = verifier.Verify(interceptedData, "topic")
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
		err := verifier.Verify(interceptedData, "topic")
		require.Equal(t, process.ErrInvalidInterceptedData, err)
		require.Equal(t, int64(1), checkValidityCounter.Get())

		// Second retrieval
		err = verifier.Verify(interceptedData, "topic")
		require.Equal(t, process.ErrInvalidInterceptedData, err)
		require.Equal(t, int64(2), checkValidityCounter.Get())

		// Wait for the cache expiry
		<-time.After(expiryTestDuration + 100*time.Millisecond)

		// Third retrieval should reach validity check again.
		err = verifier.Verify(interceptedData, "topic")
		require.Equal(t, process.ErrInvalidInterceptedData, err)
		require.Equal(t, int64(3), checkValidityCounter.Get())
	})
}
