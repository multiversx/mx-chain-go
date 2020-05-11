package mock

// GenesisNodeInfoHandlerMock -
type GenesisNodeInfoHandlerMock struct {
	AssignedShardValue uint32
	AddressBytesValue  []byte
	PubKeyBytesValue   []byte
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
