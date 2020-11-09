package forking

import (
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

// GasScheduleMap (alias) is the map for gas schedule
type GasScheduleMap = map[string]map[string]uint64

type gasScheduleNotifier struct {
	mutNotifier       sync.RWMutex
	configDir         string
	gasScheduleConfig config.GasScheduleConfig
	currentEpoch      uint32
	lastGasSchedule   GasScheduleMap
	handlers          []core.GasScheduleSubscribeHandler
}

// NewGasScheduleNotifier creates a new instance of a gasScheduleNotifier component
func NewGasScheduleNotifier(
	gasScheduleConfig config.GasScheduleConfig,
	configDir string,
	epochNotifier core.EpochNotifier,
) (*gasScheduleNotifier, error) {
	if len(gasScheduleConfig.GasScheduleByEpochs) == 0 {
		return nil, core.ErrInvalidGasScheduleConfigs
	}
	if check.IfNil(epochNotifier) {
		return nil, core.ErrNilEpochStartNotifier
	}

	g := &gasScheduleNotifier{
		gasScheduleConfig: gasScheduleConfig,
		handlers:          make([]core.GasScheduleSubscribeHandler, 0),
	}

	for _, gasScheduleConf := range g.gasScheduleConfig.GasScheduleByEpochs {
		_, err := core.LoadGasScheduleConfig(filepath.Join(configDir, gasScheduleConf.FileName))
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(g.gasScheduleConfig.GasScheduleByEpochs, func(i, j int) bool {
		return g.gasScheduleConfig.GasScheduleByEpochs[i].StartEpoch < g.gasScheduleConfig.GasScheduleByEpochs[j].StartEpoch
	})
	var err error
	g.lastGasSchedule, err = core.LoadGasScheduleConfig(filepath.Join(configDir, gasScheduleConfig.GasScheduleByEpochs[0].FileName))
	if err != nil {
		return nil, err
	}

	epochNotifier.RegisterNotifyHandler(g)

	return g, nil
}

// RegisterNotifyHandler will register the provided handler to be called whenever a new epoch has changed
func (g *gasScheduleNotifier) RegisterNotifyHandler(handler core.GasScheduleSubscribeHandler) {
	if check.IfNilReflect(handler) {
		return
	}

	g.mutNotifier.Lock()
	g.handlers = append(g.handlers, handler)
	handler.GasScheduleChange(g.lastGasSchedule)
	g.mutNotifier.Unlock()
}

// UnRegisterAll removes all registered handlers queue
func (g *gasScheduleNotifier) UnRegisterAll() {
	g.mutNotifier.Lock()
	g.handlers = make([]core.GasScheduleSubscribeHandler, 0)
	g.mutNotifier.Unlock()
}

func (g *gasScheduleNotifier) getMatchingVersion(epoch uint32) config.GasScheduleByEpochs {
	currentVersion := g.gasScheduleConfig.GasScheduleByEpochs[0]
	for _, versionByEpoch := range g.gasScheduleConfig.GasScheduleByEpochs {
		if versionByEpoch.StartEpoch > epoch {
			break
		}

		currentVersion = versionByEpoch
	}
	return currentVersion
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (g *gasScheduleNotifier) EpochConfirmed(epoch uint32) {
	old := atomic.SwapUint32(&g.currentEpoch, epoch)
	sameEpoch := old == epoch
	if sameEpoch {
		return
	}

	g.mutNotifier.Lock()
	defer g.mutNotifier.Unlock()
	newVersion := g.getMatchingVersion(epoch)
	oldVersion := g.getMatchingVersion(old)

	if newVersion.StartEpoch == oldVersion.StartEpoch {
		// gasSchedule is still the same
		return
	}

	newGasSchedule, err := core.LoadGasScheduleConfig(filepath.Join(g.configDir, newVersion.FileName))
	if err != nil {
		log.Error("could not load the new gas schedule")
		return
	}

	log.Debug("gasScheduleNotifier.EpochConfirmed new gas schedule",
		"new epoch", epoch,
		"num handlers", len(g.handlers),
	)

	g.lastGasSchedule = newGasSchedule
	for _, handler := range g.handlers {
		handler.GasScheduleChange(g.lastGasSchedule)
	}
}

// LatestGasSchedule returns the latest gas schedule
func (g *gasScheduleNotifier) LatestGasSchedule() map[string]map[string]uint64 {
	g.mutNotifier.RLock()
	defer g.mutNotifier.RUnlock()
	return g.lastGasSchedule
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *gasScheduleNotifier) IsInterfaceNil() bool {
	return g == nil
}
