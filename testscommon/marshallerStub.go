package testscommon

// MarshallerStub -
type MarshallerStub struct {
	MarshalCalled   func(obj interface{}) ([]byte, error)
	UnmarshalCalled func(obj interface{}, buff []byte) error
}

// Marshal -
func (stub *MarshallerStub) Marshal(obj interface{}) ([]byte, error) {
	if stub.MarshalCalled != nil {
		return stub.MarshalCalled(obj)
	}

	return make([]byte, 0), nil
}

// Unmarshal -
func (stub *MarshallerStub) Unmarshal(obj interface{}, buff []byte) error {
	if stub.UnmarshalCalled != nil {
		return stub.UnmarshalCalled(obj, buff)
	}

	return nil
}

// IsInterfaceNil -
func (stub *MarshallerStub) IsInterfaceNil() bool {
	return stub == nil
}
