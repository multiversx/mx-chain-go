package mock

import "github.com/ElrondNetwork/elrond-go-sandbox/consensus"

type ValidatorGroupSelectorStub struct {
	GetSelectedPublicKeysCalled  func(selection []byte) (publicKeys []string, err error)
	LoadEligibleListCalled       func(eligibleList []consensus.Validator) error
	ComputeValidatorsGroupCalled func(randomness []byte) (validatorsGroup []consensus.Validator, err error)
	ConsensusGroupSizeCalled     func() int
	SetConsensusGroupSizeCalled  func(int) error
}

func (vgss *ValidatorGroupSelectorStub) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	return vgss.GetSelectedPublicKeysCalled(selection)
}

func (vgss *ValidatorGroupSelectorStub) LoadEligibleList(eligibleList []consensus.Validator) error {
	return vgss.LoadEligibleListCalled(eligibleList)
}

func (vgss *ValidatorGroupSelectorStub) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []consensus.Validator, err error) {
	return vgss.ComputeValidatorsGroupCalled(randomness)
}

func (vgss *ValidatorGroupSelectorStub) ConsensusGroupSize() int {
	return vgss.ConsensusGroupSizeCalled()
}

func (vgss *ValidatorGroupSelectorStub) SetConsensusGroupSize(size int) error {
	return vgss.SetConsensusGroupSize(size)
}
