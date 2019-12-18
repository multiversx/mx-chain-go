package notifier

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// MakeHandlerForEpochStart will return a struct which will satisfy the above interface
func MakeHandlerForEpochStart(
	actionFunc func(hdr data.HeaderHandler),
	prepareFunc func(metaHeader data.HeaderHandler),
) epochStart.EpochStartHandler {
	handler := handlerStruct{
		act:     actionFunc,
		prepare: prepareFunc,
	}

	return &handler
}

// handlerStruct represents a struct which satisfies the SubscribeFunctionHandler interface
type handlerStruct struct {
	act     func(hdr data.HeaderHandler)
	prepare func(metaHeader data.HeaderHandler)
}

// EpochStartPrepare will notify the subscriber to prepare for a start of epoch.
// The event can be triggered multiple times
func (hs *handlerStruct) EpochStartPrepare(metaHeader data.HeaderHandler) {
	if hs.act != nil {
		hs.prepare(metaHeader)
	}
}

// EpochStartAction will notify the subscribed function if not nil
func (hs *handlerStruct) EpochStartAction(hdr data.HeaderHandler) {
	if hs.act != nil {
		hs.act(hdr)
	}
}
