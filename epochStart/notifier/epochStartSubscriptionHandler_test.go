package notifier_test

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/stretchr/testify/assert"
)

func TestNewEpochStartSubscriptionHandler(t *testing.T) {
	t.Parallel()

	essh := notifier.NewEpochStartSubscriptionHandler()
	assert.NotNil(t, essh)
	assert.False(t, essh.IsInterfaceNil())
}

func TestEpochStartSubscriptionHandler_RegisterHandlerNilHandlerShouldNotAdd(t *testing.T) {
	t.Parallel()

	essh := notifier.NewEpochStartSubscriptionHandler()
	essh.RegisterHandler(nil)

	handlers, mutHandlers := essh.RegisteredHandlers()
	mutHandlers.RLock()
	assert.Equal(t, 0, len(handlers))
	mutHandlers.RUnlock()
}

func TestEpochStartSubscriptionHandler_RegisterHandlerOkHandlerShouldAdd(t *testing.T) {
	t.Parallel()

	essh := notifier.NewEpochStartSubscriptionHandler()
	handler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {}, nil, 0)

	essh.RegisterHandler(handler)

	handlers, mutHandlers := essh.RegisteredHandlers()
	mutHandlers.RLock()
	assert.Equal(t, 1, len(handlers))
	mutHandlers.RUnlock()
}

func TestEpochStartSubscriptionHandler_RegisterHandlerShouldNotAllowDuplicates(t *testing.T) {
	t.Parallel()

	essh := notifier.NewEpochStartSubscriptionHandler()
	handler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {}, nil, 0)

	essh.RegisterHandler(handler)
	essh.RegisterHandler(handler)

	handlers, mutHandlers := essh.RegisteredHandlers()
	mutHandlers.RLock()
	assert.Len(t, handlers, 1)
	mutHandlers.RUnlock()

	// check unregister twice to ensure there is no problem
	essh.UnregisterHandler(handler)
	essh.UnregisterHandler(handler)

	handlers, mutHandlers = essh.RegisteredHandlers()
	mutHandlers.RLock()
	assert.Len(t, handlers, 0)
	mutHandlers.RUnlock()
}

func TestEpochStartSubscriptionHandler_UnregisterHandlerNilHandlerShouldDoNothing(t *testing.T) {
	t.Parallel()

	essh := notifier.NewEpochStartSubscriptionHandler()

	// first register a handler
	handler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {}, nil, 0)
	essh.RegisterHandler(handler)

	// then try to unregister but a nil handler is given
	essh.UnregisterHandler(nil)
	handlers, mutHandlers := essh.RegisteredHandlers()
	mutHandlers.RLock()
	// length of the slice should still be 1
	assert.Equal(t, 1, len(handlers))
	mutHandlers.RUnlock()
}

func TestEpochStartSubscriptionHandler_UnregisterHandlerOkHandlerShouldRemove(t *testing.T) {
	t.Parallel()

	essh := notifier.NewEpochStartSubscriptionHandler()

	// first register a handler
	handler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {}, nil, 0)
	essh.RegisterHandler(handler)

	// then unregister the same handler
	essh.UnregisterHandler(handler)
	handlers, mutHandlers := essh.RegisteredHandlers()
	mutHandlers.RLock()
	// length of the slice should be 0 because the handler was unregistered
	assert.Equal(t, 0, len(handlers))
	mutHandlers.RUnlock()
}

func TestEpochStartSubscriptionHandler_NotifyAll(t *testing.T) {
	t.Parallel()

	calledHandlers := make(map[int]struct{})
	calledHandlersIndices := make([]int, 0)
	essh := notifier.NewEpochStartSubscriptionHandler()

	handler1 := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		calledHandlers[1] = struct{}{}
		calledHandlersIndices = append(calledHandlersIndices, 1)
	}, nil, 1)
	handler2 := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		calledHandlers[2] = struct{}{}
		calledHandlersIndices = append(calledHandlersIndices, 2)
	}, nil, 2)
	handler3 := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		calledHandlers[3] = struct{}{}
		calledHandlersIndices = append(calledHandlersIndices, 3)
	}, nil, 3)

	essh.RegisterHandler(handler2)
	essh.RegisterHandler(handler1)
	essh.RegisterHandler(handler3)

	// make sure that the handler were not called yet
	assert.Empty(t, calledHandlers)

	// now we call the NotifyAll method and all handlers should be called
	essh.NotifyAll(&block.Header{})
	assert.Len(t, calledHandlers, 3)
	assert.Equal(t, []int{1, 2, 3}, calledHandlersIndices)
}

func TestEpochStartSubscriptionHandler_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	handler := notifier.NewEpochStartSubscriptionHandler()

	numOperations := 500
	wg := sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			switch idx & 6 {
			case 0:
				handler.RegisterHandler(notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {}, func(hdr data.HeaderHandler) {}, 0))
			case 1:
				handler.UnregisterHandler(notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {}, func(hdr data.HeaderHandler) {}, 0))
			case 2:
				handler.NotifyAll(&block.Header{})
			case 3:
				handler.NotifyAllPrepare(&block.Header{}, &block.Body{})
			case 4:
				handler.NotifyEpochChangeConfirmed(uint32(idx + 1))
			case 5:
				handler.RegisterForEpochChangeConfirmed(func(epoch uint32) {})
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
