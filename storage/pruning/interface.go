package pruning

import (
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
)

// EpochStartNotifier defines
type EpochStartNotifier interface {
	RegisterHandler(handler notifier.SubscribeFunctionHandler)
	UnregisterHandler(handler notifier.SubscribeFunctionHandler)
	IsInterfaceNil() bool
}
