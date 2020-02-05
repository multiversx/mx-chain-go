package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type ValidatorStatisticsProcessorMock struct {
	UpdatePeerStateCalled           func(header data.HeaderHandler) ([]byte, error)
	RevertPeerStateCalled           func(header data.HeaderHandler) error
	IsInterfaceNilCalled            func() bool
	RevertPeerStateToSnapshotCalled func(snapshot int) error
	GetPeerAccountCalled            func(address []byte) (state.PeerAccountHandler, error)
	GetTrieFromRootHashCalled       func(rootHash []byte) ([]state.ValidatorInfo, error)
	CommitCalled                    func() ([]byte, error)
	RootHashCalled                  func() ([]byte, error)
}

func (vsp *ValidatorStatisticsProcessorMock) UpdatePeerState(header data.HeaderHandler) ([]byte, error) {
	if vsp.UpdatePeerStateCalled != nil {
		return vsp.UpdatePeerStateCalled(header)
	}
	return nil, nil
}

func (vsp *ValidatorStatisticsProcessorMock) RevertPeerState(header data.HeaderHandler) error {
	if vsp.RevertPeerStateCalled != nil {
		return vsp.RevertPeerStateCalled(header)
	}
	return nil
}

func (vsp *ValidatorStatisticsProcessorMock) RevertPeerStateToSnapshot(snapshot int) error {
	if vsp.RevertPeerStateToSnapshotCalled != nil {
		return vsp.RevertPeerStateToSnapshotCalled(snapshot)
	}
	return nil
}

func (vsp *ValidatorStatisticsProcessorMock) Commit() ([]byte, error) {
	if vsp.CommitCalled != nil {
		return vsp.CommitCalled()
	}
	return nil, nil
}

func (vsp *ValidatorStatisticsProcessorMock) RootHash() ([]byte, error) {
	if vsp.RootHashCalled != nil {
		return vsp.RootHashCalled()
	}
	return nil, nil
}

func (vsp *ValidatorStatisticsProcessorMock) GetPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	if vsp.GetPeerAccountCalled != nil {
		return vsp.GetPeerAccountCalled(address)
	}

	return nil, nil
}

func (vsp *ValidatorStatisticsProcessorMock) GetValidatorInfosForHash(rootHash []byte) ([]state.ValidatorInfo, error) {
	if vsp.GetTrieFromRootHashCalled != nil {
		return vsp.GetTrieFromRootHashCalled(rootHash)
	}

	return nil, nil
}

func (vsp *ValidatorStatisticsProcessorMock) IsInterfaceNil() bool {
	if vsp.IsInterfaceNilCalled != nil {
		return vsp.IsInterfaceNilCalled()
	}
	return false
}
