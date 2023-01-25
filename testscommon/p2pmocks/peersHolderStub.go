package p2pmocks

import "github.com/multiversx/mx-chain-core-go/core"

// PeersHolderStub -
type PeersHolderStub struct {
	PutConnectionAddressCalled func(peerID core.PeerID, address string)
	PutShardIDCalled           func(peerID core.PeerID, shardID uint32)
	GetCalled                  func() map[uint32][]core.PeerID
	ContainsCalled             func(peerID core.PeerID) bool
	RemoveCalled               func(peerID core.PeerID)
	ClearCalled                func()
}

// PutConnectionAddress -
func (p *PeersHolderStub) PutConnectionAddress(peerID core.PeerID, address string) {
	if p.PutConnectionAddressCalled != nil {
		p.PutConnectionAddressCalled(peerID, address)
	}
}

// PutShardID -
func (p *PeersHolderStub) PutShardID(peerID core.PeerID, shardID uint32) {
	if p.PutShardIDCalled != nil {
		p.PutShardIDCalled(peerID, shardID)
	}
}

// Get -
func (p *PeersHolderStub) Get() map[uint32][]core.PeerID {
	if p.GetCalled != nil {
		return p.GetCalled()
	}

	return map[uint32][]core.PeerID{0: {"peer"}}
}

// Contains -
func (p *PeersHolderStub) Contains(peerID core.PeerID) bool {
	if p.ContainsCalled != nil {
		return p.ContainsCalled(peerID)
	}

	return false
}

// Remove -
func (p *PeersHolderStub) Remove(peerID core.PeerID) {
	if p.RemoveCalled != nil {
		p.RemoveCalled(peerID)
	}
}

// Clear -
func (p *PeersHolderStub) Clear() {
	p.ClearCalled()
}

// IsInterfaceNil -
func (p *PeersHolderStub) IsInterfaceNil() bool {
	return p == nil
}
