package forking

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
)

type roundNotifier struct {
	round    uint64
	handlers []process.RoundSubscriberHandler
	mutex    sync.RWMutex
}

// NewRoundNotifier creates a new instance of interface type process.RoundNotifier
func NewRoundNotifier() process.RoundNotifier {
	return &roundNotifier{
		handlers: make([]process.RoundSubscriberHandler, 0),
	}
}

// RegisterNotifyHandler will register the provided handler to be called whenever a new round has changed
func (rn *roundNotifier) RegisterNotifyHandler(handler process.RoundSubscriberHandler) {
	if check.IfNil(handler) {
		return
	}

	rn.mutex.Lock()
	rn.handlers = append(rn.handlers, handler)
	handler.RoundConfirmed(rn.round)
	rn.mutex.Unlock()
}

// CheckRound should be called whenever a new round is known. It will trigger the notifications of the
// registered handlers only if the current stored round is different from the one provided
func (rn *roundNotifier) CheckRound(round uint64) {
	rn.mutex.Lock()
	defer rn.mutex.Unlock()

	if rn.round == round {
		return
	}
	rn.round = round

	for _, handler := range rn.handlers {
		handler.RoundConfirmed(round)
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (rn *roundNotifier) IsInterfaceNil() bool {
	return rn == nil
}
