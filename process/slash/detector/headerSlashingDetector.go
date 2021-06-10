package detector

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// HeaderSlashingDetector - checks for slashable events for headers
type HeaderSlashingDetector struct {
}

// NewHeaderSlashingDetector - creates a new header slashing detector for multiple propose
func NewHeaderSlashingDetector() slash.SlashingDetector {
	return &HeaderSlashingDetector{}
}

// VerifyData - checks if an intercepted data represents a slashable event
func (hsd *HeaderSlashingDetector) VerifyData(data process.InterceptedData) slash.SlashingDetectorResultHandler {
	currentHeader := data.(*interceptedBlocks.InterceptedHeader)
	// check another header with the same round and proposer exists, but a different hash
	// if yes a slashingDetectorResult is returned with a message and the two headers
	return slash.NewSlashingDetectorResult("message", currentHeader, nil)
}

// GenerateProof - creates the SlashingProofHandler for the DetectorResult to be added to the Tx Data Field
func (hsd *HeaderSlashingDetector) GenerateProof(result slash.SlashingDetectorResultHandler) slash.SlashingProofHandler {
	return slash.NewSlashingProof("level", result.GetType(), result.GetData1(), result.GetData2())
}
