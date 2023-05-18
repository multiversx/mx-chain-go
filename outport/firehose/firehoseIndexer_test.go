package firehose

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewFirehoseIndexer(t *testing.T) {
	t.Parallel()

	t.Run("nil io writer, should return error", func(t *testing.T) {
		t.Parallel()

		fi, err := NewFirehoseIndexer(nil, block.NewEmptyBlockCreatorsContainer())
		require.Nil(t, fi)
		require.Equal(t, errNilWriter, err)
	})

	t.Run("nil block creator, should return error", func(t *testing.T) {
		t.Parallel()

		fi, err := NewFirehoseIndexer(&testscommon.IoWriterStub{}, nil)
		require.Nil(t, fi)
		require.Equal(t, errNilBlockCreator, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		fi, err := NewFirehoseIndexer(&testscommon.IoWriterStub{}, block.NewEmptyBlockCreatorsContainer())
		require.Nil(t, err)
		require.NotNil(t, fi)
	})
}

func TestFirehoseIndexer_SaveBlock(t *testing.T) {
	t.Parallel()

	protoMarshaller := &marshal.GogoProtoMarshalizer{}

	t.Run("nil outport block, should return error", func(t *testing.T) {
		t.Parallel()

		fi, _ := NewFirehoseIndexer(&testscommon.IoWriterStub{}, block.NewEmptyBlockCreatorsContainer())

		err := fi.SaveBlock(nil)
		require.Equal(t, errOutportBlock, err)

		err = fi.SaveBlock(&outportcore.OutportBlock{BlockData: nil})
		require.Equal(t, errOutportBlock, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		header := &block.Header{
			Nonce:     1,
			PrevHash:  []byte("prev hash"),
			TimeStamp: 100,
		}
		headerBytes, err := protoMarshaller.Marshal(header)
		require.Nil(t, err)

		outportBlock := &outportcore.OutportBlock{
			BlockData: &outportcore.BlockData{
				HeaderHash:  []byte("hash"),
				HeaderBytes: headerBytes,
				HeaderType:  string(core.ShardHeaderV1),
			},
		}
		outportBlockBytes, err := protoMarshaller.Marshal(outportBlock)
		require.Nil(t, err)

		ioWriterCalledCt := 0
		ioWriter := &testscommon.IoWriterStub{
			WriteCalled: func(p []byte) (n int, err error) {
				defer func() {
					ioWriterCalledCt++
				}()

				switch ioWriterCalledCt {
				case 0:
					require.Equal(t, []byte("FIRE BLOCK_BEGIN 1\n"), p)
				case 1:
					require.Equal(t, []byte(fmt.Sprintf("FIRE BLOCK_END 1 %s 100 %x\n",
						hex.EncodeToString(header.PrevHash),
						outportBlockBytes)), p)
				default:
					require.Fail(t, "should not write again")
				}
				return 0, nil
			},
		}

		container := block.NewEmptyBlockCreatorsContainer()
		_ = container.Add(core.ShardHeaderV1, block.NewEmptyHeaderCreator())

		fi, _ := NewFirehoseIndexer(ioWriter, container)
		err = fi.SaveBlock(outportBlock)
		require.Nil(t, err)
		require.Equal(t, 2, ioWriterCalledCt)
	})
}
