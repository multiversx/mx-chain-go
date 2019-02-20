package mock

type ObjectsContainerStub struct {
	GetCalled     func(key string) (interface{}, error)
	AddCalled     func(key string, val interface{}) error
	ReplaceCalled func(key string, val interface{}) error
	RemoveCalled  func(key string)
	LenCalled     func() int
}

func (ocs *ObjectsContainerStub) Get(key string) (interface{}, error) {
	return ocs.GetCalled(key)
}

func (ocs *ObjectsContainerStub) Add(key string, val interface{}) error {
	return ocs.AddCalled(key, val)
}

func (ocs *ObjectsContainerStub) Replace(key string, val interface{}) error {
	return ocs.ReplaceCalled(key, val)
}

func (ocs *ObjectsContainerStub) Remove(key string) {
	ocs.RemoveCalled(key)
}

func (ocs *ObjectsContainerStub) Len() int {
	return ocs.LenCalled()
}
