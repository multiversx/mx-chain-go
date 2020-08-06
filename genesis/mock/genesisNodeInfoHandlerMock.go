package mock

// GenesisNodeInfoHandlerMock -
type GenesisNodeInfoHandlerMock struct {
	AssignedShardValue uint32
	AddressBytesValue  []byte
	PubKeyBytesValue   []byte
	InitialRatingValue uint32
}

// GetInitialRating -
func (gnihm *GenesisNodeInfoHandlerMock) GetInitialRating() uint32 {
	initialRating := gnihm.InitialRatingValue
	if initialRating == uint32(0) {
		initialRating = uint32(50)
	}

	return initialRating
}

// AssignedShard -
func (gnihm *GenesisNodeInfoHandlerMock) AssignedShard() uint32 {
	return gnihm.AssignedShardValue
}

// AddressBytes -
func (gnihm *GenesisNodeInfoHandlerMock) AddressBytes() []byte {
	return gnihm.AddressBytesValue
}

// PubKeyBytes -
func (gnihm *GenesisNodeInfoHandlerMock) PubKeyBytes() []byte {
	return gnihm.PubKeyBytesValue
}

// IsInterfaceNil -
func (gnihm *GenesisNodeInfoHandlerMock) IsInterfaceNil() bool {
	return gnihm == nil
}
