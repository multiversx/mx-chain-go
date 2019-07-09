package mock

type SpecialAddressHandlerMock struct {
	GetElrondCommunityAddressCalled func() []byte
	GetLeaderAddressCalled          func() []byte
	ShardIdForAddressCalled         func([]byte) uint32
}

func (sh *SpecialAddressHandlerMock) GetElrondCommunityAddress() []byte {
	if sh.GetElrondCommunityAddressCalled == nil {
		return []byte("elrond")
	}

	return sh.GetElrondCommunityAddressCalled()
}

func (sh *SpecialAddressHandlerMock) GetLeaderAddress() []byte {
	if sh.GetLeaderAddressCalled == nil {
		return []byte("leader")
	}

	return sh.GetLeaderAddressCalled()
}

func (sh *SpecialAddressHandlerMock) ShardIdForAddress(addr []byte) uint32 {
	if sh.ShardIdForAddressCalled == nil {
		return 0
	}

	return sh.ShardIdForAddressCalled(addr)
}
