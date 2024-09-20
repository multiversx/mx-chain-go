package interceptors

import (
	"errors"
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
	interceptedData := &testscommon.InterceptedDataStub{
		CheckValidityCalled: func() error {
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
	require.Nil(t, err)

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
}

func TestInterceptedDataVerifier_CheckValidityShouldNotWork(t *testing.T) {
	interceptedData := &testscommon.InterceptedDataStub{
		CheckValidityCalled: func() error {
			return nil
		},
		IsForCurrentShardCalled: func() bool {
			return false
		},
		HashCalled: func() []byte {
			return []byte("hash")
		},
	}

	interceptedDataWithErr := &testscommon.InterceptedDataStub{
		CheckValidityCalled: func() error {
			return errors.New("error")
		},
		IsForCurrentShardCalled: func() bool {
			return false
		},
		HashCalled: func() []byte {
			return []byte("hash")
		},
	}

	verifier := defaultInterceptedDataVerifier(defaultSpan)

	err := verifier.Verify(interceptedDataWithErr)
	require.Equal(t, ErrInvalidInterceptedData, err)

	err = verifier.Verify(interceptedData)
	require.NotNil(t, ErrInvalidInterceptedData, err)

	<-time.After(defaultSpan)

	err = verifier.Verify(interceptedData)
	require.Nil(t, err)
}
