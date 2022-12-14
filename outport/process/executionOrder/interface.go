package executionOrder

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

type mbsGetter interface {
	GetScheduledMBs(header data.HeaderHandler) ([]*block.MiniBlock, error)
}
