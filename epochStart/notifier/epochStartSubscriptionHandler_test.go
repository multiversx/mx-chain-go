package notifier_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
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
	handler := notifier.MakeHandlerForEpochStart(func(hdr data.HeaderHandler) {})

	essh.RegisterHandler(handler)

	handlers, mutHandlers := essh.RegisteredHandlers()
	mutHandlers.RLock()
	assert.Equal(t, 1, len(handlers))
	mutHandlers.RUnlock()
}

func TestEpochStartSubscriptionHandler_UnregisterHandlerNilHandlerShouldDoNothing(t *testing.T) {
	t.Parallel()

	essh := notifier.NewEpochStartSubscriptionHandler()

	// first register a handler
	handler := notifier.MakeHandlerForEpochStart(func(hdr data.HeaderHandler) {})
	essh.RegisterHandler(handler)

	// then try to unregister but a nil handler is given
	essh.UnregisterHandler(nil)
	handlers, mutHandlers := essh.RegisteredHandlers()
	mutHandlers.RLock()
	// length of the slice should still be 1
	assert.Equal(t, 1, len(handlers))
	mutHandlers.RUnlock()
}

func TestEpochStartSubscriptionHandler_UnregisterHandlerOklHandlerShouldRemove(t *testing.T) {
	t.Parallel()

	essh := notifier.NewEpochStartSubscriptionHandler()

	// first register a handler
	handler := notifier.MakeHandlerForEpochStart(func(hdr data.HeaderHandler) {})
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

	firstHandlerWasCalled := false
	secondHandlerWasCalled := false
	essh := notifier.NewEpochStartSubscriptionHandler()

	// register 2 handlers
	handler1 := notifier.MakeHandlerForEpochStart(func(hdr data.HeaderHandler) {
		firstHandlerWasCalled = true
	})
	handler2 := notifier.MakeHandlerForEpochStart(func(hdr data.HeaderHandler) {
		secondHandlerWasCalled = true
	})

	essh.RegisterHandler(handler1)
	essh.RegisterHandler(handler2)

	// make sure that the handler were not called yet
	assert.False(t, firstHandlerWasCalled)
	assert.False(t, secondHandlerWasCalled)

	// now we call the NotifyAll method and all handlers should be called
	essh.NotifyAll(&block.Header{})
	assert.True(t, firstHandlerWasCalled)
	assert.True(t, secondHandlerWasCalled)
}
