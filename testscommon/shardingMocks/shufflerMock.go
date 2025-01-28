package shardingMocks

import (
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// NodeShufflerMock -
type NodeShufflerMock struct {
}

// UpdateNodeLists -
func (nsm *NodeShufflerMock) UpdateNodeLists(args nodesCoordinator.ArgsUpdateNodes) (*nodesCoordinator.ResUpdateNodes, error) {
	return &nodesCoordinator.ResUpdateNodes{
		Eligible:       args.Eligible,
		Waiting:        args.Waiting,
		Leaving:        args.UnStakeLeaving,
		StillRemaining: make([]nodesCoordinator.Validator, 0),
	}, nil
}

// IsInterfaceNil -
func (nsm *NodeShufflerMock) IsInterfaceNil() bool {
	return nsm == nil
}
