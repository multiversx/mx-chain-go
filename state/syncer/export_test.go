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
	mainTrie common.Trie,
	ctx context.Context,
) error {
	return u.syncAccountDataTries(mainTrie, ctx)
}
