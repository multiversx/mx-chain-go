package v2

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/consensus/spos"
)

func checkGoRoutinesThrottler(ctx context.Context, throttler core.Throttler) error {
	for {
		if throttler.CanProcess() {
			break
		}

		select {
		case <-time.After(time.Millisecond):
			continue
		case <-ctx.Done():
			return fmt.Errorf("%w while checking the throttler", spos.ErrTimeIsOut)
		}
	}
	return nil
}
