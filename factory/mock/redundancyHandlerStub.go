package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// RedundancyHandlerStub -
type RedundancyHandlerStub struct {
	IsRedundancyNodeCalled         func() bool
	IsMainMachineActiveCalled      func() bool
	ObserverPrivateKeyCalled       func() crypto.PrivateKey
	AdjustInactivityIfNeededCalled func(selfPubKey string, consensusPubKeys []string, roundIndex int64)
	ResetInactivityIfNeededCalled  func(selfPubKey string, consensusMsgPubKey string, consensusMsgPeerID core.PeerID)
}

// IsRedundancyNode -
func (rhs *RedundancyHandlerStub) IsRedundancyNode() bool {
	if rhs.IsRedundancyNodeCalled != nil {
		return rhs.IsRedundancyNodeCalled()
	}

	return false
}

// IsMainMachineActive -
func (rhs *RedundancyHandlerStub) IsMainMachineActive() bool {
	if rhs.IsMainMachineActiveCalled != nil {
		return rhs.IsMainMachineActiveCalled()
	}

	return true
}

// ObserverPrivateKey -
func (rhs *RedundancyHandlerStub) ObserverPrivateKey() crypto.PrivateKey {
	if rhs.ObserverPrivateKeyCalled != nil {
		return rhs.ObserverPrivateKeyCalled()
	}

	return &PrivateKeyStub{}
}

// AdjustInactivityIfNeeded -
func (rhs *RedundancyHandlerStub) AdjustInactivityIfNeeded(selfPubKey string, consensusPubKeys []string, roundIndex int64) {
	if rhs.AdjustInactivityIfNeededCalled != nil {
		rhs.AdjustInactivityIfNeededCalled(selfPubKey, consensusPubKeys, roundIndex)
	}
}

// ResetInactivityIfNeeded -
func (rhs *RedundancyHandlerStub) ResetInactivityIfNeeded(selfPubKey string, consensusMsgPubKey string, consensusMsgPeerID core.PeerID) {
	if rhs.ResetInactivityIfNeededCalled != nil {
		rhs.ResetInactivityIfNeededCalled(selfPubKey, consensusMsgPubKey, consensusMsgPeerID)
	}
}

// IsInterfaceNil -
func (rhs *RedundancyHandlerStub) IsInterfaceNil() bool {
	return rhs == nil
}
