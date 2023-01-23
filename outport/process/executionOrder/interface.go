package executionOrder

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type mbsGetter interface {
	GetScheduledMBs(currentHeader, prevHeader data.HeaderHandler) ([]*block.MiniBlock, error)
}
