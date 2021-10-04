package slash

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

type multipleProposalProof struct {
	slashableHeaders headersWithSlashingLevel
}

// NewMultipleProposalProof - creates a new double block proposal slashing proof with a level, type and data
func NewMultipleProposalProof(slashableData DataWithSlashingLevel) (MultipleProposalProofHandler, error) {
	headers, err := convertInterceptedDataToHeader(slashableData.Data)
	if err != nil {
		return nil, err
	}

	return &multipleProposalProof{
		slashableHeaders: headersWithSlashingLevel{
			slashingLevel: slashableData.SlashingLevel,
			headers:       headers,
		},
	}, nil
}

// GetLevel - gets the slashing proofs level
func (mpp *multipleProposalProof) GetLevel() SlashingLevel {
	return mpp.slashableHeaders.slashingLevel
}

// GetType - gets the slashing proofs type
func (mpp *multipleProposalProof) GetType() SlashingType {
	return MultipleProposal
}

// GetHeaders - gets the slashing proofs headers
func (mpp *multipleProposalProof) GetHeaders() []*interceptedBlocks.InterceptedHeader {
	return mpp.slashableHeaders.headers
}
