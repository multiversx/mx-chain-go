package mock

type StableTagProviderStub struct {
	FetchTagVersionCalled func() (string, error)
}

func (s *StableTagProviderStub) FetchTagVersion() (string, error) {
	if s.FetchTagVersionCalled != nil {
		return s.FetchTagVersionCalled()
	}

	return "dummy", nil
}

func (s *StableTagProviderStub) IsInterfaceNil() bool {
	return s == nil
}
