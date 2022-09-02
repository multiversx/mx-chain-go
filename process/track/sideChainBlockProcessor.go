package track

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type sideChainBlockProcessor struct {
	*blockProcessor
}

// NewSideChainBlockProcessor creates an object for processing the received tracked blocks
func NewSideChainBlockProcessor(blockProcessor *blockProcessor) (*sideChainBlockProcessor, error) {
	if blockProcessor == nil {
		return nil, process.ErrNilBlockProcessor
	}

	return &sideChainBlockProcessor{
		blockProcessor,
	}, nil
}

func (scbp *sideChainBlockProcessor) shouldProcessReceivedHeader(headerHandler data.HeaderHandler) bool {
	lastNotarizedHeader, _, err := scbp.selfNotarizer.GetLastNotarizedHeader(headerHandler.GetShardID())
	if err != nil {
		log.Warn("shouldProcessReceivedHeader: selfNotarizer.GetLastNotarizedHeader",
			"shard", headerHandler.GetShardID(), "error", err.Error())
		return false
	}

	shouldProcessReceivedHeader := headerHandler.GetNonce() > lastNotarizedHeader.GetNonce()
	return shouldProcessReceivedHeader
}
