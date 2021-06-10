package slash

import "github.com/ElrondNetwork/elrond-go/process"

type slashingProof struct {
	level        string
	slashingType string
	data1        process.InterceptedData
	data2        process.InterceptedData
}

// NewSlashingProof - creates a new slashing proof with a level, type and data
func NewSlashingProof(level string, sType string, data1 process.InterceptedData, data2 process.InterceptedData) SlashingProofHandler {
	return &slashingProof{
		level:        level,
		slashingType: sType,
		data1:        data1,
		data2:        data2,
	}
}

// GetLevel - gets the slashing proofs level
func (sp *slashingProof) GetLevel() string {
	return sp.level
}

// GetType - gets the slashing proofs type
func (sp *slashingProof) GetType() string {
	return sp.slashingType
}

// GetData1 - gets the slashing proofs first duplicate data
func (sp *slashingProof) GetData1() process.InterceptedData {
	return sp.data1
}

// GetData2 - gets the slashing proofs second duplicate data
func (sp *slashingProof) GetData2() process.InterceptedData {
	return sp.data2
}
