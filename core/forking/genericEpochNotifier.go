package forking

import (
	"sync"
	"sync/atomic"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
)

var log = logger.GetOrCreate("core/forking")

type genericEpochNotifier struct {
	currentEpoch uint32
	mutHandler   sync.RWMutex
	handlers     []core.EpochNotifiedHandler
}

// NewGenericEpochNotifier creates a new instance of a genericEpochNotifier component
func NewGenericEpochNotifier() *genericEpochNotifier {
	return &genericEpochNotifier{
		handlers: make([]core.EpochNotifiedHandler, 0),
	}
}

// NotifyEpochChangeConfirmed should be called whenever a new start of epoch block becomes final
// This call will trigger the calls to all registered handlers
func (gen *genericEpochNotifier) NotifyEpochChangeConfirmed(epoch uint32) {
	old := atomic.SwapUint32(&gen.currentEpoch, epoch)
	sameEpoch := old == epoch
	if sameEpoch {
		return
	}

	gen.mutHandler.RLock()
	handlersCopy := make([]core.EpochNotifiedHandler, len(gen.handlers))
	copy(handlersCopy, gen.handlers)
	gen.mutHandler.RUnlock()

	log.Debug("genericEpochNotifier.NotifyEpochChangeConfirmed",
		"new epoch", epoch,
		"num handlers", len(handlersCopy),
	)

	for _, handler := range gen.handlers {
		handler.EpochConfirmed(epoch)
	}
}

// NotifyAll does nothing
func (gen *genericEpochNotifier) NotifyAll(_ data.HeaderHandler) {
}

// NotifyAllPrepare does nothing
func (gen *genericEpochNotifier) NotifyAllPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// RegisterNotifyHandler will register the provided handler to be called whenever a new epoch has changed
func (gen *genericEpochNotifier) RegisterNotifyHandler(handler core.EpochNotifiedHandler) {
	if check.IfNil(handler) {
		return
	}

	gen.mutHandler.Lock()
	gen.handlers = append(gen.handlers, handler)
	gen.mutHandler.Unlock()

	handler.EpochConfirmed(atomic.LoadUint32(&gen.currentEpoch))
}

// CurrentEpoch returns the stored epoch. Useful when the epoch has already changed and a new component needs to be
// created.
func (gen *genericEpochNotifier) CurrentEpoch() uint32 {
	return atomic.LoadUint32(&gen.currentEpoch)
}

// UnRegisterAll removes all registered handlers queue
func (gen *genericEpochNotifier) UnRegisterAll() {
	gen.mutHandler.Lock()
	gen.handlers = make([]core.EpochNotifiedHandler, 0)
	gen.mutHandler.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (gen *genericEpochNotifier) IsInterfaceNil() bool {
	return gen == nil
}
