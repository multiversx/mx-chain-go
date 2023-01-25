package executionOrder

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestMiniblockGetter_GetScheduledMBs(t *testing.T) {
	t.Parallel()

	marshaller := &marshal.GogoProtoMarshalizer{}
	mbhr := &block.MiniBlockHeaderReserved{
		ExecutionType: block.ProcessingType(1),
	}
	mbhrBytes, _ := marshaller.Marshal(mbhr)

	mbHash1, mbHash2, scheduledMbHash := []byte("mb1"), []byte("mb2"), []byte("scheduled")
	headerPrevHeader := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: mbHash1,
			},
			{
				Hash: mbHash2,
			},
			{
				Hash:     scheduledMbHash,
				Reserved: mbhrBytes,
			},
		},
	}

	header := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type: block.InvalidBlock,
			},
		},
	}

	storer := &storage.StorerStub{}
	mbsG := newMiniblocksGetter(storer, marshaller)

	scheduledMbs, err := mbsG.GetScheduledMBs(header, headerPrevHeader)
	require.Nil(t, err)
	require.Equal(t, 1, len(scheduledMbs))
}
