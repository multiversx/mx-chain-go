package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// ValidatorStatisticsProcessorMock -
type ValidatorStatisticsProcessorMock struct {
	UpdatePeerStateCalled func(header data.MetaHeaderHandler) ([]byte, error)
	RevertPeerStateCalled func(header data.MetaHeaderHandler) error
	IsInterfaceNilCalled  func() bool

	GetPeerAccountCalled                     func(address []byte) (state.PeerAccountHandler, error)
	RootHashCalled                           func() ([]byte, error)
	ResetValidatorStatisticsAtNewEpochCalled func(vInfos map[uint32][]*state.ValidatorInfo) error
	GetValidatorInfoForRootHashCalled        func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error)
	ProcessCalled                            func(validatorInfo data.ShardValidatorInfoHandler) error
	CommitCalled                             func() ([]byte, error)
	ProcessRatingsEndOfEpochCalled           func(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error
	PeerAccountToValidatorInfoCalled         func(peerAccount state.PeerAccountHandler) *state.ValidatorInfo
	SaveNodesCoordinatorUpdatesCalled        func(epoch uint32) (bool, error)
}

// SaveNodesCoordinatorUpdates -
func (vsp *ValidatorStatisticsProcessorMock) SaveNodesCoordinatorUpdates(epoch uint32) (bool, error) {
	if vsp.SaveNodesCoordinatorUpdatesCalled != nil {
		return vsp.SaveNodesCoordinatorUpdatesCalled(epoch)
	}
	return false, nil
}

// PeerAccountToValidatorInfo -
func (vsp *ValidatorStatisticsProcessorMock) PeerAccountToValidatorInfo(peerAccount state.PeerAccountHandler) *state.ValidatorInfo {
	if vsp.PeerAccountToValidatorInfoCalled != nil {
		return vsp.PeerAccountToValidatorInfoCalled(peerAccount)
	}
	return nil
}

// UpdatePeerState -
func (vsp *ValidatorStatisticsProcessorMock) UpdatePeerState(header data.MetaHeaderHandler, _ map[string]data.HeaderHandler) ([]byte, error) {
	if vsp.UpdatePeerStateCalled != nil {
		return vsp.UpdatePeerStateCalled(header)
	}
	return nil, nil
}

// Process -
func (vsp *ValidatorStatisticsProcessorMock) Process(validatorInfo data.ShardValidatorInfoHandler) error {
	if vsp.ProcessCalled != nil {
		return vsp.ProcessCalled(validatorInfo)
	}

	return nil
}

// Commit -
func (vsp *ValidatorStatisticsProcessorMock) Commit() ([]byte, error) {
	if vsp.CommitCalled != nil {
		return vsp.CommitCalled()
	}

	return nil, nil
}

// ProcessRatingsEndOfEpoch -
func (vsp *ValidatorStatisticsProcessorMock) ProcessRatingsEndOfEpoch(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error {
	if vsp.ProcessRatingsEndOfEpochCalled != nil {
		return vsp.ProcessRatingsEndOfEpochCalled(validatorInfos, epoch)
	}

	return nil
}

// ResetValidatorStatisticsAtNewEpoch -
func (vsp *ValidatorStatisticsProcessorMock) ResetValidatorStatisticsAtNewEpoch(vInfos map[uint32][]*state.ValidatorInfo) error {
	if vsp.ResetValidatorStatisticsAtNewEpochCalled != nil {
		return vsp.ResetValidatorStatisticsAtNewEpochCalled(vInfos)
	}
	return nil
}

// GetValidatorInfoForRootHash -
func (vsp *ValidatorStatisticsProcessorMock) GetValidatorInfoForRootHash(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
	if vsp.GetValidatorInfoForRootHashCalled != nil {
		return vsp.GetValidatorInfoForRootHashCalled(rootHash)
	}
	return nil, nil
}

// RevertPeerState -
func (vsp *ValidatorStatisticsProcessorMock) RevertPeerState(header data.MetaHeaderHandler) error {
	if vsp.RevertPeerStateCalled != nil {
		return vsp.RevertPeerStateCalled(header)
	}
	return nil
}

// RootHash -
func (vsp *ValidatorStatisticsProcessorMock) RootHash() ([]byte, error) {
	if vsp.RootHashCalled != nil {
		return vsp.RootHashCalled()
	}
	return nil, nil
}

// GetExistingPeerAccount -
func (vsp *ValidatorStatisticsProcessorMock) GetExistingPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	if vsp.GetPeerAccountCalled != nil {
		return vsp.GetPeerAccountCalled(address)
	}

	return nil, nil
}

// DisplayRatings -
func (vsp *ValidatorStatisticsProcessorMock) DisplayRatings(_ uint32) {
}

// SetLastFinalizedRootHash -
func (vsp *ValidatorStatisticsProcessorMock) SetLastFinalizedRootHash(_ []byte) {
}

// LastFinalizedRootHash -
func (vsp *ValidatorStatisticsProcessorMock) LastFinalizedRootHash() []byte {
	return nil
}

// IsInterfaceNil -
func (vsp *ValidatorStatisticsProcessorMock) IsInterfaceNil() bool {
	return false
}
