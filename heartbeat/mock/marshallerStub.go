package mock

// MarshallerStub -
type MarshallerStub struct {
	MarshalHandler   func(obj interface{}) ([]byte, error)
	UnmarshalHandler func(obj interface{}, buff []byte) error
}

// Marshal -
func (ms MarshallerStub) Marshal(obj interface{}) ([]byte, error) {
	if ms.MarshalHandler != nil {
		return ms.MarshalHandler(obj)
	}
	return nil, nil
}

// Unmarshal -
func (ms MarshallerStub) Unmarshal(obj interface{}, buff []byte) error {
	if ms.UnmarshalHandler != nil {
		return ms.UnmarshalHandler(obj, buff)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MarshallerStub) IsInterfaceNil() bool {
	return ms == nil
}
