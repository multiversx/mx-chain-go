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
	headerHash []byte,
) (data.HeaderHandler, error) {
	if isHeaderV3 {
		header, err := headersPool.GetHeaderByHash(headerHash)
		// TODO: debug only, remove after test
		if err != nil {
			log.Error("getHeaderFromHash - failed to get header from headers pool", "hash", headerHash, "error", err)
		}
		return header, err
	}

	headerInfo, ok := hdrsForCurrBlock.GetHeaderInfo(string(headerHash))
	if !ok {
		return nil, process.ErrMissingHeader
	}

	return headerInfo.GetHeader(), nil
}
