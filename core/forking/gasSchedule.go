package forking

import (
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

// GasScheduleMap (alias) is the map for gas schedule
type GasScheduleMap = map[string]map[string]uint64

type gasScheduleNotifier struct {
	gasScheduleConfig config.GasScheduleConfig

	currentEpoch    uint32
	lastGasSchedule GasScheduleMap

	mutNotifier sync.RWMutex

	handlers []core.GasScheduleSubscribeHandler
}

// NewGasScheduleNotifier creates a new instance of a gasScheduleNotifier component
func NewGasScheduleNotifier(gasScheduleConfig config.GasScheduleConfig) *gasScheduleNotifier {
	return &gasScheduleNotifier{
		gasScheduleConfig: gasScheduleConfig,
		handlers:          make([]core.GasScheduleSubscribeHandler, 0),
	}
}

// RegisterNotifyHandler will register the provided handler to be called whenever a new epoch has changed
func (g *gasScheduleNotifier) RegisterNotifyHandler(handler core.GasScheduleSubscribeHandler) {
	if check.IfNil(handler) {
		return
	}

	g.mutNotifier.Lock()
	g.handlers = append(g.handlers, handler)
	g.mutNotifier.Unlock()

	handler.EpochConfirmed(atomic.LoadUint32(&g.currentEpoch))
}

// UnRegisterAll removes all registered handlers queue
func (g *gasScheduleNotifier) UnRegisterAll() {
	g.mutNotifier.Lock()
	g.handlers = make([]core.GasScheduleSubscribeHandler, 0)
	g.mutNotifier.Unlock()
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (g *gasScheduleNotifier) EpochConfirmed(epoch uint32) {
	old := atomic.SwapUint32(&g.currentEpoch, epoch)
	sameEpoch := old == epoch
	if sameEpoch {
		return
	}

	g.mutNotifier.RLock()

	handlersCopy := make([]core.GasScheduleSubscribeHandler, len(g.handlers))
	copy(handlersCopy, g.handlers)
	g.mutNotifier.RUnlock()

	log.Debug("genericEpochNotifier.NotifyEpochChangeConfirmed",
		"new epoch", epoch,
		"num handlers", len(handlersCopy),
	)

	for _, handler := range handlersCopy {
		handler.GasScheduleChanged(epoch)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *gasScheduleNotifier) IsInterfaceNil() bool {
	return g == nil
}
