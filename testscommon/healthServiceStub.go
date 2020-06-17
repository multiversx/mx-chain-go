package testscommon

// HealthServiceStub -
type HealthServiceStub struct {
	StartCalled            func()
	MonitorComponentCalled func(component interface{})
}

// NewHealthServiceStub -
func NewHealthServiceStub() *HealthServiceStub {
	return &HealthServiceStub{}
}

// Start -
func (stub *HealthServiceStub) Start() {
	if stub.StartCalled != nil {
		stub.StartCalled()
	}
}

// MonitorComponent -
func (stub *HealthServiceStub) MonitorComponent(component interface{}) {
	if stub.MonitorComponentCalled != nil {
		stub.MonitorComponentCalled(component)
	}
}

// IsInterfaceNil -
func (stub *HealthServiceStub) IsInterfaceNil() bool {
	return stub == nil
}
