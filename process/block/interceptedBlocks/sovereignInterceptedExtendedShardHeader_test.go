package interceptedBlocks

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/require"
)

func createShardExtendedHeader() *block.ShardHeaderExtended {
	return &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Nonce: 4,
			},
		},
	}
}

func createShardExtendedHeaderBytes(marshaller marshal.Marshalizer) []byte {
	hdr := createShardExtendedHeader()
	hdrBytes, _ := marshaller.Marshal(hdr)

	return hdrBytes
}

func createSovereignShardArgs() ArgsSovereignInterceptedHeader {
	marshaller := &mock.MarshalizerMock{}

	return ArgsSovereignInterceptedHeader{
		Marshaller:  marshaller,
		Hasher:      &hashingMocks.HasherMock{},
		HeaderBytes: createShardExtendedHeaderBytes(marshaller),
	}
}

func TestNewSovereignInterceptedHeader(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createSovereignShardArgs()
		args.Marshaller = nil
		sovHdr, err := NewSovereignInterceptedHeader(args)
		require.Equal(t, process.ErrNilMarshalizer, err)
		require.Nil(t, sovHdr)
	})

	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createSovereignShardArgs()
		args.Hasher = nil
		sovHdr, err := NewSovereignInterceptedHeader(args)
		require.Equal(t, process.ErrNilHasher, err)
		require.Nil(t, sovHdr)
	})

	t.Run("nil extended header bytes, should return error", func(t *testing.T) {
		args := createSovereignShardArgs()
		args.HeaderBytes = nil
		sovHdr, err := NewSovereignInterceptedHeader(args)
		require.NotNil(t, err)
		require.Nil(t, sovHdr)
	})

	t.Run("should work", func(t *testing.T) {
		args := createSovereignShardArgs()
		sovHdr, err := NewSovereignInterceptedHeader(args)
		require.Nil(t, err)
		require.False(t, sovHdr.IsInterfaceNil())
	})
}

func TestSovereignInterceptedHeader_CheckValidity(t *testing.T) {
	t.Parallel()

	args := createSovereignShardArgs()
	sovHdr, _ := NewSovereignInterceptedHeader(args)

	require.Nil(t, sovHdr.CheckValidity())
}

func TestSovereignInterceptedHeader_Hash(t *testing.T) {
	t.Parallel()

	args := createSovereignShardArgs()
	sovHdr, _ := NewSovereignInterceptedHeader(args)

	expectedHash := args.Hasher.Compute(string(args.HeaderBytes))
	require.Equal(t, expectedHash, sovHdr.Hash())
}

func TestSovereignInterceptedHeader_GetExtendedHeader(t *testing.T) {
	t.Parallel()

	args := createSovereignShardArgs()
	sovHdr, _ := NewSovereignInterceptedHeader(args)

	expectedHdr := createShardExtendedHeader()
	require.Equal(t, expectedHdr, sovHdr.GetExtendedHeader())
}

func TestSovereignInterceptedHeader_IsForCurrentShard(t *testing.T) {
	t.Parallel()

	args := createSovereignShardArgs()
	sovHdr, _ := NewSovereignInterceptedHeader(args)
	require.True(t, sovHdr.IsForCurrentShard())
}

func TestSovereignInterceptedHeader_Type(t *testing.T) {
	t.Parallel()

	args := createSovereignShardArgs()
	sovHdr, _ := NewSovereignInterceptedHeader(args)
	require.Equal(t, sovInterceptedHeaderType, sovHdr.Type())
}

func TestSovereignInterceptedHeader_String(t *testing.T) {
	t.Parallel()

	args := createSovereignShardArgs()
	sovHdr, _ := NewSovereignInterceptedHeader(args)

	hdr := createShardExtendedHeader()
	require.Equal(t, fmt.Sprintf("%s, shardId=%d, shardEpoch=%d, round=%d, nonce=%d",
		sovHdr.Type(),
		hdr.GetShardID(),
		hdr.GetEpoch(),
		hdr.GetRound(),
		hdr.GetNonce(),
	), sovHdr.String())
}

func TestSovereignInterceptedHeader_Identifiers(t *testing.T) {
	t.Parallel()

	args := createSovereignShardArgs()
	sovHdr, _ := NewSovereignInterceptedHeader(args)

	hdr := createShardExtendedHeader()
	keyNonce := []byte(fmt.Sprintf("%d-%d", hdr.GetShardID(), hdr.GetNonce()))
	keyEpoch := []byte(core.EpochStartIdentifier(hdr.GetEpoch()))
	keyType := []byte(sovHdr.Type())

	require.Equal(t, [][]byte{keyType, sovHdr.Hash(), keyNonce, keyEpoch}, sovHdr.Identifiers())
}
