package mock

type InterceptedDataStub struct {
	CheckValidityCalled func() error
	IsForMyShardCalled  func() bool
}

func (ids *InterceptedDataStub) CheckValidity() error {
	return ids.CheckValidityCalled()
}

func (ids *InterceptedDataStub) IsForMyShard() bool {
	return ids.IsForMyShardCalled()
}

func (ids *InterceptedDataStub) IsInterfaceNil() bool {
	if ids == nil {
		return true
	}
	return false
}
