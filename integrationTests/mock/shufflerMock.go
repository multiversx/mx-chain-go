package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

type NodeShufflerMock struct {
}

func (nsm *NodeShufflerMock) UpdateParams(
	numNodesShard uint32,
	numNodesMeta uint32,
	hysteresis float32,
	adaptivity bool,
) {

}

func (nsm *NodeShufflerMock) UpdateNodeLists(
	args sharding.ArgsUpdateNodes,
) (map[uint32][]sharding.Validator, map[uint32][]sharding.Validator, []sharding.Validator) {
	return args.Eligible, args.Waiting, args.Leaving
}

func (nsm *NodeShufflerMock) IsInterfaceNil() bool {
	return nsm == nil
}
