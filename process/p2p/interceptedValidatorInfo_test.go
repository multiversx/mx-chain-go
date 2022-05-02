package p2p

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

const providedShard = uint32(5)

func createMockArgInterceptedValidatorInfo() ArgInterceptedValidatorInfo {
	marshaller := testscommon.MarshalizerMock{}
	msg := &message.ShardValidatorInfo{
		ShardId: providedShard,
	}
	msgBuff, _ := marshaller.Marshal(msg)

	return ArgInterceptedValidatorInfo{
		Marshaller:  marshaller,
		DataBuff:    msgBuff,
		NumOfShards: 10,
	}
}
func TestNewInterceptedValidatorInfo(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		args.Marshaller = nil

		isvi, err := NewInterceptedValidatorInfo(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(isvi))
	})
	t.Run("nil data buff should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		args.DataBuff = nil

		isvi, err := NewInterceptedValidatorInfo(args)
		assert.Equal(t, process.ErrNilBuffer, err)
		assert.True(t, check.IfNil(isvi))
	})
	t.Run("invalid num of shards should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		args.NumOfShards = 0

		isvi, err := NewInterceptedValidatorInfo(args)
		assert.Equal(t, process.ErrInvalidValue, err)
		assert.True(t, check.IfNil(isvi))
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		args.DataBuff = []byte("invalid data")

		isvi, err := NewInterceptedValidatorInfo(args)
		assert.NotNil(t, err)
		assert.True(t, check.IfNil(isvi))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		isvi, err := NewInterceptedValidatorInfo(createMockArgInterceptedValidatorInfo())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(isvi))
	})
}

func Test_interceptedValidatorInfo_CheckValidity(t *testing.T) {
	t.Parallel()

	t.Run("invalid shard should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		args.NumOfShards = providedShard - 1

		isvi, err := NewInterceptedValidatorInfo(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(isvi))

		err = isvi.CheckValidity()
		assert.Equal(t, process.ErrInvalidValue, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		isvi, err := NewInterceptedValidatorInfo(createMockArgInterceptedValidatorInfo())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(isvi))

		err = isvi.CheckValidity()
		assert.Nil(t, err)
	})
}

func Test_interceptedValidatorInfo_Getters(t *testing.T) {
	t.Parallel()

	isvi, err := NewInterceptedValidatorInfo(createMockArgInterceptedValidatorInfo())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(isvi))

	assert.True(t, isvi.IsForCurrentShard())
	assert.True(t, bytes.Equal([]byte(""), isvi.Hash()))
	assert.Equal(t, interceptedValidatorInfoType, isvi.Type())
	identifiers := isvi.Identifiers()
	assert.Equal(t, 1, len(identifiers))
	assert.True(t, bytes.Equal([]byte(""), identifiers[0]))
	assert.Equal(t, fmt.Sprintf("shard=%d", providedShard), isvi.String())
	assert.Equal(t, providedShard, isvi.ShardID())
}
