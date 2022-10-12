package mock

// HTTPClientStub -
type HTTPClientStub struct {
	PostCalled func(route string, payload interface{}) error
}

// Post -
func (stub *HTTPClientStub) Post(route string, payload interface{}) error {
	if stub.PostCalled != nil {
		return stub.PostCalled(route, payload)
	}

	return nil
}

// IsInterfaceNil -
func (stub *HTTPClientStub) IsInterfaceNil() bool {
	return stub == nil
}
