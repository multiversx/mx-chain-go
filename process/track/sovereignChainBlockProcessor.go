package track

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type sovereignChainBlockProcessor struct {
	*blockProcessor
}

// NewSovereignChainBlockProcessor creates an object for processing the received tracked blocks
func NewSovereignChainBlockProcessor(blockProcessor *blockProcessor) (*sovereignChainBlockProcessor, error) {
	if blockProcessor == nil {
		return nil, process.ErrNilBlockProcessor
	}

	scbp := &sovereignChainBlockProcessor{
		blockProcessor,
	}

	scbp.shouldProcessReceivedHeaderFunc = scbp.shouldProcessReceivedHeader

	return scbp, nil
}

func (scbp *sovereignChainBlockProcessor) shouldProcessReceivedHeader(headerHandler data.HeaderHandler) bool {
	lastNotarizedHeader, _, err := scbp.selfNotarizer.GetLastNotarizedHeader(headerHandler.GetShardID())
	if err != nil {
		log.Warn("shouldProcessReceivedHeader: selfNotarizer.GetLastNotarizedHeader",
			"shard", headerHandler.GetShardID(), "error", err.Error())
		return false
	}

	shouldProcessReceivedHeader := headerHandler.GetNonce() > lastNotarizedHeader.GetNonce()
	return shouldProcessReceivedHeader
}
