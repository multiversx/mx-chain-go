package forking

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type genericRoundNotifier struct {
	mutData          sync.RWMutex
	wasInitialized   bool
	currentRound     uint64
	currentTimestamp uint64
	mutHandler       sync.RWMutex
	handlers         []vmcommon.RoundSubscriberHandler
}

// NewGenericRoundNotifier creates a new instance of a genericRoundNotifier component
func NewGenericRoundNotifier() *genericRoundNotifier {
	return &genericRoundNotifier{
		wasInitialized: false,
		handlers:       make([]vmcommon.RoundSubscriberHandler, 0),
	}
}

// CheckRound should be called whenever a new Round is known. It will trigger the notifications of the registered handlers
// only if the current stored Round is different from the one provided
func (gen *genericRoundNotifier) CheckRound(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	gen.mutData.Lock()
	Round := header.GetRound()
	timestamp := header.GetTimeStamp()
	shouldSkipHeader := gen.wasInitialized && gen.currentRound == Round
	if shouldSkipHeader {
		gen.mutData.Unlock()

		return
	}
	gen.wasInitialized = true
	gen.currentRound = Round
	gen.currentTimestamp = timestamp
	gen.mutData.Unlock()

	gen.mutHandler.RLock()
	handlersCopy := make([]vmcommon.RoundSubscriberHandler, len(gen.handlers))
	copy(handlersCopy, gen.handlers)
	gen.mutHandler.RUnlock()

	log.Debug("genericRoundNotifier.NotifyRoundChangeConfirmed",
		"new Round", Round,
		"new Round at timestamp", timestamp,
		"num handlers", len(handlersCopy),
	)

	for _, handler := range handlersCopy {
		handler.RoundConfirmed(Round, timestamp)
	}
}

// RegisterNotifyHandler will register the provided handler to be called whenever a new Round has changed
func (gen *genericRoundNotifier) RegisterNotifyHandler(handler vmcommon.RoundSubscriberHandler) {
	if check.IfNil(handler) {
		return
	}

	gen.mutHandler.Lock()
	gen.handlers = append(gen.handlers, handler)
	gen.mutHandler.Unlock()

	Round, timestamp := gen.getRoundTimestamp()
	handler.RoundConfirmed(Round, timestamp)
}

func (gen *genericRoundNotifier) getRoundTimestamp() (uint64, uint64) {
	gen.mutData.RLock()
	defer gen.mutData.RUnlock()

	return gen.currentRound, gen.currentTimestamp
}

// CurrentRound returns the stored Round. Useful when the Round has already changed and a new component needs to be
// created.
func (gen *genericRoundNotifier) CurrentRound() uint64 {
	Round, _ := gen.getRoundTimestamp()

	return Round
}

// UnRegisterAll removes all registered handlers queue
func (gen *genericRoundNotifier) UnRegisterAll() {
	gen.mutHandler.Lock()
	gen.handlers = make([]vmcommon.RoundSubscriberHandler, 0)
	gen.mutHandler.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (gen *genericRoundNotifier) IsInterfaceNil() bool {
	return gen == nil
}
