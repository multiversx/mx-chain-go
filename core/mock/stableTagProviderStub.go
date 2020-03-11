package mock

// StableTagProviderStub --
type StableTagProviderStub struct {
	FetchTagVersionCalled func() (string, error)
}

// FetchTagVersion --
func (s *StableTagProviderStub) FetchTagVersion() (string, error) {
	if s.FetchTagVersionCalled != nil {
		return s.FetchTagVersionCalled()
	}

	return "dummy", nil
}

// IsInterfaceNil --
func (s *StableTagProviderStub) IsInterfaceNil() bool {
	return s == nil
}
