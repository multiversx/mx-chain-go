package slash

// SlashingType defines slashing type event
type SlashingType string

const (
	// None = no slashing event
	None SlashingType = "no slashing"
	// MultipleProposal = a proposer has proposed multiple headers
	MultipleProposal SlashingType = "multiple header proposal"
	// MultipleSigning = one or more validators have signed multiple headers
	MultipleSigning SlashingType = "multiple header signing"
)

// ThreatLevel defines the slashing threat level/severity
type ThreatLevel uint8

const (
	// Low = low threat level
	Low ThreatLevel = 0
	// Medium = medium threat level
	Medium ThreatLevel = 1
	// High = high threat level
	High ThreatLevel = 2
)

// Used by slashing notifier to create a slashing transaction
// from a proof. Each transaction identifies a different
// slashing event based on this ID
const (
	// MultipleProposalProofID = MultipleProposal's ID
	MultipleProposalProofID byte = 0x1
	// MultipleSigningProofID = MultipleSigning's ID
	MultipleSigningProofID byte = 0x2
)

// ProofIDs is a convenience map containing all slashing events
// with their corresponding ID
var ProofIDs = map[SlashingType]byte{
	MultipleProposal: MultipleProposalProofID,
	MultipleSigning:  MultipleSigningProofID,
}
