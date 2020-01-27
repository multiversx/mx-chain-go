package mock

type InterceptedDataStub struct {
	CheckValidityCalled     func() error
	IsForCurrentShardCalled func() bool
	HashCalled              func() []byte
}

func (ids *InterceptedDataStub) Hash() []byte {
	if ids.HashCalled != nil {
		return ids.HashCalled()
	}

	return []byte("mock hash")
}

func (ids *InterceptedDataStub) CheckValidity() error {
	return ids.CheckValidityCalled()
}

func (ids *InterceptedDataStub) IsForCurrentShard() bool {
	return ids.IsForCurrentShardCalled()
}

func (ids *InterceptedDataStub) Type() string {
	return "intercepted data stub"
}

func (ids *InterceptedDataStub) IsInterfaceNil() bool {
	return ids == nil
}
