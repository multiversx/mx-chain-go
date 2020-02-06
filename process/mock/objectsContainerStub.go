package mock

// ObjectsContainerStub -
type ObjectsContainerStub struct {
	GetCalled     func(key string) (interface{}, error)
	AddCalled     func(key string, val interface{}) error
	ReplaceCalled func(key string, val interface{}) error
	RemoveCalled  func(key string)
	LenCalled     func() int
}

// Get -
func (ocs *ObjectsContainerStub) Get(key string) (interface{}, error) {
	return ocs.GetCalled(key)
}

// Add -
func (ocs *ObjectsContainerStub) Add(key string, val interface{}) error {
	return ocs.AddCalled(key, val)
}

// Replace -
func (ocs *ObjectsContainerStub) Replace(key string, val interface{}) error {
	return ocs.ReplaceCalled(key, val)
}

// Remove -
func (ocs *ObjectsContainerStub) Remove(key string) {
	ocs.RemoveCalled(key)
}

// Len -
func (ocs *ObjectsContainerStub) Len() int {
	return ocs.LenCalled()
}
