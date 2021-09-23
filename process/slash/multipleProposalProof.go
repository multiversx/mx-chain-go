package slash

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

type multipleProposalProof struct {
	level        string
	slashingType SlashingType
	headers      []*interceptedBlocks.InterceptedHeader
}

// NewMultipleProposalProof - creates a new double block proposal slashing proof with a level, type and data
func NewMultipleProposalProof(level string, sType SlashingType, data []process.InterceptedData) (MultipleProposalProofHandler, error) {
	headers, err := convertInterceptedDataToHeader(data)
	if err != nil {
		return nil, err
	}

	return &multipleProposalProof{
		level:        level,
		slashingType: sType,
		headers:      headers,
	}, nil
}

// GetLevel - gets the slashing proofs level
func (mpp *multipleProposalProof) GetLevel() string {
	return mpp.level
}

// GetType - gets the slashing proofs type
func (mpp *multipleProposalProof) GetType() SlashingType {
	return mpp.slashingType
}

// GetHeaders - gets the slashing proofs level
func (mpp *multipleProposalProof) GetHeaders() []*interceptedBlocks.InterceptedHeader {
	return mpp.headers
}

func convertInterceptedDataToHeader(data []process.InterceptedData) ([]*interceptedBlocks.InterceptedHeader, error) {
	headers := make([]*interceptedBlocks.InterceptedHeader, 0, len(data))

	for _, d := range data {
		header, castOk := d.(*interceptedBlocks.InterceptedHeader)
		if !castOk {
			return nil, process.ErrCannotCastInterceptedDataToHeader
		}
		headers = append(headers, header)
	}

	return headers, nil
}
