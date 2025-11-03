package economics

import (
	"testing"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
)

type mockEnableEpochsHandler struct {
	currentEpoch uint32
}

func (m *mockEnableEpochsHandler) GetCurrentEpoch() uint32 {
	return m.currentEpoch
}

func (m *mockEnableEpochsHandler) GetActivationEpoch(feature core.EnableEpochFlag) uint32 {
	return 0
}

func (m *mockEnableEpochsHandler) IsFlagDefined(feature core.EnableEpochFlag) bool {
	return true
}

func (m *mockEnableEpochsHandler) IsFlagEnabled(feature core.EnableEpochFlag) bool {
	return true
}

func (m *mockEnableEpochsHandler) IsFlagEnabledInEpoch(feature core.EnableEpochFlag, epoch uint32) bool {
	return true
}

func (m *mockEnableEpochsHandler) IsInterfaceNil() bool {
	return m == nil
}

func TestGlobalSettingsHandler_maxInflationRate(t *testing.T) {
	t.Parallel()

	tailInflationActivationEpoch := uint32(100)
	startYearInflation := 0.02
	year1Inflation := 0.1
	minInflation := 0.01

	gsh := &globalSettingsHandler{
		minInflation:                 minInflation,
		yearSettings:                 make(map[uint32]*config.YearSetting),
		tailInflationActivationEpoch: tailInflationActivationEpoch,
		startYearInflation:           startYearInflation,
		decayPercentage:              0.0025,
		mutYearSettings:              sync.RWMutex{},
		enableEpochHandler:           &mockEnableEpochsHandler{},
	}
	gsh.yearSettings[1] = &config.YearSetting{Year: 1, MaximumInflation: year1Inflation}

	// Test before tail inflation activation
	gsh.enableEpochHandler.(*mockEnableEpochsHandler).currentEpoch = tailInflationActivationEpoch - 1
	rate := gsh.maxInflationRate(1)
	assert.Equal(t, year1Inflation, rate)

	// Test at tail inflation activation
	gsh.enableEpochHandler.(*mockEnableEpochsHandler).currentEpoch = tailInflationActivationEpoch
	rate = gsh.maxInflationRate(1)
	assert.Equal(t, year1Inflation, rate)

	// Test after tail inflation activation
	gsh.enableEpochHandler.(*mockEnableEpochsHandler).currentEpoch = tailInflationActivationEpoch + 1
	rate = gsh.maxInflationRate(1)
	assert.Equal(t, startYearInflation, rate)

	// Test with a year that is not in the map
	gsh.enableEpochHandler.(*mockEnableEpochsHandler).currentEpoch = tailInflationActivationEpoch - 1
	rate = gsh.maxInflationRate(2)
	assert.Equal(t, minInflation, rate)
}
