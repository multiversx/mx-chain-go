package notifier

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// MakeHandlerForEpochStart will return a struct which will satisfy the above interface
func MakeHandlerForEpochStart(funcForSubscription func(hdr data.HeaderHandler)) epochStart.EpochStartHandler {
	handler := handlerStruct{subscribedFunc: funcForSubscription}
	return &handler
}

// handlerStruct represents a struct which satisfies the SubscribeFunctionHandler interface
type handlerStruct struct {
	subscribedFunc func(hdr data.HeaderHandler)
}

// EpochStartAction will notify the subscribed function if not nil
func (hs *handlerStruct) EpochStartAction(hdr data.HeaderHandler) {
	if hs.subscribedFunc != nil {
		hs.subscribedFunc(hdr)
	}
}
