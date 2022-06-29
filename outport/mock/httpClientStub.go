package mock

// HTTPClientStub -
type HTTPClientStub struct {
	PostCalled func(route string, payload interface{}, response interface{}) error
}

// Post -
func (stub *HTTPClientStub) Post(route string, payload interface{}, response interface{}) error {
	if stub.PostCalled != nil {
		return stub.PostCalled(route, payload, response)
	}

	return nil
}

// IsInterfaceNil -
func (stub *HTTPClientStub) IsInterfaceNil() bool {
	return stub == nil
}
