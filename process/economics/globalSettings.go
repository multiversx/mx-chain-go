package economics

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"math"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type inflationForEpochCompound struct {
	enableEpoch  uint32
	maxInflation float64
}

type globalSettingsHandler struct {
	minInflation                 float64
	yearSettings                 map[uint32]*config.YearSetting
	chainParameters              process.ChainParametersHandler
	tailInflationActivationEpoch uint32
	startYearInflation           float64
	inflationForEpoch            []inflationForEpochCompound
	decayPercentage              float64
	mutYearSettings              sync.RWMutex
}

var numMillisecondsPerSeconds = uint64(1000)
var numSecondsPerMinute = uint64(60)
var numMinutesPerHour = uint64(60)
var numHoursPerDay = uint64(24)
var numDaysInYear = uint64(365)
var numberOfMillisecondsInYear = numMillisecondsPerSeconds * numSecondsPerMinute * numMinutesPerHour * numHoursPerDay * numDaysInYear

// newGlobalSettingsHandler creates a new global settings provider
func newGlobalSettingsHandler(
	economics *config.EconomicsConfig,
	chainParameters process.ChainParametersHandler,
) (*globalSettingsHandler, error) {
	if check.IfNil(chainParameters) {
		return nil, process.ErrNilChainParametersHandler
	}

	g := &globalSettingsHandler{
		minInflation:                 economics.GlobalSettings.MinimumInflation,
		yearSettings:                 make(map[uint32]*config.YearSetting),
		tailInflationActivationEpoch: economics.GlobalSettings.TailInflation.EnableEpoch,
		startYearInflation:           economics.GlobalSettings.TailInflation.StartYearInflation,
		decayPercentage:              economics.GlobalSettings.TailInflation.DecayPercentage,
		mutYearSettings:              sync.RWMutex{},
		chainParameters:              chainParameters,
	}

	g.yearSettings = make(map[uint32]*config.YearSetting)
	for _, yearSetting := range economics.GlobalSettings.YearSettings {
		g.yearSettings[yearSetting.Year] = &config.YearSetting{
			Year:             yearSetting.Year,
			MaximumInflation: yearSetting.MaximumInflation,
		}
	}

	err := g.calculateInflationForEpochCompound()
	if err != nil {
		return nil, err
	}

	if isPercentageInvalid(g.minInflation) ||
		isPercentageInvalid(g.startYearInflation) ||
		isPercentageInvalid(g.decayPercentage) {
		return nil, process.ErrInvalidInflationPercentages
	}

	return g, nil
}

func (g *globalSettingsHandler) calculateInflationForEpochCompound() error {
	allChainParams := g.chainParameters.AllChainParameters()
	if len(allChainParams) == 0 {
		return process.ErrEmptyChainParametersConfiguration
	}

	g.inflationForEpoch = make([]inflationForEpochCompound, len(allChainParams))

	for index, chainParams := range allChainParams {
		roundsPerEpoch := chainParams.RoundsPerEpoch
		roundDuration := chainParams.RoundDuration

		epochDurationInMilliseconds := roundDuration * uint64(roundsPerEpoch)
		if epochDurationInMilliseconds == 0 {
			return process.ErrZeroDurationForEpoch
		}

		numberOfEpochsPerYear := float64(numberOfMillisecondsInYear) / float64(epochDurationInMilliseconds)

		inflationForCompound := numberOfEpochsPerYear * (math.Pow(1.0+g.startYearInflation, 1.0/numberOfEpochsPerYear) - 1)
		g.inflationForEpoch[index] = inflationForEpochCompound{
			enableEpoch:  chainParams.EnableEpoch,
			maxInflation: inflationForCompound,
		}

		if isPercentageInvalid(inflationForCompound) {
			return process.ErrInvalidInflationPercentages
		}
	}

	sort.SliceStable(g.inflationForEpoch, func(i, j int) bool {
		return g.inflationForEpoch[i].enableEpoch > g.inflationForEpoch[j].enableEpoch
	})

	return nil
}

// TODO: implement decay, implement growth, calculations will change after supernova
func (g *globalSettingsHandler) maxInflationRate(year uint32, epoch uint32) float64 {
	if g.isTailInflationActive(epoch) {
		for _, inflationForEpoch := range g.inflationForEpoch {
			if inflationForEpoch.enableEpoch <= epoch {
				return inflationForEpoch.maxInflation
			}
		}
		return 0.0
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
	return epoch > g.tailInflationActivationEpoch
}
