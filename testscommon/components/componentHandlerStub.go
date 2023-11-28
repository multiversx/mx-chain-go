package components

// ComponentHandlerStub -
type ComponentHandlerStub struct {
	Name                     string
	CreateCalled             func() error
	CloseCalled              func() error
	CheckSubcomponentsCalled func() error
}

// NewComponentHandlerStub -
func NewComponentHandlerStub(name string) *ComponentHandlerStub {
	return &ComponentHandlerStub{
		Name: name,
	}
}

// Create -
func (stub *ComponentHandlerStub) Create() error {
	if stub.CreateCalled != nil {
		return stub.CreateCalled()
	}

	return nil
}

// Close -
func (stub *ComponentHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// CheckSubcomponents -
func (stub *ComponentHandlerStub) CheckSubcomponents() error {
	if stub.CheckSubcomponentsCalled != nil {
		return stub.CheckSubcomponentsCalled()
	}

	return nil
}

// String -
func (stub *ComponentHandlerStub) String() string {
	return stub.Name
}
