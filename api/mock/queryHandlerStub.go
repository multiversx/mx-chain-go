package mock

// QueryHandlerStub -
type QueryHandlerStub struct {
	EnabledCalled func() bool
	QueryCalled   func(search string) []string
}

// Enabled -
func (qhs *QueryHandlerStub) Enabled() bool {
	if qhs.EnabledCalled != nil {
		return qhs.EnabledCalled()
	}

	return false
}

// Query -
func (qhs *QueryHandlerStub) Query(search string) []string {
	if qhs.QueryCalled != nil {
		return qhs.QueryCalled(search)
	}

	return make([]string, 0)
}

// IsInterfaceNil -
func (qhs *QueryHandlerStub) IsInterfaceNil() bool {
	return qhs == nil
}
