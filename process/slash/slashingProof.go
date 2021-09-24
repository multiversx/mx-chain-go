package slash

type slashingProof struct {
	level        SlashingLevel
	slashingType SlashingType
}

// NewSlashingProof - creates a new double block proposal slashing proof with a level, type and data
func NewSlashingProof(sType SlashingType, level SlashingLevel) SlashingProofHandler {
	return &slashingProof{
		level:        level,
		slashingType: sType,
	}
}

// GetLevel - gets the slashing proofs level
func (sp *slashingProof) GetLevel() SlashingLevel {
	return sp.level
}

// GetType - gets the slashing proofs type
func (sp *slashingProof) GetType() SlashingType {
	return sp.slashingType
}
