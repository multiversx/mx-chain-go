package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// ValidatorStatisticsProcessorStub -
type ValidatorStatisticsProcessorStub struct {
	UpdatePeerStateCalled                    func(header data.HeaderHandler) ([]byte, error)
	RevertPeerStateCalled                    func(header data.HeaderHandler) error
	IsInterfaceNilCalled                     func() bool
	GetPeerAccountCalled                     func(address []byte) (state.PeerAccountHandler, error)
	RootHashCalled                           func() ([]byte, error)
	ResetValidatorStatisticsAtNewEpochCalled func(vInfos map[uint32][]*state.ValidatorInfo) error
	GetValidatorInfoForRootHashCalled        func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error)
	ProcessCalled                            func(vid *state.ValidatorInfo) error
}

// Process -
func (vsp *ValidatorStatisticsProcessorStub) Process(vid *state.ValidatorInfo) error {
	if vsp.ProcessCalled != nil {
		return vsp.ProcessCalled(vid)
	}

	return nil
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

// UpdatePeerState -
func (vsp *ValidatorStatisticsProcessorStub) UpdatePeerState(header data.HeaderHandler) ([]byte, error) {
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

// GetPeerAccount -
func (vsp *ValidatorStatisticsProcessorStub) GetPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	if vsp.GetPeerAccountCalled != nil {
		return vsp.GetPeerAccountCalled(address)
	}

	return nil, nil
}

// IsInterfaceNil -
func (vsp *ValidatorStatisticsProcessorStub) IsInterfaceNil() bool {
	if vsp.IsInterfaceNilCalled != nil {
		return vsp.IsInterfaceNilCalled()
	}
	return false
}
