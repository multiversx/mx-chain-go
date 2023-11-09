package forking

import (
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

// GasScheduleMap (alias) is the map for gas schedule
type GasScheduleMap = map[string]map[string]uint64

type gasScheduleNotifier struct {
	mutNotifier        sync.RWMutex
	configDir          string
	gasScheduleConfig  config.GasScheduleConfig
	currentEpoch       uint32
	lastGasSchedule    GasScheduleMap
	handlers           []core.GasScheduleSubscribeHandler
	wasmVMChangeLocker common.Locker
}

// ArgsNewGasScheduleNotifier defines the gas schedule notifier arguments
type ArgsNewGasScheduleNotifier struct {
	GasScheduleConfig  config.GasScheduleConfig
	ConfigDir          string
	EpochNotifier      process.EpochNotifier
	WasmVMChangeLocker common.Locker
}

// NewGasScheduleNotifier creates a new instance of a gasScheduleNotifier component
func NewGasScheduleNotifier(args ArgsNewGasScheduleNotifier) (*gasScheduleNotifier, error) {
	if len(args.GasScheduleConfig.GasScheduleByEpochs) == 0 {
		return nil, core.ErrInvalidGasScheduleConfig
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, core.ErrNilEpochStartNotifier
	}
	if check.IfNilReflect(args.WasmVMChangeLocker) {
		return nil, common.ErrNilWasmChangeLocker
	}

	g := &gasScheduleNotifier{
		gasScheduleConfig:  args.GasScheduleConfig,
		handlers:           make([]core.GasScheduleSubscribeHandler, 0),
		configDir:          args.ConfigDir,
		wasmVMChangeLocker: args.WasmVMChangeLocker,
	}
	log.Debug("gasSchedule: enable epoch for gas schedule directories paths epoch", "epoch", g.gasScheduleConfig.GasScheduleByEpochs)

	for _, gasScheduleConf := range g.gasScheduleConfig.GasScheduleByEpochs {
		_, err := common.LoadGasScheduleConfig(filepath.Join(g.configDir, gasScheduleConf.FileName))
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(g.gasScheduleConfig.GasScheduleByEpochs, func(i, j int) bool {
		return g.gasScheduleConfig.GasScheduleByEpochs[i].StartEpoch < g.gasScheduleConfig.GasScheduleByEpochs[j].StartEpoch
	})
	var err error
	g.lastGasSchedule, err = common.LoadGasScheduleConfig(filepath.Join(g.configDir, args.GasScheduleConfig.GasScheduleByEpochs[0].FileName))
	if err != nil {
		return nil, err
	}

	args.EpochNotifier.RegisterNotifyHandler(g)

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
func (g *gasScheduleNotifier) EpochConfirmed(epoch uint32, _ uint64) {
	old := atomic.SwapUint32(&g.currentEpoch, epoch)
	sameEpoch := old == epoch
	if sameEpoch {
		return
	}

	newGasSchedule := g.changeLatestGasSchedule(epoch, old)
	if newGasSchedule == nil {
		return
	}

	g.wasmVMChangeLocker.Lock()
	for _, handler := range g.handlers {
		if !check.IfNil(handler) {
			handler.GasScheduleChange(newGasSchedule)
		}
	}
	g.wasmVMChangeLocker.Unlock()
}

func (g *gasScheduleNotifier) changeLatestGasSchedule(epoch uint32, oldEpoch uint32) map[string]map[string]uint64 {
	g.mutNotifier.Lock()
	defer g.mutNotifier.Unlock()
	newVersion := g.getMatchingVersion(epoch)
	oldVersion := g.getMatchingVersion(oldEpoch)

	if newVersion.StartEpoch == oldVersion.StartEpoch {
		// gasSchedule is still the same
		return nil
	}

	newGasSchedule, err := common.LoadGasScheduleConfig(filepath.Join(g.configDir, newVersion.FileName))
	if err != nil {
		log.Error("could not load the new gas schedule")
		return nil
	}

	log.Debug("gasScheduleNotifier.EpochConfirmed new gas schedule",
		"new epoch", epoch,
		"num handlers", len(g.handlers),
	)

	g.lastGasSchedule = newGasSchedule

	return newGasSchedule
}

// LatestGasSchedule returns the latest gas schedule
func (g *gasScheduleNotifier) LatestGasSchedule() map[string]map[string]uint64 {
	g.mutNotifier.RLock()
	defer g.mutNotifier.RUnlock()

	return g.lastGasSchedule
}

// LatestGasScheduleCopy returns a copy of the latest gas schedule
func (g *gasScheduleNotifier) LatestGasScheduleCopy() map[string]map[string]uint64 {
	g.mutNotifier.RLock()
	defer g.mutNotifier.RUnlock()

	return copyLatestGasScheduleMap(g.lastGasSchedule)
}

func copyLatestGasScheduleMap(src map[string]map[string]uint64) map[string]map[string]uint64 {
	newMap := make(map[string]map[string]uint64)
	for key, innerMap := range src {
		newMap[key] = make(map[string]uint64)
		for innerKey, v := range innerMap {
			newMap[key][innerKey] = v
		}
	}

	return newMap
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *gasScheduleNotifier) IsInterfaceNil() bool {
	return g == nil
}
