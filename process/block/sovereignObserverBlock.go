package block

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
)

type sovereignObserverBlockProcessor struct {
	*shardProcessor
}

// NewSovereignObserverBlockProcessor creates a new sovereign observer block processor
func NewSovereignObserverBlockProcessor(shardProcessor *shardProcessor) (*sovereignObserverBlockProcessor, error) {
	if shardProcessor.IsInterfaceNil() {
		return nil, core.ErrNilShardCoordinator
	}

	sbp := &sovereignObserverBlockProcessor{
		shardProcessor: shardProcessor,
	}

	return sbp, nil
}

func (osb *sovereignObserverBlockProcessor) CommitBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	return osb.shardProcessor.CommitBlock(headerHandler, bodyHandler)

	// TODO: Add here custom functionality in next PRs
}
