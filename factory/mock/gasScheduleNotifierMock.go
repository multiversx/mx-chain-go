package mock

import "github.com/ElrondNetwork/elrond-go/core"

// GasScheduleNotifierMock -
type GasScheduleNotifierMock struct {
	GasSchedule map[string]map[string]uint64
}

// NewGasScheduleNotifierMock -
func NewGasScheduleNotifierMock(gasSchedule map[string]map[string]uint64) *GasScheduleNotifierMock {
	g := &GasScheduleNotifierMock{
		GasSchedule: gasSchedule,
	}
	return g
}

// RegisterNotifyHandler -
func (g *GasScheduleNotifierMock) RegisterNotifyHandler(handler core.GasScheduleSubscribeHandler) {
	handler.GasScheduleChange(g.GasSchedule)
}

// LatestGasSchedule -
func (g *GasScheduleNotifierMock) LatestGasSchedule() map[string]map[string]uint64 {
	return g.GasSchedule
}

// UnRegisterAll -
func (g *GasScheduleNotifierMock) UnRegisterAll() {
}

// IsInterfaceNil -
func (g *GasScheduleNotifierMock) IsInterfaceNil() bool {
	return g == nil
}
