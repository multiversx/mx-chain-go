package firehose

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/firehose"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestFirehoseIndexer_SaveBlockHeader(t *testing.T) {
	t.Parallel()

	protoMarshaller := &marshal.GogoProtoMarshalizer{}
	t.Run("nil header, should return error", func(t *testing.T) {
		t.Parallel()

		fi, _ := NewFirehoseIndexer(&testscommon.IoWriterStub{})
		err := fi.SaveBlock(&outportcore.ArgsSaveBlockData{Header: nil})
		require.Equal(t, errNilHeader, err)
	})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		t.Parallel()

		fi, _ := NewFirehoseIndexer(&testscommon.IoWriterStub{})
		err := fi.SaveBlock(&outportcore.ArgsSaveBlockData{Header: &testscommon.HeaderHandlerStub{}})
		require.Equal(t, errInvalidHeaderType, err)
	})

	t.Run("meta header", func(t *testing.T) {
		t.Parallel()

		metaBlockHeader := &block.MetaBlock{
			Nonce:     1,
			PrevHash:  []byte("prevHashMeta"),
			TimeStamp: 100,
		}
		marshalledHeader, err := protoMarshaller.Marshal(metaBlockHeader)
		require.Nil(t, err)

		headerHashMeta := []byte("headerHashMeta")
		firehoseBlock := &firehose.FirehoseBlock{
			HeaderHash:  headerHashMeta,
			HeaderType:  string(core.MetaHeader),
			HeaderBytes: marshalledHeader,
		}
		marshalledFirehoseBlock, err := protoMarshaller.Marshal(firehoseBlock)
		require.Nil(t, err)

		ioWriterCalledCt := 0
		ioWriter := &testscommon.IoWriterStub{
			WriteCalled: func(p []byte) (n int, err error) {
				ioWriterCalledCt++
				switch ioWriterCalledCt {
				case 1:
					require.Equal(t, []byte("FIRE BLOCK_BEGIN 1\n"), p)
				case 2:

					require.Equal(t, []byte(fmt.Sprintf("FIRE BLOCK_END 1 %s 100 %x\n",
						hex.EncodeToString(metaBlockHeader.PrevHash),
						marshalledFirehoseBlock)), p)
				default:
					require.Fail(t, "should not write again")
				}
				return 0, nil
			},
		}

		fi, _ := NewFirehoseIndexer(ioWriter)
		err = fi.SaveBlock(&outportcore.ArgsSaveBlockData{HeaderHash: headerHashMeta, Header: metaBlockHeader})
		require.Nil(t, err)
	})

	t.Run("shard header v1", func(t *testing.T) {
		t.Parallel()

		shardHeaderV1 := &block.Header{
			Nonce:     2,
			PrevHash:  []byte("prevHashV1"),
			TimeStamp: 200,
		}
		marshalledHeader, err := protoMarshaller.Marshal(shardHeaderV1)
		require.Nil(t, err)

		headerHashShardV1 := []byte("headerHashShardV1")
		firehoseBlock := &firehose.FirehoseBlock{
			HeaderHash:  headerHashShardV1,
			HeaderType:  string(core.ShardHeaderV1),
			HeaderBytes: marshalledHeader,
		}
		marshalledFirehoseBlock, err := protoMarshaller.Marshal(firehoseBlock)
		require.Nil(t, err)

		ioWriterCalledCt := 0
		ioWriter := &testscommon.IoWriterStub{
			WriteCalled: func(p []byte) (n int, err error) {
				ioWriterCalledCt++
				switch ioWriterCalledCt {
				case 1:
					require.Equal(t, []byte("FIRE BLOCK_BEGIN 2\n"), p)
				case 2:

					require.Equal(t, []byte(fmt.Sprintf("FIRE BLOCK_END 2 %s 200 %x\n",
						hex.EncodeToString(shardHeaderV1.PrevHash),
						marshalledFirehoseBlock)), p)
				default:
					require.Fail(t, "should not write again")
				}
				return 0, nil
			},
		}

		fi, _ := NewFirehoseIndexer(ioWriter)
		err = fi.SaveBlock(&outportcore.ArgsSaveBlockData{HeaderHash: headerHashShardV1, Header: shardHeaderV1})
		require.Nil(t, err)
	})

	t.Run("shard header v2", func(t *testing.T) {
		t.Parallel()

		shardHeaderV2 := &block.HeaderV2{
			Header: &block.Header{
				Nonce:     3,
				PrevHash:  []byte("prevHashV2"),
				TimeStamp: 300,
			},
		}
		marshalledHeader, err := protoMarshaller.Marshal(shardHeaderV2)
		require.Nil(t, err)

		headerHashShardV2 := []byte("headerHashShardV2")
		firehoseBlock := &firehose.FirehoseBlock{
			HeaderHash:  headerHashShardV2,
			HeaderType:  string(core.ShardHeaderV2),
			HeaderBytes: marshalledHeader,
		}
		marshalledFirehoseBlock, err := protoMarshaller.Marshal(firehoseBlock)
		require.Nil(t, err)

		ioWriterCalledCt := 0
		ioWriter := &testscommon.IoWriterStub{
			WriteCalled: func(p []byte) (n int, err error) {
				ioWriterCalledCt++
				switch ioWriterCalledCt {
				case 1:
					require.Equal(t, []byte("FIRE BLOCK_BEGIN 3\n"), p)
				case 2:

					require.Equal(t, []byte(fmt.Sprintf("FIRE BLOCK_END 3 %s 300 %x\n",
						hex.EncodeToString(shardHeaderV2.Header.PrevHash),
						marshalledFirehoseBlock)), p)
				default:
					require.Fail(t, "should not write again")
				}
				return 0, nil
			},
		}

		fi, _ := NewFirehoseIndexer(ioWriter)
		err = fi.SaveBlock(&outportcore.ArgsSaveBlockData{HeaderHash: headerHashShardV2, Header: shardHeaderV2})
		require.Nil(t, err)
	})
}
