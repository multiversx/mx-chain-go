package slash

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

type multipleProposalProof struct {
	slashableHeaders slashingHeaders
}

// NewMultipleProposalProof - creates a new double block proposal slashing proof with a level, type and data
func NewMultipleProposalProof(slashableData *SlashingResult) (MultipleProposalProofHandler, error) {
	headers, err := convertInterceptedDataToInterceptedHeaders(slashableData.Data)
	if err != nil {
		return nil, err
	}

	return &multipleProposalProof{
		slashableHeaders: slashingHeaders{
			slashingLevel: slashableData.SlashingLevel,
			headers:       headers,
		},
	}, nil
}

// GetLevel - returns the slashing proof level
func (mpp *multipleProposalProof) GetLevel() ThreatLevel {
	return mpp.slashableHeaders.slashingLevel
}

// GetType - returns MultipleProposal
func (mpp *multipleProposalProof) GetType() SlashingType {
	return MultipleProposal
}

// GetHeaders - returns the slashing proofs headers
func (mpp *multipleProposalProof) GetHeaders() []*interceptedBlocks.InterceptedHeader {
	return mpp.slashableHeaders.headers
}
