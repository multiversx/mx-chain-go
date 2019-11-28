package notifier

import "github.com/ElrondNetwork/elrond-go/data"

// SubscribeFunctionHandler defines what a struct which contain a handler function for epoch start should do
type SubscribeFunctionHandler interface {
	EpochStartAction(hdr data.HeaderHandler)
}

// MakeHandlerForEpochStart will return a struct which will satisfy the above interface
func MakeHandlerForEpochStart(f func(hdr data.HeaderHandler)) SubscribeFunctionHandler {
	t := handlerStruct{f: f}
	return &t
}

// handlerStruct represents a struct which satisfies the SubscribeFunctionHandler interface
type handlerStruct struct {
	f func(hdr data.HeaderHandler)
}

// EpochStartAction will notify the subscribed function if not nil
func (hs *handlerStruct) EpochStartAction(hdr data.HeaderHandler) {
	if hs.f != nil {
		hs.f(hdr)
	}
}
