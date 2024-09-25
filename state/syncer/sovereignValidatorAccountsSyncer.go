package syncer

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// NewSovereignValidatorAccountsSyncer creates a validator account syncer for sovereign
func NewSovereignValidatorAccountsSyncer(args ArgsNewValidatorAccountsSyncer) (*validatorAccountsSyncer, error) {
	return newValidatorAccountsSyncer(args, core.SovereignChainShardId)
}
