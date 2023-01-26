package receipts

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/stretchr/testify/require"
)

func TestMarshalAndUnmarshalReceiptsHolder(t *testing.T) {
	t.Run("when empty holder", func(t *testing.T) {
		marshaller := &marshal.GogoProtoMarshalizer{}

		holder := holders.NewReceiptsHolder(nil)
		data, err := marshalReceiptsHolder(holder, marshaller)
		require.Nil(t, err)
		require.Equal(t, []byte{}, data)

		unmarshalledHolder, err := unmarshalReceiptsHolder(data, marshaller)
		require.Nil(t, err)
		require.Equal(t, createEmptyReceiptsHolder(), unmarshalledHolder)
	})

	t.Run("when holder has content", func(t *testing.T) {
		marshaller := &marshal.GogoProtoMarshalizer{}

		// One miniblock in batch.Data
		holder := holders.NewReceiptsHolder([]*block.MiniBlock{{SenderShardID: 42}})
		data, err := marshalReceiptsHolder(holder, marshaller)
		require.Nil(t, err)
		require.Equal(t, []byte{0xa, 0x2, 0x18, 42}, data)

		unmarshalledHolder, err := unmarshalReceiptsHolder(data, marshaller)
		require.Nil(t, err)
		require.Equal(t, holder, unmarshalledHolder)

		// More miniblocks in batch.Data
		holder = holders.NewReceiptsHolder([]*block.MiniBlock{{SenderShardID: 42}, {SenderShardID: 43}})
		data, err = marshalReceiptsHolder(holder, marshaller)
		require.Nil(t, err)
		require.Equal(t, []byte{0xa, 0x2, 0x18, 42, 0xa, 0x2, 0x18, 43}, data)
	})

	t.Run("no panic when badly saved data", func(t *testing.T) {
		marshaller := &marshal.GogoProtoMarshalizer{}

		// Bad container (batch)
		holder, err := unmarshalReceiptsHolder([]byte("bad"), marshaller)
		require.ErrorIs(t, err, errCannotUnmarshalReceipts)
		require.Nil(t, holder)

		// Bad content (miniblocks)
		holder, err = unmarshalReceiptsHolder([]byte{0xa, 0x2, 0xff, 42}, marshaller)
		require.ErrorIs(t, err, errCannotUnmarshalReceipts)
		require.Nil(t, holder)
	})
}
