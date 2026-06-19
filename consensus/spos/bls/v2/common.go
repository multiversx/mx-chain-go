package v2

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

func checkGoRoutinesThrottler(
	ctx context.Context,
	signatureThrottler core.Throttler,
) error {
	for {
		if signatureThrottler.CanProcess() {
			break
		}

		select {
		case <-time.After(timeSpentBetweenChecks):
			continue
		case <-ctx.Done():
			return fmt.Errorf("%w while checking the throttler", spos.ErrTimeIsOut)
		}
	}

	return nil
}
