package databasereader_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/mock"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderMarshalizer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	hm, err := databasereader.NewHeaderMarshalizer(nil)
	require.Equal(t, databasereader.ErrNilMarshalizer, err)
	require.True(t, check.IfNil(hm))
}

func TestNewHeaderMarshalizer_ShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hm, err := databasereader.NewHeaderMarshalizer(marshalizer)
	require.NoError(t, err)
	require.False(t, check.IfNil(hm))
}

func TestHeaderMarshalizer_UnmarshalMetaBlock_InvalidBytesShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hm, _ := databasereader.NewHeaderMarshalizer(marshalizer)
	require.NotNil(t, hm)

	metaBlock, err := hm.UnmarshalMetaBlock([]byte("invalid meta block bytes"))
	require.Error(t, err)
	require.Nil(t, metaBlock)
}

func TestHeaderMarshalizer_UnmarshalMetaBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	initialMetaBlock := &block.MetaBlock{Nonce: 77}
	metaBlockBytes, _ := marshalizer.Marshal(initialMetaBlock)

	hm, _ := databasereader.NewHeaderMarshalizer(marshalizer)
	require.NotNil(t, hm)

	metaBlock, err := hm.UnmarshalMetaBlock(metaBlockBytes)
	require.NoError(t, err)
	require.Equal(t, initialMetaBlock, metaBlock)
}

func TestHeaderMarshalizer_UnmarshalShardHeader_InvalidBytesShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hm, _ := databasereader.NewHeaderMarshalizer(marshalizer)
	require.NotNil(t, hm)

	metaBlock, err := hm.UnmarshalShardHeader([]byte("invalid shard headers bytes"))
	require.Error(t, err)
	require.Nil(t, metaBlock)
}

func TestHeaderMarshalizer_UnmarshalShardHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	initialShardHeader := &block.Header{Nonce: 77}
	shardHeaderBytes, _ := marshalizer.Marshal(initialShardHeader)

	hm, _ := databasereader.NewHeaderMarshalizer(marshalizer)
	require.NotNil(t, hm)

	shardHeader, err := hm.UnmarshalShardHeader(shardHeaderBytes)
	require.NoError(t, err)
	require.Equal(t, initialShardHeader, shardHeader)
}
