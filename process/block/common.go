package block

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

func getHeaderFromHash(
	headersPool dataRetriever.HeadersPool,
	hdrsForCurrBlock HeadersForBlock,
	isHeaderV3 bool,
	shardHeaderHash []byte,
) (data.HeaderHandler, error) {
	if isHeaderV3 {
		header, err := headersPool.GetHeaderByHash(shardHeaderHash)
		// TODO: debug only, remove after test
		if err != nil {
			log.Error("getHeaderFromHash - failed to get header from headers pool", "hash", shardHeaderHash, "error", err)
		}
		return header, err
	}

	headerInfo, ok := hdrsForCurrBlock.GetHeaderInfo(string(shardHeaderHash))
	if !ok {
		return nil, process.ErrMissingHeader
	}

	return headerInfo.GetHeader(), nil
}
