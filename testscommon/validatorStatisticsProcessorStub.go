package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state"
)

// ValidatorStatisticsProcessorStub -
type ValidatorStatisticsProcessorStub struct {
	UpdatePeerStateCalled                    func(header data.MetaHeaderHandler) ([]byte, error)
	RevertPeerStateCalled                    func(header data.MetaHeaderHandler) error
	GetPeerAccountCalled                     func(address []byte) (state.PeerAccountHandler, error)
	RootHashCalled                           func() ([]byte, error)
	LastFinalizedRootHashCalled              func() []byte
	ResetValidatorStatisticsAtNewEpochCalled func(vInfos state.ShardValidatorsInfoMapHandler) error
	GetValidatorInfoForRootHashCalled        func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error)
	ProcessRatingsEndOfEpochCalled           func(validatorInfos state.ShardValidatorsInfoMapHandler, epoch uint32) error
	ProcessCalled                            func(validatorInfo data.ShardValidatorInfoHandler) error
	CommitCalled                             func() ([]byte, error)
	PeerAccountToValidatorInfoCalled         func(peerAccount state.PeerAccountHandler) *state.ValidatorInfo
	SaveNodesCoordinatorUpdatesCalled        func(epoch uint32) (bool, error)
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
func (vsp *ValidatorStatisticsProcessorStub) ResetValidatorStatisticsAtNewEpoch(vInfos state.ShardValidatorsInfoMapHandler) error {
	if vsp.ResetValidatorStatisticsAtNewEpochCalled != nil {
		return vsp.ResetValidatorStatisticsAtNewEpochCalled(vInfos)
	}
	return nil
}

// GetValidatorInfoForRootHash -
func (vsp *ValidatorStatisticsProcessorStub) GetValidatorInfoForRootHash(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
	if vsp.GetValidatorInfoForRootHashCalled != nil {
		return vsp.GetValidatorInfoForRootHashCalled(rootHash)
	}
	return state.NewShardValidatorsInfoMap(), nil
}

// UpdatePeerState -
func (vsp *ValidatorStatisticsProcessorStub) UpdatePeerState(header data.MetaHeaderHandler, _ map[string]data.HeaderHandler) ([]byte, error) {
	if vsp.UpdatePeerStateCalled != nil {
		return vsp.UpdatePeerStateCalled(header)
	}
	return nil, nil
}

// ProcessRatingsEndOfEpoch -
func (vsp *ValidatorStatisticsProcessorStub) ProcessRatingsEndOfEpoch(validatorInfos state.ShardValidatorsInfoMapHandler, epoch uint32) error {
	if vsp.ProcessRatingsEndOfEpochCalled != nil {
		return vsp.ProcessRatingsEndOfEpochCalled(validatorInfos, epoch)
	}
	return nil
}

// RevertPeerState -
func (vsp *ValidatorStatisticsProcessorStub) RevertPeerState(header data.MetaHeaderHandler) error {
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

// GetPeerAccount -
func (vsp *ValidatorStatisticsProcessorStub) GetPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	if vsp.GetPeerAccountCalled != nil {
		return vsp.GetPeerAccountCalled(address)
	}

	return nil, nil
}

// DisplayRatings -
func (vsp *ValidatorStatisticsProcessorStub) DisplayRatings(_ uint32) {
}

// SaveNodesCoordinatorUpdates -
func (vsp *ValidatorStatisticsProcessorStub) SaveNodesCoordinatorUpdates(epoch uint32) (bool, error) {
	if vsp.SaveNodesCoordinatorUpdatesCalled != nil {
		return vsp.SaveNodesCoordinatorUpdatesCalled(epoch)
	}
	return false, nil
}

// IsInterfaceNil -
func (vsp *ValidatorStatisticsProcessorStub) IsInterfaceNil() bool {
	return vsp == nil
}
