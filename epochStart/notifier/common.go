package notifier

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/epochStart"
)

var _ epochStart.ActionHandler = (*handlerStruct)(nil)

// handlerStruct represents a struct which satisfies the SubscribeFunctionHandler interface
type handlerStruct struct {
	act     func(hdr data.HeaderHandler)
	prepare func(metaHeader data.HeaderHandler)
	id      uint32
}

// NewHandlerForEpochStart will return a struct which will satisfy the above interface
func NewHandlerForEpochStart(
	actionFunc func(hdr data.HeaderHandler),
	prepareFunc func(metaHeader data.HeaderHandler),
	id uint32,
) epochStart.ActionHandler {
	handler := handlerStruct{
		act:     actionFunc,
		prepare: prepareFunc,
		id:      id,
	}

	return &handler
}

// EpochStartPrepare will notify the subscriber to prepare for a start of epoch.
// The event can be triggered multiple times
func (hs *handlerStruct) EpochStartPrepare(metaHdr data.HeaderHandler, _ data.BodyHandler) {
	if hs.act != nil {
		hs.prepare(metaHdr)
	}
}

// EpochStartAction will notify the subscribed function if not nil
func (hs *handlerStruct) EpochStartAction(hdr data.HeaderHandler) {
	if hs.act != nil {
		hs.act(hdr)
	}
}

// NotifyOrder returns the notification order for a start of epoch event
func (hs *handlerStruct) NotifyOrder() uint32 {
	return hs.id
}
