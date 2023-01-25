package p2p

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/assert"
)

const providedShard = "5"

func createMockArgInterceptedPeerShard() ArgInterceptedPeerShard {
	marshaller := &marshal.GogoProtoMarshalizer{}
	msg := &p2pFactory.PeerShard{
		ShardId: providedShard,
	}
	msgBuff, _ := marshaller.Marshal(msg)

	return ArgInterceptedPeerShard{
		Marshaller:  marshaller,
		DataBuff:    msgBuff,
		NumOfShards: 10,
	}
}
func TestNewInterceptedPeerShard(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedPeerShard()
		args.Marshaller = nil

		idci, err := NewInterceptedPeerShard(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(idci))
	})
	t.Run("nil data buff should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedPeerShard()
		args.DataBuff = nil

		idci, err := NewInterceptedPeerShard(args)
		assert.Equal(t, process.ErrNilBuffer, err)
		assert.True(t, check.IfNil(idci))
	})
	t.Run("invalid num of shards should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedPeerShard()
		args.NumOfShards = 0

		idci, err := NewInterceptedPeerShard(args)
		assert.Equal(t, process.ErrInvalidValue, err)
		assert.True(t, check.IfNil(idci))
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedPeerShard()
		args.DataBuff = []byte("invalid data")

		idci, err := NewInterceptedPeerShard(args)
		assert.NotNil(t, err)
		assert.True(t, check.IfNil(idci))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		idci, err := NewInterceptedPeerShard(createMockArgInterceptedPeerShard())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(idci))
	})
}

func TestInterceptedPeerShard_CheckValidity(t *testing.T) {
	t.Parallel()

	t.Run("invalid shard string should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedPeerShard()
		msg := &p2pFactory.PeerShard{
			ShardId: "invalid shard",
		}
		msgBuff, _ := args.Marshaller.Marshal(msg)
		args.DataBuff = msgBuff
		idci, err := NewInterceptedPeerShard(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(idci))

		err = idci.CheckValidity()
		assert.NotNil(t, err)
	})
	t.Run("invalid shard should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedPeerShard()
		ps, _ := strconv.ParseInt(providedShard, 10, 32)
		args.NumOfShards = uint32(ps - 1)

		idci, err := NewInterceptedPeerShard(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(idci))

		err = idci.CheckValidity()
		assert.Equal(t, process.ErrInvalidValue, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		idci, err := NewInterceptedPeerShard(createMockArgInterceptedPeerShard())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(idci))

		err = idci.CheckValidity()
		assert.Nil(t, err)
	})
}

func TestInterceptedPeerShard_Getters(t *testing.T) {
	t.Parallel()

	idci, err := NewInterceptedPeerShard(createMockArgInterceptedPeerShard())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(idci))

	assert.True(t, idci.IsForCurrentShard())
	assert.True(t, bytes.Equal([]byte(""), idci.Hash()))
	assert.Equal(t, interceptedPeerShardType, idci.Type())
	identifiers := idci.Identifiers()
	assert.Equal(t, 1, len(identifiers))
	assert.True(t, bytes.Equal([]byte(""), identifiers[0]))
	assert.Equal(t, fmt.Sprintf("shard=%s", providedShard), idci.String())
	assert.Equal(t, providedShard, idci.ShardID())
}
