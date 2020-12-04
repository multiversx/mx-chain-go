package mock

// GasScheduleSubscribeHandlerStub -
type GasScheduleSubscribeHandlerStub struct {
	GasScheduleChangeCalled func(gasSchedule map[string]map[string]uint64)
}

// GasScheduleChange -
func (g *GasScheduleSubscribeHandlerStub) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	if g.GasScheduleChangeCalled != nil {
		g.GasScheduleChangeCalled(gasSchedule)
	}
}

// IsInterfaceNil -
func (g *GasScheduleSubscribeHandlerStub) IsInterfaceNil() bool {
	return g == nil
}
