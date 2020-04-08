package mock

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

// Address -
func (n *NodeInfoMock) Address() []byte {
	return n.address
}

// PubKey -
func (n *NodeInfoMock) PubKey() []byte {
	return n.pubKey
}

// IsInterfaceNil -
func (n *NodeInfoMock) IsInterfaceNil() bool {
	return n == nil
}
