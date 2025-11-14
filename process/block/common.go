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
		hdr, err := headersPool.GetHeaderByHash(shardHeaderHash)
		if err != nil {
			return nil, err
		}

		return hdr, nil
	}

	headerInfo, ok := hdrsForCurrBlock.GetHeaderInfo(string(shardHeaderHash))
	if !ok {
		return nil, process.ErrMissingHeader
	}

	return headerInfo.GetHeader(), nil
}
