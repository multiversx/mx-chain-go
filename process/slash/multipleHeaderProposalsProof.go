package slash

import "github.com/ElrondNetwork/elrond-go/process"

type multipleProposalProof struct {
	slashableHeaders *SlashingResult
}

// NewMultipleProposalProof - creates a new multiple block proposal slashing proof with a level, type and headers
func NewMultipleProposalProof(slashResult *SlashingResult) (MultipleProposalProofHandler, error) {
	if slashResult == nil {
		return nil, process.ErrNilSlashResult
	}
	if slashResult.Headers == nil {
		return nil, process.ErrNilHeaderHandler
	}

	return &multipleProposalProof{
		slashableHeaders: slashResult,
	}, nil
}

// GetLevel - returns the slashing proof threat level
func (mpp *multipleProposalProof) GetLevel() ThreatLevel {
	return mpp.slashableHeaders.SlashingLevel
}

// GetType - returns MultipleProposal
func (mpp *multipleProposalProof) GetType() SlashingType {
	return MultipleProposal
}

// GetHeaders - returns the slashing proofs headers
func (mpp *multipleProposalProof) GetHeaders() HeaderInfoList {
	return mpp.slashableHeaders.Headers
}
