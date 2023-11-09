package mock

import "github.com/multiversx/mx-chain-core-go/core"

// GasScheduleNotifierMock -
type GasScheduleNotifierMock struct {
	GasSchedule map[string]map[string]uint64
	Handlers    []core.GasScheduleSubscribeHandler
}

// NewGasScheduleNotifierMock -
func NewGasScheduleNotifierMock(gasSchedule map[string]map[string]uint64) *GasScheduleNotifierMock {
	g := &GasScheduleNotifierMock{
		GasSchedule: gasSchedule,
		Handlers:    make([]core.GasScheduleSubscribeHandler, 0),
	}
	return g
}

// RegisterNotifyHandler -
func (g *GasScheduleNotifierMock) RegisterNotifyHandler(handler core.GasScheduleSubscribeHandler) {
	handler.GasScheduleChange(g.GasSchedule)
	g.Handlers = append(g.Handlers, handler)
}

// LatestGasSchedule -
func (g *GasScheduleNotifierMock) LatestGasSchedule() map[string]map[string]uint64 {
	return g.GasSchedule
}

// UnRegisterAll -
func (g *GasScheduleNotifierMock) UnRegisterAll() {
}

// ChangeGasSchedule -
func (g *GasScheduleNotifierMock) ChangeGasSchedule(gasSchedule map[string]map[string]uint64) {
	g.GasSchedule = gasSchedule
	for _, handler := range g.Handlers {
		handler.GasScheduleChange(g.GasSchedule)
	}
}

// IsInterfaceNil -
func (g *GasScheduleNotifierMock) IsInterfaceNil() bool {
	return g == nil
}
