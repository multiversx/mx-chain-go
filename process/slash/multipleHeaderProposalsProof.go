package slash

type multipleProposalProof struct {
	slashableHeaders *SlashingResult
}

// NewMultipleProposalProof - creates a new double block proposal slashing proof with a level, type and data
func NewMultipleProposalProof(slashableData *SlashingResult) (MultipleProposalProofHandler, error) {
	return &multipleProposalProof{
		slashableHeaders: slashableData,
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
	return mpp.slashableHeaders.Data
}
