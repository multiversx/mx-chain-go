package sync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/random"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

func TestNewConcurrentTriesMap(t *testing.T) {
	t.Parallel()

	ctm := newConcurrentTriesMap()
	require.NotNil(t, ctm)
}

func TestConcurrentTriesMap_ConcurrentAccesses(t *testing.T) {
	t.Parallel()

	// when running with -race, this should not throw race conditions. If we remove the mutex protection inside the struct,
	// this test will fail
	testDuration := 50 * time.Millisecond
	rnd := random.ConcurrentSafeIntRandomizer{}
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	ctm := newConcurrentTriesMap()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
			default:
				randomID := rnd.Intn(100)
				ctm.setTrie(fmt.Sprintf("%d", randomID), &trieMock.TrieStub{})
			}
		}
	}(ctx)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
			default:
				randomID := rnd.Intn(100)
				ctm.getTrie(fmt.Sprintf("%d", randomID))
			}
		}
	}(ctx)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
			default:
				_ = ctm.getTries()
			}
		}
	}(ctx)
	time.Sleep(testDuration)
	cancel()
}
