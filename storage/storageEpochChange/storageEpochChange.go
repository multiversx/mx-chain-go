package storageEpochChange

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("storage")

const (
	// WaitTimeForSnapshotEpochCheck is the time to wait before checking the storage epoch
	WaitTimeForSnapshotEpochCheck = time.Millisecond * 100

	// SnapshotWaitTimeout is the timeout for waiting for the storage epoch to change
	SnapshotWaitTimeout = time.Minute * 3
)

// StorageEpochChangeWaitArgs are the args needed for calling the WaitForStorageEpochChange function
type StorageEpochChangeWaitArgs struct {
	TrieStorageManager            common.StorageManager
	Epoch                         uint32
	WaitTimeForSnapshotEpochCheck time.Duration
	SnapshotWaitTimeout           time.Duration
}

// WaitForStorageEpochChange waits for the storage epoch to change to the given epoch
func WaitForStorageEpochChange(args StorageEpochChangeWaitArgs) error {
	log.Debug("waiting for storage epoch change", "epoch", args.Epoch, "wait timeout", args.SnapshotWaitTimeout)

	if args.SnapshotWaitTimeout < args.WaitTimeForSnapshotEpochCheck {
		return fmt.Errorf("timeout (%s) must be greater than wait time between snapshot epoch check (%s)", args.SnapshotWaitTimeout, args.WaitTimeForSnapshotEpochCheck)
	}

	ctx, cancel := context.WithTimeout(context.Background(), args.SnapshotWaitTimeout)
	defer cancel()

	timer := time.NewTimer(args.WaitTimeForSnapshotEpochCheck)
	defer timer.Stop()

	for {
		timer.Reset(args.WaitTimeForSnapshotEpochCheck)

		if args.TrieStorageManager.IsClosed() {
			return core.ErrContextClosing
		}

		latestStorageEpoch, err := args.TrieStorageManager.GetLatestStorageEpoch()
		if err != nil {
			return err
		}

		if latestStorageEpoch == args.Epoch {
			return nil
		}

		select {
		case <-timer.C:
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for storage epoch change, snapshot epoch %d", args.Epoch)
		}
	}
}
