package mock

// MarshalizerMock -
type MarshalizerMock struct {
	MarshalHandler   func(obj interface{}) ([]byte, error)
	UnmarshalHandler func(obj interface{}, buff []byte) error
}

// Marshal -
func (j MarshalizerMock) Marshal(obj interface{}) ([]byte, error) {
	if j.MarshalHandler != nil {
		return j.MarshalHandler(obj)
	}
	return nil, nil
}

// Unmarshal -
func (j MarshalizerMock) Unmarshal(obj interface{}, buff []byte) error {
	if j.UnmarshalHandler != nil {
		return j.UnmarshalHandler(obj, buff)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (j *MarshalizerMock) IsInterfaceNil() bool {
	return j == nil
}
