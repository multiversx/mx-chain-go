package slash

type slashingProof struct {
	level        string
	slashingType SlashingType
}

// NewSlashingProof - creates a new double block proposal slashing proof with a level, type and data
func NewSlashingProof(level string, sType SlashingType) SlashingProofHandler {
	return &slashingProof{
		level:        level,
		slashingType: sType,
	}
}

// GetLevel - gets the slashing proofs level
func (mpp *slashingProof) GetLevel() string {
	return mpp.level
}

// GetType - gets the slashing proofs type
func (mpp *slashingProof) GetType() SlashingType {
	return mpp.slashingType
}
