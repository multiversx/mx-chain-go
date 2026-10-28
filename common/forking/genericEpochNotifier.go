package forking

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("common/forking")

type genericEpochNotifier struct {
	mutData          sync.RWMutex
	wasInitialized   bool
	currentEpoch     uint32
	currentTimestamp uint64
	mutHandler       sync.RWMutex
	handlers         []vmcommon.EpochSubscriberHandler
}

// NewGenericEpochNotifier creates a new instance of a genericEpochNotifier component
func NewGenericEpochNotifier() *genericEpochNotifier {
	return &genericEpochNotifier{
		wasInitialized: false,
		handlers:       make([]vmcommon.EpochSubscriberHandler, 0),
	}
}

// CheckEpoch should be called whenever a new epoch is known. It will trigger the notifications of the registered handlers
// only if the current stored epoch is different from the one provided
func (gen *genericEpochNotifier) CheckEpoch(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	gen.mutData.Lock()
	epoch := getEpoch(header)
	timestamp := header.GetTimeStamp()
	shouldSkipHeader := gen.wasInitialized && gen.currentEpoch == epoch
	if shouldSkipHeader {
		gen.mutData.Unlock()

		return
	}
	gen.wasInitialized = true
	gen.currentEpoch = epoch
	gen.currentTimestamp = timestamp
	gen.mutData.Unlock()

	gen.mutHandler.RLock()
	handlersCopy := make([]vmcommon.EpochSubscriberHandler, len(gen.handlers))
	copy(handlersCopy, gen.handlers)
	gen.mutHandler.RUnlock()

	log.Debug("genericEpochNotifier.NotifyEpochChangeConfirmed",
		"new epoch", epoch,
		"new epoch at timestamp", timestamp,
		"num handlers", len(handlersCopy),
	)

	for _, handler := range handlersCopy {
		handler.EpochConfirmed(epoch, timestamp)
	}
}

func getEpoch(header data.HeaderHandler) uint32 {
	epoch := header.GetEpoch()
	if !header.IsHeaderV3() {
		return epoch
	}

	metaBlock, isMeta := header.(data.MetaHeaderHandler)
	if !isMeta {
		return epoch
	}

	if !metaBlock.IsEpochChangeProposed() {
		return epoch
	}

	// If it is epoch start proposed meta header, we should use next epoch
	return epoch + 1
}

// RegisterNotifyHandler will register the provided handler to be called whenever a new epoch has changed
func (gen *genericEpochNotifier) RegisterNotifyHandler(handler vmcommon.EpochSubscriberHandler) {
	if check.IfNil(handler) {
		return
	}

	gen.mutHandler.Lock()
	gen.handlers = append(gen.handlers, handler)
	gen.mutHandler.Unlock()

	epoch, timestamp := gen.getEpochTimestamp()
	handler.EpochConfirmed(epoch, timestamp)
}

func (gen *genericEpochNotifier) getEpochTimestamp() (uint32, uint64) {
	gen.mutData.RLock()
	defer gen.mutData.RUnlock()

	return gen.currentEpoch, gen.currentTimestamp
}

// CurrentEpoch returns the stored epoch. Useful when the epoch has already changed and a new component needs to be
// created.
func (gen *genericEpochNotifier) CurrentEpoch() uint32 {
	epoch, _ := gen.getEpochTimestamp()

	return epoch
}

// UnRegisterAll removes all registered handlers queue
func (gen *genericEpochNotifier) UnRegisterAll() {
	gen.mutHandler.Lock()
	gen.handlers = make([]vmcommon.EpochSubscriberHandler, 0)
	gen.mutHandler.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (gen *genericEpochNotifier) IsInterfaceNil() bool {
	return gen == nil
}
