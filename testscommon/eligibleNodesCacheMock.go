package testscommon

import "github.com/multiversx/mx-chain-core-go/core"

// EligibleNodesCacheMock -
type EligibleNodesCacheMock struct {
	IsPeerEligibleCalled func(pid core.PeerID, shard uint32, epoch uint32) bool
}

// IsPeerEligible -
func (mock *EligibleNodesCacheMock) IsPeerEligible(pid core.PeerID, shard uint32, epoch uint32) bool {
	if mock.IsPeerEligibleCalled != nil {
		return mock.IsPeerEligibleCalled(pid, shard, epoch)
	}

	return true
}

// IsInterfaceNil -
func (mock *EligibleNodesCacheMock) IsInterfaceNil() bool {
	return mock == nil
}
