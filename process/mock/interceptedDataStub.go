package mock

type InterceptedDataStub struct {
	CheckValidCalled              func() error
	IsAddressedToOtherShardCalled func() bool
}

func (ids InterceptedDataStub) CheckValid() error {
	return ids.CheckValidCalled()
}

func (ids InterceptedDataStub) IsAddressedToOtherShard() bool {
	return ids.IsAddressedToOtherShardCalled()
}
