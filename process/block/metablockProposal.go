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
