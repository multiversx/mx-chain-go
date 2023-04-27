package track

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
)

type extendedShardHeaderRequestHandler interface {
	RequestExtendedShardHeaderByNonce(nonce uint64)
}

type sovereignChainBlockProcessor struct {
	*blockProcessor
	extendedShardHeaderRequester extendedShardHeaderRequestHandler
}

// NewSovereignChainBlockProcessor creates an object for processing the received tracked blocks
func NewSovereignChainBlockProcessor(blockProcessor *blockProcessor) (*sovereignChainBlockProcessor, error) {
	if blockProcessor == nil {
		return nil, process.ErrNilBlockProcessor
	}

	scbp := &sovereignChainBlockProcessor{
		blockProcessor: blockProcessor,
	}

	scbp.shouldProcessReceivedHeaderFunc = scbp.shouldProcessReceivedHeader
	scbp.processReceivedHeaderFunc = scbp.processReceivedHeader
	scbp.doJobOnReceivedCrossNotarizedHeaderFunc = scbp.doJobOnReceivedCrossNotarizedHeader
	scbp.requestHeaderWithShardAndNonceFunc = scbp.requestHeaderWithShardAndNonce

	extendedShardHeaderRequester, ok := scbp.requestHandler.(extendedShardHeaderRequestHandler)
	if !ok {
		return nil, fmt.Errorf("%w in NewSovereignChainBlockProcessor", process.ErrWrongTypeAssertion)
	}

	scbp.extendedShardHeaderRequester = extendedShardHeaderRequester

	return scbp, nil
}

func (scbp *sovereignChainBlockProcessor) shouldProcessReceivedHeader(headerHandler data.HeaderHandler) bool {
	var lastNotarizedHeader data.HeaderHandler
	var err error

	_, isExtendedShardHeaderReceived := headerHandler.(*block.ShardHeaderExtended)
	if isExtendedShardHeaderReceived {
		lastNotarizedHeader, _, err = scbp.crossNotarizer.GetLastNotarizedHeader(core.SovereignChainShardId)
		if err != nil {
			log.Warn("shouldProcessReceivedHeader: crossNotarizer.GetLastNotarizedHeader",
				"shard", headerHandler.GetShardID(), "error", err.Error())
			return false
		}
	} else {
		lastNotarizedHeader, _, err = scbp.selfNotarizer.GetLastNotarizedHeader(headerHandler.GetShardID())
		if err != nil {
			log.Warn("shouldProcessReceivedHeader: selfNotarizer.GetLastNotarizedHeader",
				"shard", headerHandler.GetShardID(), "error", err.Error())
			return false
		}
	}

	shouldProcessReceivedHeader := headerHandler.GetNonce() > lastNotarizedHeader.GetNonce()
	return shouldProcessReceivedHeader
}

func (scbp *sovereignChainBlockProcessor) processReceivedHeader(headerHandler data.HeaderHandler) {
	_, isExtendedShardHeaderReceived := headerHandler.(*block.ShardHeaderExtended)
	if isExtendedShardHeaderReceived {
		scbp.doJobOnReceivedCrossNotarizedHeaderFunc(core.SovereignChainShardId)
		return
	}

	scbp.doJobOnReceivedHeader(headerHandler.GetShardID())
}

func (scbp *sovereignChainBlockProcessor) doJobOnReceivedCrossNotarizedHeader(shardID uint32) {
	_, _, crossNotarizedHeaders, crossNotarizedHeadersHashes := scbp.computeLongestChainFromLastCrossNotarized(shardID)
	if len(crossNotarizedHeaders) > 0 {
		scbp.crossNotarizedHeadersNotifier.CallHandlers(shardID, crossNotarizedHeaders, crossNotarizedHeadersHashes)
	}
}

func (scbp *sovereignChainBlockProcessor) requestHeaderWithShardAndNonce(shardID uint32, nonce uint64) {
	if shardID == core.SovereignChainShardId {
		scbp.extendedShardHeaderRequester.RequestExtendedShardHeaderByNonce(nonce)
	} else {
		scbp.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
	}
}
