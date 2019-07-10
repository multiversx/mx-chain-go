package mock

type SpecialAddressHandlerMock struct {
	ElrondCommunityAddressCalled func() []byte
	OwnAddressCalled             func() []byte
	ShardIdForAddressCalled      func([]byte) uint32
}

func (sh *SpecialAddressHandlerMock) ElrondCommunityAddress() []byte {
	if sh.ElrondCommunityAddressCalled == nil {
		return []byte("elrond")
	}

	return sh.ElrondCommunityAddressCalled()
}

func (sh *SpecialAddressHandlerMock) LeaderAddress() []byte {
	if sh.OwnAddressCalled == nil {
		return []byte("leader")
	}

	return sh.OwnAddressCalled()
}

func (sh *SpecialAddressHandlerMock) ShardIdForAddress(addr []byte) uint32 {
	if sh.ShardIdForAddressCalled == nil {
		return 0
	}

	return sh.ShardIdForAddressCalled(addr)
}
