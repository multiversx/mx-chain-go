package slash

type SlashingType string

const (
	None             SlashingType = "no slashing"
	MultipleProposal SlashingType = "multiple header proposal"
	MultipleSigning  SlashingType = "multiple header signing"
)

type ThreatLevel uint8

const (
	Low    ThreatLevel = 0
	Medium ThreatLevel = 1
	High   ThreatLevel = 2
)

const (
	MultipleProposalProofID byte = 0x1
	MultipleSigningProofID  byte = 0x2
)

var ProofIDs = map[SlashingType]byte{
	MultipleProposal: MultipleProposalProofID,
	MultipleSigning:  MultipleSigningProofID,
}
