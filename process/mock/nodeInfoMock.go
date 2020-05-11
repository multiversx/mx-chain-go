package mock

// NodeInfoMock -
type NodeInfoMock struct {
	address []byte
	pubKey  []byte
	shardId uint32
}

// NewNodeInfo -
func NewNodeInfo(address []byte, pubKey []byte, shardId uint32) *NodeInfoMock {
	return &NodeInfoMock{
		address: address,
		pubKey:  pubKey,
		shardId: shardId,
	}
}

// AssignedShard -
func (n *NodeInfoMock) AssignedShard() uint32 {
	return n.shardId
}

// AddressBytes -
func (n *NodeInfoMock) AddressBytes() []byte {
	return n.address
}

// PubKeyBytes -
func (n *NodeInfoMock) PubKeyBytes() []byte {
	return n.pubKey
}

// IsInterfaceNil -
func (n *NodeInfoMock) IsInterfaceNil() bool {
	return n == nil
}
