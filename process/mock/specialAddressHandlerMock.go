package mock

type SpecialAddressHandlerMock struct {
	ElrondCommunityAddressCalled func() []byte
	LeaderAddressCalled          func() []byte
	BurnAddressCalled            func() []byte
	ShardIdForAddressCalled      func([]byte) (uint32, error)

	addresses []string
}

func (sh *SpecialAddressHandlerMock) SetElrondCommunityAddress(elrond []byte) {
}

func (sh *SpecialAddressHandlerMock) SetConsensusRewardAddresses(consensusRewardAddresses []string) {
	sh.addresses = consensusRewardAddresses
}

func (sh *SpecialAddressHandlerMock) ConsensusRewardAddresses() []string {
	return sh.addresses
}

func (sh *SpecialAddressHandlerMock) BurnAddress() []byte {
	if sh.BurnAddressCalled == nil {
		return []byte("burn")
	}

	return sh.BurnAddressCalled()
}

func (sh *SpecialAddressHandlerMock) ElrondCommunityAddress() []byte {
	if sh.ElrondCommunityAddressCalled == nil {
		return []byte("elrond")
	}

	return sh.ElrondCommunityAddressCalled()
}

func (sh *SpecialAddressHandlerMock) LeaderAddress() []byte {
	if sh.LeaderAddressCalled == nil {
		return []byte("leader")
	}

	return sh.LeaderAddressCalled()
}

func (sh *SpecialAddressHandlerMock) ShardIdForAddress(addr []byte) (uint32, error) {
	if sh.ShardIdForAddressCalled == nil {
		return 0, nil
	}

	return sh.ShardIdForAddressCalled(addr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sh *SpecialAddressHandlerMock) IsInterfaceNil() bool {
	if sh == nil {
		return true
	}
	return false
}
