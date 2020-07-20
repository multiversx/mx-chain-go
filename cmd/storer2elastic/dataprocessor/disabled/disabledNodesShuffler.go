package disabled

import "github.com/ElrondNetwork/elrond-go/sharding"

type disabledNodesShuffler struct {
}

// NewNodesShuffler will return a new instance of disabledNodesShuffler
func NewNodesShuffler() *disabledNodesShuffler {
	return &disabledNodesShuffler{}
}

// UpdateParams won't do anything
func (d *disabledNodesShuffler) UpdateParams(numNodesShard uint32, numNodesMeta uint32, hysteresis float32, adaptivity bool) {
}

// UpdateNodeLists will return already existing data
func (d *disabledNodesShuffler) UpdateNodeLists(args sharding.ArgsUpdateNodes) (*sharding.ResUpdateNodes, error) {
	return &sharding.ResUpdateNodes{
		Eligible:       args.Eligible,
		Waiting:        args.Waiting,
		Leaving:        args.UnStakeLeaving,
		StillRemaining: args.NewNodes,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledNodesShuffler) IsInterfaceNil() bool {
	return d == nil
}
