package economics

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestGlobalSettingsHandler_maxInflationRate(t *testing.T) {
	t.Parallel()

	tailInflationActivationEpoch := uint32(100)
	startYearInflation := 0.02
	year1Inflation := 0.1
	minInflation := 0.01
	enableEpochHandler := &mock.EnableEpochsHandlerMock{}
	gsh := &globalSettingsHandler{
		minInflation:                 minInflation,
		yearSettings:                 make(map[uint32]*config.YearSetting),
		tailInflationActivationEpoch: tailInflationActivationEpoch,
		startYearInflation:           startYearInflation,
		decayPercentage:              0.0025,
		mutYearSettings:              sync.RWMutex{},
		enableEpochHandler:           enableEpochHandler,
	}
	gsh.yearSettings[1] = &config.YearSetting{Year: 1, MaximumInflation: year1Inflation}

	// Test before tail inflation activation
	enableEpochHandler.CurrentEpoch = tailInflationActivationEpoch - 1
	rate := gsh.maxInflationRate(1)
	require.Equal(t, year1Inflation, rate)

	// Test at tail inflation activation
	enableEpochHandler.CurrentEpoch = tailInflationActivationEpoch
	rate = gsh.maxInflationRate(1)
	require.Equal(t, year1Inflation, rate)

	// Test after tail inflation activation
	enableEpochHandler.CurrentEpoch = tailInflationActivationEpoch + 1
	rate = gsh.maxInflationRate(1)
	require.Equal(t, startYearInflation, rate)

	// Test with a year that is not in the map
	enableEpochHandler.CurrentEpoch = tailInflationActivationEpoch - 1
	rate = gsh.maxInflationRate(2)
	require.Equal(t, minInflation, rate)
}
