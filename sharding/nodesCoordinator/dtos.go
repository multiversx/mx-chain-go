package nodesCoordinator

// ArgsUpdateNodes holds the parameters required by the shuffler to generate a new nodes configuration
type ArgsUpdateNodes struct {
	Eligible          map[uint32][]Validator
	Waiting           map[uint32][]Validator
	NewNodes          []Validator
	UnStakeLeaving    []Validator
	AdditionalLeaving []Validator
	Auction           []Validator
	Rand              []byte
	NbShards          uint32
	Epoch             uint32
}

// ResUpdateNodes holds the result of the UpdateNodes method
type ResUpdateNodes struct {
	Eligible       map[uint32][]Validator
	Waiting        map[uint32][]Validator
	ShuffledOut    map[uint32][]Validator
	Leaving        []Validator
	StillRemaining []Validator
	LowWaitingList bool
}
