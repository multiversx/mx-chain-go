package nodesCoordinator

import "github.com/ElrondNetwork/elrond-go-core/core"

// ArgsUpdateNodes holds the parameters required by the shuffler to generate a new nodes configuration
type ArgsUpdateNodes struct {
	Eligible          map[uint32][]core.Validator
	Waiting           map[uint32][]core.Validator
	NewNodes          []core.Validator
	UnStakeLeaving    []core.Validator
	AdditionalLeaving []core.Validator
	Rand              []byte
	NbShards          uint32
	Epoch             uint32
}

// ResUpdateNodes holds the result of the UpdateNodes method
type ResUpdateNodes struct {
	Eligible       map[uint32][]core.Validator
	Waiting        map[uint32][]core.Validator
	Leaving        []core.Validator
	StillRemaining []core.Validator
}
