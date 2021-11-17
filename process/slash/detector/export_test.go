package detector

import "github.com/ElrondNetwork/elrond-go-core/data"

func (mhp *multipleHeaderProposalsDetector) CheckProposedHeaders(headers []data.HeaderHandler) error {
	return mhp.checkProposedHeaders(headers)
}

func (mhp *multipleHeaderProposalsDetector) CheckHeaderHasSameProposerAndRound(
	header data.HeaderHandler,
	round uint64,
	proposer []byte,
) error {
	return mhp.checkHeaderHasSameProposerAndRound(header, round, proposer)
}
