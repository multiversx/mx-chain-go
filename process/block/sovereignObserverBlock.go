package block

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
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
	err := osb.shardProcessor.CommitBlock(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	finalizedBlockHash := osb.forkDetector.GetHighestFinalBlockHash()

	shardHeader, err := process.GetShardHeader(finalizedBlockHash, osb.dataPool.Headers(), osb.marshalizer, osb.store)

	for _, mbHash := range shardHeader.GetMiniBlockHeadersHashes() {
		_ = mbHash // Get mb from storer
	}

	return nil
}
