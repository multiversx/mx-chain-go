package mock

type MarshalizerMock struct {
	MarshalHandler   func(obj interface{}) ([]byte, error)
	UnmarshalHandler func(obj interface{}, buff []byte) error
}

func (j MarshalizerMock) Marshal(obj interface{}) ([]byte, error) {
	if j.MarshalHandler != nil {
		return j.MarshalHandler(obj)
	}
	return nil, nil
}
func (j MarshalizerMock) Unmarshal(obj interface{}, buff []byte) error {
	if j.UnmarshalHandler != nil {
		return j.UnmarshalHandler(obj, buff)
	}
	return nil
}
