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
func (nsm *NodeShufflerMock) UpdateNodeLists(
	args sharding.ArgsUpdateNodes,
) (map[uint32][]sharding.Validator, map[uint32][]sharding.Validator, []sharding.Validator) {
	return args.Eligible, args.Waiting, args.Leaving
}

// IsInterfaceNil -
func (nsm *NodeShufflerMock) IsInterfaceNil() bool {
	return nsm == nil
}
