package incomingHeader

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

type extendedHeaderProcessor struct {
	headersPool HeadersPool
	marshaller  marshal.Marshalizer
	hasher      hashing.Hasher
	selfShardID uint32
}

func (ehp *extendedHeaderProcessor) createExtendedHeader(headerV2 *block.HeaderV2, scrs []*scrInfo) *block.ShardHeaderExtended {
	return &block.ShardHeaderExtended{
		Header:             headerV2,
		IncomingMiniBlocks: ehp.createIncomingMb(scrs),
	}
}

func (ehp *extendedHeaderProcessor) createIncomingMb(scrs []*scrInfo) []*block.MiniBlock {
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
			ReceiverShardID: ehp.selfShardID,
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

	ehp.headersPool.AddHeader(extendedHeaderHash, extendedHeader)
	return nil
}
