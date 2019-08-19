package mock

type InterceptedDataStub struct {
	CheckValidCalled               func() error
	IsAddressedToOtherShardsCalled func() bool
}

func (ids InterceptedDataStub) CheckValid() error {
	return ids.CheckValidCalled()
}

func (ids InterceptedDataStub) IsAddressedToOtherShards() bool {
	return ids.IsAddressedToOtherShardsCalled()
}
