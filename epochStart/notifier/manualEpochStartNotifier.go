package notifier

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

var log = logger.GetOrCreate("epochstart/notifier")

// manualEpochStartNotifier will notice all recorded handlers to a provided epoch (from other components)
// TODO think about a better name as this does not get its data from the exterior of the node
type manualEpochStartNotifier struct {
	mutHandlers     sync.RWMutex
	handlers        []epochStart.ActionHandler
	mutCurrentEpoch sync.RWMutex
	currentEpoch    uint32
}

// NewManualEpochStartNotifier creates a new instance of a manual epoch start notifier
func NewManualEpochStartNotifier() *manualEpochStartNotifier {
	return &manualEpochStartNotifier{
		handlers: make([]epochStart.ActionHandler, 0),
	}
}

// RegisterHandler registers an epoch start action handler
func (mesn *manualEpochStartNotifier) RegisterHandler(handler epochStart.ActionHandler) {
	mesn.mutHandlers.Lock()
	defer mesn.mutHandlers.Unlock()

	mesn.handlers = append(mesn.handlers, handler)
}

// NewEpoch signals that a new epoch event has occurred
func (mesn *manualEpochStartNotifier) NewEpoch(epoch uint32) {
	mesn.mutCurrentEpoch.Lock()
	if mesn.currentEpoch >= epoch {
		mesn.mutCurrentEpoch.Unlock()
		return
	}
	mesn.currentEpoch = epoch
	mesn.mutCurrentEpoch.Unlock()

	log.Info("manualEpochStartNotifier.NewEpoch", "epoch", epoch)

	mesn.mutHandlers.RLock()
	defer mesn.mutHandlers.RUnlock()

	for _, handler := range mesn.handlers {
		hdr := &block.Header{
			Epoch: epoch,
		}
		handler.EpochStartAction(hdr)
	}
}

// CurrentEpoch returns the current epoch saved
func (mesn *manualEpochStartNotifier) CurrentEpoch() uint32 {
	mesn.mutCurrentEpoch.RLock()
	defer mesn.mutCurrentEpoch.RUnlock()

	return mesn.currentEpoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (mesn *manualEpochStartNotifier) IsInterfaceNil() bool {
	return mesn == nil
}
