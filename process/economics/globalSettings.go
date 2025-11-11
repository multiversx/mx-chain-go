package economics

import (
	"math"
	"sync"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type globalSettingsHandler struct {
	minInflation                 float64
	yearSettings                 map[uint32]*config.YearSetting
	tailInflationActivationEpoch uint32
	startYearInflation           float64
	inflationForEpochCompound    float64
	decayPercentage              float64
	mutYearSettings              sync.RWMutex
}

const numberOfDaysInYear = 365.0

// newGlobalSettingsHandler creates a new global settings provider
func newGlobalSettingsHandler(
	economics *config.EconomicsConfig,
) (*globalSettingsHandler, error) {
	g := &globalSettingsHandler{
		minInflation:                 economics.GlobalSettings.MinimumInflation,
		yearSettings:                 make(map[uint32]*config.YearSetting),
		tailInflationActivationEpoch: economics.GlobalSettings.TailInflation.EnableEpoch,
		startYearInflation:           economics.GlobalSettings.TailInflation.StartYearInflation,
		decayPercentage:              economics.GlobalSettings.TailInflation.DecayPercentage,
		mutYearSettings:              sync.RWMutex{},
	}

	g.yearSettings = make(map[uint32]*config.YearSetting)
	for _, yearSetting := range economics.GlobalSettings.YearSettings {
		g.yearSettings[yearSetting.Year] = &config.YearSetting{
			Year:             yearSetting.Year,
			MaximumInflation: yearSetting.MaximumInflation,
		}
	}

	g.calculateInflationForEpochCompound()

	if isPercentageInvalid(g.minInflation) ||
		isPercentageInvalid(g.startYearInflation) ||
		isPercentageInvalid(g.decayPercentage) ||
		isPercentageInvalid(g.inflationForEpochCompound) {
		return nil, process.ErrInvalidInflationPercentages
	}

	return g, nil
}

func (g *globalSettingsHandler) calculateInflationForEpochCompound() {
	g.inflationForEpochCompound = numberOfDaysInYear * (math.Pow(1.0+g.startYearInflation, 1.0/numberOfDaysInYear) - 1)
}

// TODO: implement decay, implement growth, calculations will change after supernova
func (g *globalSettingsHandler) maxInflationRate(year uint32, epoch uint32) float64 {
	if g.isTailInflationActive(epoch) {
		return g.inflationForEpochCompound
	}

	return g.yearSettingsInflation(year)
}

func (g *globalSettingsHandler) yearSettingsInflation(year uint32) float64 {
	g.mutYearSettings.RLock()
	yearSetting, ok := g.yearSettings[year]
	g.mutYearSettings.RUnlock()

	if !ok {
		return g.minInflation
	}

	return yearSetting.MaximumInflation
}

func (g *globalSettingsHandler) isTailInflationActive(epoch uint32) bool {
	return epoch >= g.tailInflationActivationEpoch
}
