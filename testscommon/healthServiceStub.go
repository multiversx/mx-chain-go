package testscommon

// HealthServiceStub -
type HealthServiceStub struct {
	StartCalled             func()
	RegisterComponentCalled func(component interface{})
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

// RegisterComponent -
func (stub *HealthServiceStub) RegisterComponent(component interface{}) {
	if stub.RegisterComponentCalled != nil {
		stub.RegisterComponentCalled(component)
	}
}

// IsInterfaceNil -
func (stub *HealthServiceStub) IsInterfaceNil() bool {
	return stub == nil
}
