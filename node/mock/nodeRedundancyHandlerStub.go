package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// NodeRedundancyHandlerStub -
type NodeRedundancyHandlerStub struct {
	IsRedundancyNodeCalled         func() bool
	IsMasterMachineActiveCalled    func() bool
	AdjustInactivityIfNeededCalled func(selfPubKey string, consensusPubKeys []string, roundIndex int64)
	ResetInactivityIfNeededCalled  func(selfPubKey string, consensusMsgPubKey string, consensusMsgPeerID core.PeerID)
}

// IsRedundancyNode -
func (nrhs *NodeRedundancyHandlerStub) IsRedundancyNode() bool {
	if nrhs.IsRedundancyNodeCalled != nil {
		return nrhs.IsRedundancyNodeCalled()
	}
	return false
}

// IsMasterMachineActive -
func (nrhs *NodeRedundancyHandlerStub) IsMasterMachineActive() bool {
	if nrhs.IsMasterMachineActiveCalled != nil {
		return nrhs.IsMasterMachineActiveCalled()
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

// IsInterfaceNil -
func (nrhs *NodeRedundancyHandlerStub) IsInterfaceNil() bool {
	return nrhs == nil
}
