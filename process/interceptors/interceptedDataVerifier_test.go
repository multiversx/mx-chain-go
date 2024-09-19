package interceptors

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestInterceptedDataVerifier_CheckValidity(t *testing.T) {
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

	span := 1 * time.Second
	c, _ := cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: span,
		CacheExpiry: time.Second,
	})

	verifier := NewInterceptedDataVerifier(c)

	err := verifier.Verify(interceptedData)
	require.Nil(t, err)

	wg := sync.WaitGroup{}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		err := verifier.Verify(interceptedData)
		if err != nil {
			return
		}
	}
}
