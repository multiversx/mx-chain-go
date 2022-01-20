package detector

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

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

func (mhs *multipleHeaderSigningDetector) CheckSignedHeaders(pubKey []byte, headers slash.HeaderList) error {
	return mhs.checkSignedHeaders(pubKey, headers)
}

func (mhs *multipleHeaderSigningDetector) SignedHeader(pubKey []byte, header data.HeaderHandler) bool {
	return mhs.signedHeader(pubKey, header)
}
