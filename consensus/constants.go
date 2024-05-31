package consensus

// TODO: inject consensus model factories and remove the switches in the code

// ConsensusModel defines the consensus models
type ConsensusModel string

const (
	// ConsensusModelInvalid defines an invalid consensus model
	ConsensusModelInvalid ConsensusModel = "consensus model invalid"
	// ConsensusModelV1 defines the first variant of the consensus model
	ConsensusModelV1 ConsensusModel = "consensus model v1"
	// ConsensusModelV2 defines the second variant of the consensus model
	ConsensusModelV2 ConsensusModel = "consensus model v2"
)
