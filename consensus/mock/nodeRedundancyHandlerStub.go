package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// NodeRedundancyHandlerStub -
type NodeRedundancyHandlerStub struct {
	IsRedundancyNodeCalled         func() bool
	IsMainMachineActiveCalled      func() bool
	AdjustInactivityIfNeededCalled func(selfPubKey string, consensusPubKeys []string, roundIndex int64)
	ResetInactivityIfNeededCalled  func(selfPubKey string, consensusMsgPubKey string, consensusMsgPeerID core.PeerID)
	ObserverPrivateKeyCalled       func() crypto.PrivateKey
}

// IsRedundancyNode -
func (nrhs *NodeRedundancyHandlerStub) IsRedundancyNode() bool {
	if nrhs.IsRedundancyNodeCalled != nil {
		return nrhs.IsRedundancyNodeCalled()
	}
	return false
}

// IsMainMachineActive -
func (nrhs *NodeRedundancyHandlerStub) IsMainMachineActive() bool {
	if nrhs.IsMainMachineActiveCalled != nil {
		return nrhs.IsMainMachineActiveCalled()
	}
	return true
}

// AdjustInactivityIfNeeded -
func (nrhs *NodeRedundancyHandlerStub) AdjustInactivityIfNeeded(selfPubKey string, consensusPubKeys []string, roundIndex int64) {
	if nrhs.AdjustInactivityIfNeededCalled != nil {
		nrhs.AdjustInactivityIfNeededCalled(selfPubKey, consensusPubKeys, roundIndex)
	}
}

// ResetInactivityIfNeeded -
func (nrhs *NodeRedundancyHandlerStub) ResetInactivityIfNeeded(selfPubKey string, consensusMsgPubKey string, consensusMsgPeerID core.PeerID) {
	if nrhs.ResetInactivityIfNeededCalled != nil {
		nrhs.ResetInactivityIfNeededCalled(selfPubKey, consensusMsgPubKey, consensusMsgPeerID)
	}
}

// ObserverPrivateKey -
func (nrhs *NodeRedundancyHandlerStub) ObserverPrivateKey() crypto.PrivateKey {
	if nrhs.ObserverPrivateKeyCalled != nil {
		return nrhs.ObserverPrivateKeyCalled()
	}

	return &PrivateKeyMock{}
}

// IsInterfaceNil -
func (nrhs *NodeRedundancyHandlerStub) IsInterfaceNil() bool {
	return nrhs == nil
}
