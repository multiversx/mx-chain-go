package block

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
)

type sovereignShardCrossNotarizer struct {
	*baseBlockNotarizer
}

func (scn *sovereignShardCrossNotarizer) getLastCrossNotarizedHeaders() []bootstrapStorage.BootstrapHeaderInfo {
	bootstrapHeaderInfo := scn.getLastCrossNotarizedHeadersForShard(core.SovereignChainShardId)
	if bootstrapHeaderInfo == nil {
		return nil
	}

	lastCrossNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0, 1)
	lastCrossNotarizedHeaders = append(lastCrossNotarizedHeaders, *bootstrapHeaderInfo)
	return trimSliceBootstrapHeaderInfo(lastCrossNotarizedHeaders)
}
