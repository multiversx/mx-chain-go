package economics

import (
	"math"
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type globalSettingsHandler struct {
	chainParametersHandler       common.ChainParametersHandler
	minInflation                 float64
	yearSettings                 map[uint32]*config.YearSetting
	tailInflationActivationEpoch uint32
	startYearInflation           float64
	decayPercentage              float64
	mutYearSettings              sync.RWMutex
}

var numMillisecondsPerSeconds = uint64(1000)
var numSecondsPerMinute = uint64(60)
var numMinutesPerHour = uint64(60)
var numHoursPerDay = uint64(24)
var numDaysInYear = uint64(365)

// newGlobalSettingsHandler creates a new global settings provider
func newGlobalSettingsHandler(
	economics *config.EconomicsConfig,
	chainParametersHandler common.ChainParametersHandler,
) (*globalSettingsHandler, error) {
	g := &globalSettingsHandler{
		chainParametersHandler:       chainParametersHandler,
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

	if isPercentageInvalid(g.minInflation) ||
		isPercentageInvalid(g.startYearInflation) ||
		isPercentageInvalid(g.decayPercentage) {
		return nil, process.ErrInvalidInflationPercentages
	}

	return g, nil
}

func (g *globalSettingsHandler) calculateInflationForEpochCompound(epoch uint32) (float64, error) {
	chainParameters, err := g.chainParametersHandler.ChainParametersForEpoch(epoch)
	if err != nil {
		return 0, err
	}

	numberOfMillisecondsInYear := numMillisecondsPerSeconds * numSecondsPerMinute * numMinutesPerHour * numHoursPerDay * numDaysInYear
	epochDurationInMilliseconds := chainParameters.RoundDuration * uint64(chainParameters.RoundsPerEpoch)

	if epochDurationInMilliseconds == 0 {
		return 0, process.ErrZeroDurationForEpoch
	}

	numberOfEpochsPerYear := float64(numberOfMillisecondsInYear) / float64(epochDurationInMilliseconds)

	inflationForEpochCompound := numberOfEpochsPerYear * (math.Pow(1.0+g.startYearInflation, 1.0/numberOfEpochsPerYear) - 1)

	if isPercentageInvalid(inflationForEpochCompound) {
		return 0, process.ErrInvalidInflationPercentages
	}

	return inflationForEpochCompound, nil
}

// TODO: implement decay, implement growth, calculations will change after supernova
func (g *globalSettingsHandler) maxInflationRate(year uint32, epoch uint32) (float64, error) {
	if g.isTailInflationActive(epoch) {
		return g.calculateInflationForEpochCompound(epoch)
	}

	return g.yearSettingsInflation(year), nil
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
	return epoch > g.tailInflationActivationEpoch
}
