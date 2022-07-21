package receipts

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestMarshalAndUnmarshalReceiptsHolder(t *testing.T) {
	t.Run("when empty holder", func(t *testing.T) {
		marshaller := &marshal.GogoProtoMarshalizer{}

		// Miniblocks as nil slice
		holder := &process.ReceiptsHolder{}
		data, err := marshalReceiptsHolder(holder, marshaller)
		require.Nil(t, err)
		require.Equal(t, []byte{}, data)

		unmarshalledHolder, err := unmarshalReceiptsHolder(data, marshaller)
		require.Nil(t, err)
		require.Equal(t, newReceiptsHolder(), unmarshalledHolder)

		// Miniblocks as empty slice
		holder = &process.ReceiptsHolder{Miniblocks: make([]*block.MiniBlock, 0)}
		data, err = marshalReceiptsHolder(holder, marshaller)
		require.Nil(t, err)
		require.Equal(t, []byte{}, data)

		unmarshalledHolder, err = unmarshalReceiptsHolder(data, marshaller)
		require.Nil(t, err)
		require.Equal(t, newReceiptsHolder(), unmarshalledHolder)
	})

	t.Run("when holder has content", func(t *testing.T) {
		marshaller := &marshal.GogoProtoMarshalizer{}

		// One miniblock in batch.Data
		holder := &process.ReceiptsHolder{Miniblocks: []*block.MiniBlock{{SenderShardID: 42}}}
		data, err := marshalReceiptsHolder(holder, marshaller)
		require.Nil(t, err)
		require.Equal(t, []byte{0xa, 0x2, 0x18, 42}, data)

		unmarshalledHolder, err := unmarshalReceiptsHolder(data, marshaller)
		require.Nil(t, err)
		require.Equal(t, holder, unmarshalledHolder)

		// More miniblocks in batch.Data
		holder = &process.ReceiptsHolder{Miniblocks: []*block.MiniBlock{{SenderShardID: 42}, {SenderShardID: 43}}}
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
