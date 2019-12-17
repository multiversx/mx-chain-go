package factory

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
)

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler notifier.SubscribeFunctionHandler)
	UnregisterHandler(handler notifier.SubscribeFunctionHandler)
	NotifyAll(hdr data.HeaderHandler)
	IsInterfaceNil() bool
}

//HeaderSigVerifierHandler is the interface needed to check a header if is correct
type HeaderSigVerifierHandler interface {
	VerifyRandSeed(header data.HeaderHandler) error
	VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error
	VerifySignature(header data.HeaderHandler) error
	IsInterfaceNil() bool
}