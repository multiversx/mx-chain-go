package economics

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type globalSettingsHandler struct {
	minInflation                 float64
	yearSettings                 map[uint32]*config.YearSetting
	tailInflationActivationEpoch uint32
	startYearInflation           float64
	decayPercentage              float64
	mutYearSettings              sync.RWMutex
	enableEpochHandler           common.EnableEpochsHandler
}

// newGlobalSettingsHandler creates a new global settings provider
func newGlobalSettingsHandler(
	economics *config.EconomicsConfig,
	enableEpochHandler common.EnableEpochsHandler,
) (*globalSettingsHandler, error) {
	if check.IfNil(enableEpochHandler) {
		return nil, process.ErrNilEpochHandler
	}

	g := &globalSettingsHandler{
		minInflation:                 economics.GlobalSettings.MinimumInflation,
		yearSettings:                 make(map[uint32]*config.YearSetting),
		tailInflationActivationEpoch: economics.GlobalSettings.TailInflation.EnableEpoch,
		startYearInflation:           economics.GlobalSettings.TailInflation.StartYearInflation,
		decayPercentage:              economics.GlobalSettings.TailInflation.DecayPercentage,
		mutYearSettings:              sync.RWMutex{},
		enableEpochHandler:           enableEpochHandler,
	}

	g.yearSettings = make(map[uint32]*config.YearSetting)
	for _, yearSetting := range economics.GlobalSettings.YearSettings {
		g.yearSettings[yearSetting.Year] = &config.YearSetting{
			Year:             yearSetting.Year,
			MaximumInflation: yearSetting.MaximumInflation,
		}
	}

	if isPercentageInvalid(g.minInflation) ||
		isPercentageInvalid(g.startYearInflation) ||
		isPercentageInvalid(g.decayPercentage) {
		return nil, process.ErrInvalidInflationPercentages
	}

	return g, nil
}

// TODO: implement decay, implement growth, calculations will change after supernova
func (g *globalSettingsHandler) maxInflationRate(year uint32) float64 {
	if g.tailInflationActivationEpoch < g.enableEpochHandler.GetCurrentEpoch() {
		return g.startYearInflation
	}

	g.mutYearSettings.RLock()
	yearSetting, ok := g.yearSettings[year]
	g.mutYearSettings.RUnlock()

	if !ok {
		return g.minInflation
	}

	return yearSetting.MaximumInflation
}
