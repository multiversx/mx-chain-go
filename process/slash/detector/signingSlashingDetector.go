package detector

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// SigningSlashingDetector - checks for slashable events for headers
type SigningSlashingDetector struct {
}

// NewSigningSlashingDetector - creates a new header slashing detector for multiple signatures
func NewSigningSlashingDetector() slash.SlashingDetector {
	return &SigningSlashingDetector{}
}

// VerifyData - checks if an intercepted data represents a slashable event
func (hsd *SigningSlashingDetector) VerifyData(data process.InterceptedData) (slash.SlashingDetectorResultHandler, error) {
	// check another signature with the same round and proposer exists, but a different header exists
	// if yes a slashingDetectorResult is returned with a message and the two signatures
	return slash.NewSlashingDetectorResult("message", data, nil), nil
}

// GenerateProof - creates the SlashingProofHandler for the DetectorResult to be added to the Tx Data Field
func (hsd *SigningSlashingDetector) GenerateProof(result slash.SlashingDetectorResultHandler) slash.SlashingProofHandler {
	return slash.NewSlashingProof("level", result.GetType(), result.GetData1(), result.GetData2())
}
