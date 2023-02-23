package block

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type sovereignObserverBlockProcessor struct {
	*shardProcessor
}

// NewSovereignObserverBlockProcessor creates a new sovereign observer block processor
func NewSovereignObserverBlockProcessor(shardProcessor *shardProcessor) (*sovereignObserverBlockProcessor, error) {
	if check.IfNil(shardProcessor) {
		return nil, errNilShardBlockProcessor
	}

	return &sovereignObserverBlockProcessor{
		shardProcessor: shardProcessor,
	}, nil
}

func (sbp *sovereignObserverBlockProcessor) CommitBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	return sbp.shardProcessor.CommitBlock(headerHandler, bodyHandler)

	// TODO: Add here custom functionality in next PRs
}
