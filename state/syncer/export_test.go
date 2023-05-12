package syncer

import (
	"context"

	"github.com/multiversx/mx-chain-go/common"
)

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
