package testscommon

import "github.com/multiversx/mx-chain-core-go/core"

// GasScheduleNotifierMock -
type GasScheduleNotifierMock struct {
	GasSchedule                 map[string]map[string]uint64
	RegisterNotifyHandlerCalled func(handler core.GasScheduleSubscribeHandler)
	LatestGasScheduleCalled     func() map[string]map[string]uint64
	LatestGasScheduleCopyCalled func() map[string]map[string]uint64
	GasScheduleForEpochCalled   func(epoch uint32) (map[string]map[string]uint64, error)
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

	handler.GasScheduleChange(g.GasSchedule)
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

// GasScheduleForEpoch -
func (g *GasScheduleNotifierMock) GasScheduleForEpoch(epoch uint32) (map[string]map[string]uint64, error) {
	if g.GasScheduleForEpochCalled != nil {
		return g.GasScheduleForEpochCalled(epoch)
	}

	return g.GasSchedule, nil
}

// IsInterfaceNil -
func (g *GasScheduleNotifierMock) IsInterfaceNil() bool {
	return g == nil
}
