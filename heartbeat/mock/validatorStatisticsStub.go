package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

// ValidatorStatisticsStub -
type ValidatorStatisticsStub struct {
	RootHashCalled                    func() ([]byte, error)
	GetValidatorInfoForRootHashCalled func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error)
}

// RootHash -
func (vss *ValidatorStatisticsStub) RootHash() ([]byte, error) {
	if vss.RootHashCalled != nil {
		return vss.RootHashCalled()
	}

	return make([]byte, 0), nil
}

// GetValidatorInfoForRootHash -
func (vss *ValidatorStatisticsStub) GetValidatorInfoForRootHash(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
	if vss.GetValidatorInfoForRootHashCalled != nil {
		return vss.GetValidatorInfoForRootHashCalled(rootHash)
	}

	return make(map[uint32][]*state.ValidatorInfo), nil
}

// IsInterfaceNil -
func (vss *ValidatorStatisticsStub) IsInterfaceNil() bool {
	return vss == nil
}
