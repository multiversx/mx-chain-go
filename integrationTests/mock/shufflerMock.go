package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// NodeShufflerMock -
type NodeShufflerMock struct {
}

// UpdateParams -
func (nsm *NodeShufflerMock) UpdateParams(
	_ uint32,
	_ uint32,
	_ float32,
	_ bool,
) {

}

// UpdateNodeLists -
func (nsm *NodeShufflerMock) UpdateNodeLists(args sharding.ArgsUpdateNodes) (*sharding.ResUpdateNodes, error) {
	return &sharding.ResUpdateNodes{
		Eligible:       args.Eligible,
		Waiting:        args.Waiting,
		Leaving:        args.UnStakeLeaving,
		StillRemaining: make([]sharding.Validator, 0),
	}, nil
}

// IsInterfaceNil -
func (nsm *NodeShufflerMock) IsInterfaceNil() bool {
	return nsm == nil
}
