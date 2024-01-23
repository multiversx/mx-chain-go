package incomingHeader

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

type extendedHeaderProcessor struct {
	headersPool HeadersPool
	marshaller  marshal.Marshalizer
	hasher      hashing.Hasher
}

func createExtendedHeader(incomingHeader sovereign.IncomingHeaderHandler, scrs []*scrInfo) (*block.ShardHeaderExtended, error) {
	headerV2, castOk := incomingHeader.GetHeaderHandler().(*block.HeaderV2)
	if !castOk {
		return nil, errInvalidHeaderType
	}
	events, err := getEvents(incomingHeader.GetIncomingEventHandlers())
	if err != nil {
		return nil, err
	}

	return &block.ShardHeaderExtended{
		Header:             headerV2,
		IncomingMiniBlocks: createIncomingMb(scrs),
		IncomingEvents:     events,
	}, nil
}

func getEvents(events []data.EventHandler) ([]*transaction.Event, error) {
	ret := make([]*transaction.Event, len(events))

	for idx, eventHandler := range events {
		event, castOk := eventHandler.(*transaction.Event)
		if !castOk {
			return nil, errInvalidEventType
		}

		ret[idx] = event
	}

	return ret, nil
}

func createIncomingMb(scrs []*scrInfo) []*block.MiniBlock {
	if len(scrs) == 0 {
		return make([]*block.MiniBlock, 0)
	}

	scrHashes := make([][]byte, len(scrs))
	for idx, scrData := range scrs {
		scrHashes[idx] = scrData.hash
	}

	return []*block.MiniBlock{
		{
			TxHashes:        scrHashes,
			ReceiverShardID: core.SovereignChainShardId,
			SenderShardID:   core.MainChainShardId,
			Type:            block.SmartContractResultBlock,
		},
	}
}

func (ehp *extendedHeaderProcessor) addExtendedHeaderToPool(extendedHeader data.ShardHeaderExtendedHandler) error {
	extendedHeaderHash, err := core.CalculateHash(ehp.marshaller, ehp.hasher, extendedHeader)
	if err != nil {
		return err
	}

	ehp.headersPool.AddHeaderInShard(extendedHeaderHash, extendedHeader, core.MainChainShardId)
	return nil
}
