package block

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/factory/mock"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/stretchr/testify/require"
)

func TestSovereignShardCrossNotarizer_getLastCrossNotarizedHeaders(t *testing.T) {
	hash := []byte("hash")
	header := &block.SovereignChainHeader{
		Header: &block.Header{
			ShardID: core.SovereignChainShardId,
			Nonce:   4,
		},
	}
	sovereignNotarzier := &sovereignShardCrossNotarizer{
		&baseBlockNotarizer{
			blockTracker: &mock.BlockTrackerStub{
				GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
					require.Equal(t, core.MainChainShardId, shardID)
					return header, hash, nil
				},
			},
		},
	}

	headers := sovereignNotarzier.getLastCrossNotarizedHeaders()
	expectedHeaders := []bootstrapStorage.BootstrapHeaderInfo{
		{
			ShardId: header.GetShardID(),
			Nonce:   header.GetNonce(),
			Hash:    hash,
		},
	}
	require.Equal(t, expectedHeaders, headers)
}
