package slashMocks

import (
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go/process/slash/notifier"
)

// ProofTxDataExtractorStub -
type ProofTxDataExtractorStub struct {
	GetProofTxDataCalled func(proof coreSlash.SlashingProofHandler) (*notifier.ProofTxData, error)
}

// GetProofTxData -
func (ptds *ProofTxDataExtractorStub) GetProofTxData(proof coreSlash.SlashingProofHandler) (*notifier.ProofTxData, error) {
	if ptds.GetProofTxDataCalled != nil {
		return ptds.GetProofTxDataCalled(proof)
	}

	return &notifier.ProofTxData{
		Round:     0,
		ShardID:   0,
		Bytes:     []byte{0x1, 0x2},
		SlashType: proof.GetType(),
	}, nil
}

// IsInterfaceNil -
func (ptds *ProofTxDataExtractorStub) IsInterfaceNil() bool {
	return ptds == nil
}
