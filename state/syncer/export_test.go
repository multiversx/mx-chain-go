package syncer

import (
	"context"

	"github.com/multiversx/mx-chain-go/common"
)

// UserAccountsSyncer -
type UserAccountsSyncer = userAccountsSyncer

// ValidatorAccountsSyncer -
type ValidatorAccountsSyncer = validatorAccountsSyncer

// CheckBaseAccountsSyncerArgs -
func CheckBaseAccountsSyncerArgs(args ArgsNewBaseAccountsSyncer) error {
	return checkArgs(args)
}

// SyncAccountDataTries -
func (u *userAccountsSyncer) SyncAccountDataTries(
	leavesChannels *common.TrieIteratorChannels,
	ctx context.Context,
) error {
	return u.syncAccountDataTries(leavesChannels, ctx)
}

// GetNumHandlers -
func (mtnn *missingTrieNodesNotifier) GetNumHandlers() int {
	return len(mtnn.handlers)
}
