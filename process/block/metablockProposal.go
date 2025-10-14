package block

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
)

// VerifyBlockProposal will be implemented in a further PR
func (mp *metaProcessor) VerifyBlockProposal(
	_ data.HeaderHandler,
	_ data.BodyHandler,
	_ func() time.Duration,
) error {
	return nil
}

// OnProposedBlock calls the OnProposedBlock from transactions pool
func (mp *metaProcessor) OnProposedBlock(
	proposedBody data.BodyHandler,
	proposedHeader data.HeaderHandler,
	proposedHash []byte,
	lastCommittedHeader data.HeaderHandler,
	lastCommittedHash []byte,
) error {
	return mp.baseProcessor.OnProposedBlock(proposedBody, proposedHeader, proposedHash, lastCommittedHeader, lastCommittedHash)
}
