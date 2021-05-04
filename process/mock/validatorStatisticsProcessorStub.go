package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// ValidatorStatisticsProcessorStub -
type ValidatorStatisticsProcessorStub struct {
	UpdatePeerStateCalled                    func(header data.HeaderHandler) ([]byte, error)
	RevertPeerStateCalled                    func(header data.HeaderHandler) error
	GetPeerAccountCalled                     func(address []byte) (state.PeerAccountHandler, error)
	RootHashCalled                           func() ([]byte, error)
	LastFinalizedRootHashCalled              func() []byte
	ResetValidatorStatisticsAtNewEpochCalled func(vInfos map[uint32][]*state.ValidatorInfo) error
	GetValidatorInfoForRootHashCalled        func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error)
	ProcessRatingsEndOfEpochCalled           func(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error
	ProcessCalled                            func(validatorInfo data.ShardValidatorInfoHandler) error
	CommitCalled                             func() ([]byte, error)
	PeerAccountToValidatorInfoCalled         func(peerAccount state.PeerAccountHandler) *state.ValidatorInfo
	SaveNodesCoordinatorUpdatesCalled        func(epoch uint32) (bool, error)
}

// SaveNodesCoordinatorUpdates -
func (vsp *ValidatorStatisticsProcessorStub) SaveNodesCoordinatorUpdates(epoch uint32) (bool, error) {
	if vsp.SaveNodesCoordinatorUpdatesCalled != nil {
		return vsp.SaveNodesCoordinatorUpdatesCalled(epoch)
	}
	return false, nil
}

// PeerAccountToValidatorInfo -
func (vsp *ValidatorStatisticsProcessorStub) PeerAccountToValidatorInfo(peerAccount state.PeerAccountHandler) *state.ValidatorInfo {
	if vsp.PeerAccountToValidatorInfoCalled != nil {
		return vsp.PeerAccountToValidatorInfoCalled(peerAccount)
	}
	return nil
}

// Process -
func (vsp *ValidatorStatisticsProcessorStub) Process(validatorInfo data.ShardValidatorInfoHandler) error {
	if vsp.ProcessCalled != nil {
		return vsp.ProcessCalled(validatorInfo)
	}

	return nil
}

// Commit -
func (vsp *ValidatorStatisticsProcessorStub) Commit() ([]byte, error) {
	if vsp.CommitCalled != nil {
		return vsp.CommitCalled()
	}

	return nil, nil
}

// ResetValidatorStatisticsAtNewEpoch -
func (vsp *ValidatorStatisticsProcessorStub) ResetValidatorStatisticsAtNewEpoch(vInfos map[uint32][]*state.ValidatorInfo) error {
	if vsp.ResetValidatorStatisticsAtNewEpochCalled != nil {
		return vsp.ResetValidatorStatisticsAtNewEpochCalled(vInfos)
	}
	return nil
}

// GetValidatorInfoForRootHash -
func (vsp *ValidatorStatisticsProcessorStub) GetValidatorInfoForRootHash(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
	if vsp.GetValidatorInfoForRootHashCalled != nil {
		return vsp.GetValidatorInfoForRootHashCalled(rootHash)
	}
	return nil, nil
}

// ProcessRatingsEndOfEpoch -
func (vsp *ValidatorStatisticsProcessorStub) ProcessRatingsEndOfEpoch(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error {
	if vsp.ProcessRatingsEndOfEpochCalled != nil {
		return vsp.ProcessRatingsEndOfEpochCalled(validatorInfos, epoch)
	}
	return nil
}

// UpdatePeerState -
func (vsp *ValidatorStatisticsProcessorStub) UpdatePeerState(header data.HeaderHandler, _ map[string]data.HeaderHandler) ([]byte, error) {
	if vsp.UpdatePeerStateCalled != nil {
		return vsp.UpdatePeerStateCalled(header)
	}
	return nil, nil
}

// RevertPeerState -
func (vsp *ValidatorStatisticsProcessorStub) RevertPeerState(header data.HeaderHandler) error {
	if vsp.RevertPeerStateCalled != nil {
		return vsp.RevertPeerStateCalled(header)
	}
	return nil
}

// RootHash -
func (vsp *ValidatorStatisticsProcessorStub) RootHash() ([]byte, error) {
	if vsp.RootHashCalled != nil {
		return vsp.RootHashCalled()
	}
	return nil, nil
}

// GetExistingPeerAccount -
func (vsp *ValidatorStatisticsProcessorStub) GetExistingPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	if vsp.GetPeerAccountCalled != nil {
		return vsp.GetPeerAccountCalled(address)
	}

	return nil, nil
}

// DisplayRatings -
func (vsp *ValidatorStatisticsProcessorStub) DisplayRatings(_ uint32) {
}

// SetLastFinalizedRootHash -
func (vsp *ValidatorStatisticsProcessorStub) SetLastFinalizedRootHash(_ []byte) {
}

// LastFinalizedRootHash -
func (vsp *ValidatorStatisticsProcessorStub) LastFinalizedRootHash() []byte {
	if vsp.LastFinalizedRootHashCalled != nil {
		return vsp.LastFinalizedRootHashCalled()
	}
	return nil
}

// IsInterfaceNil -
func (vsp *ValidatorStatisticsProcessorStub) IsInterfaceNil() bool {
	return false
}
