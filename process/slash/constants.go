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

var ProofIDs = map[SlashingType]byte{
	None:             0x0,
	MultipleProposal: 0x1,
	MultipleSigning:  0x2,
}
