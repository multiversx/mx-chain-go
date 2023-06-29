package forking

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
func (grn *genericRoundNotifier) CheckRound(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	grn.mutData.Lock()
	round := header.GetRound()
	timestamp := header.GetTimeStamp()
	shouldSkipHeader := grn.wasInitialized && grn.currentRound == round
	if shouldSkipHeader {
		grn.mutData.Unlock()

		return
	}
	grn.wasInitialized = true
	grn.currentRound = round
	grn.currentTimestamp = timestamp
	grn.mutData.Unlock()

	grn.mutHandler.RLock()
	handlersCopy := make([]vmcommon.RoundSubscriberHandler, len(grn.handlers))
	copy(handlersCopy, grn.handlers)
	grn.mutHandler.RUnlock()

	log.Debug("genericRoundNotifier.NotifyRoundChangeConfirmed",
		"new Round", round,
		"new Round at timestamp", timestamp,
		"num handlers", len(handlersCopy),
	)

	for _, handler := range handlersCopy {
		handler.RoundConfirmed(round, timestamp)
	}
}

// RegisterNotifyHandler will register the provided handler to be called whenever a new Round has changed
func (grn *genericRoundNotifier) RegisterNotifyHandler(handler vmcommon.RoundSubscriberHandler) {
	if check.IfNil(handler) {
		return
	}

	grn.mutHandler.Lock()
	grn.handlers = append(grn.handlers, handler)
	grn.mutHandler.Unlock()

	round, timestamp := grn.getRoundTimestamp()
	handler.RoundConfirmed(round, timestamp)
}

func (grn *genericRoundNotifier) getRoundTimestamp() (uint64, uint64) {
	grn.mutData.RLock()
	defer grn.mutData.RUnlock()

	return grn.currentRound, grn.currentTimestamp
}

// CurrentRound returns the stored Round. Useful when the Round has already changed and a new component needs to be
// created.
func (grn *genericRoundNotifier) CurrentRound() uint64 {
	round, _ := grn.getRoundTimestamp()

	return round
}

// UnRegisterAll removes all registered handlers queue
func (grn *genericRoundNotifier) UnRegisterAll() {
	grn.mutHandler.Lock()
	grn.handlers = make([]vmcommon.RoundSubscriberHandler, 0)
	grn.mutHandler.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (grn *genericRoundNotifier) IsInterfaceNil() bool {
	return grn == nil
}
