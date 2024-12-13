package testscommon

import "github.com/multiversx/mx-chain-core-go/core"

// GasScheduleNotifierMock -
type GasScheduleNotifierMock struct {
	GasSchedule                 map[string]map[string]uint64
	RegisterNotifyHandlerCalled func(handler core.GasScheduleSubscribeHandler)
	LatestGasScheduleCalled     func() map[string]map[string]uint64
	LatestGasScheduleCopyCalled func() map[string]map[string]uint64
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
	if g.RegisterNotifyHandlerCalled != nil {
		g.RegisterNotifyHandlerCalled(handler)
		return
	}

	handler.GasScheduleChange(g.LatestGasSchedule())
}

// LatestGasSchedule -
func (g *GasScheduleNotifierMock) LatestGasSchedule() map[string]map[string]uint64 {
	if g.LatestGasScheduleCalled != nil {
		return g.LatestGasScheduleCalled()
	}

	return g.GasSchedule
}

// LatestGasScheduleCopy -
func (g *GasScheduleNotifierMock) LatestGasScheduleCopy() map[string]map[string]uint64 {
	if g.LatestGasScheduleCopyCalled != nil {
		return g.LatestGasScheduleCopyCalled()
	}

	return g.GasSchedule
}

// UnRegisterAll -
func (g *GasScheduleNotifierMock) UnRegisterAll() {
}

// IsInterfaceNil -
func (g *GasScheduleNotifierMock) IsInterfaceNil() bool {
	return g == nil
}
