package executionOrder

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestMiniblockGetter_GetScheduledMBs(t *testing.T) {
	t.Parallel()

	marshalizer := &marshal.GogoProtoMarshalizer{}
	mbhr := &block.MiniBlockHeaderReserved{
		ExecutionType: block.ProcessingType(1),
	}
	mbhrBytes, _ := marshalizer.Marshal(mbhr)

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
	mbsG := newMiniblocksGetter(storer, marshalizer)

	scheduledMbs, err := mbsG.GetScheduledMBs(header, headerPrevHeader)
	require.Nil(t, err)
	require.Equal(t, 1, len(scheduledMbs))
}
