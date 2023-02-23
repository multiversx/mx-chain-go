package block

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
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
	err := sbp.shardProcessor.CommitBlock(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	finalizedBlockHash := sbp.forkDetector.GetHighestFinalBlockHash()

	shardHeader, err := process.GetShardHeader(finalizedBlockHash, sbp.dataPool.Headers(), sbp.marshalizer, sbp.store)

	for _, mbHash := range shardHeader.GetMiniBlockHeadersHashes() {
		_ = mbHash // Get mb from storer
	}

	return nil
}
