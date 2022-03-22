package processor

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	heartbeatMessages "github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgShardValidatorInfoInterceptorProcessor() ArgShardValidatorInfoInterceptorProcessor {
	return ArgShardValidatorInfoInterceptorProcessor{
		Marshaller:      testscommon.MarshalizerMock{},
		PeerShardMapper: &mock.PeerShardMapperStub{},
	}
}

func TestNewShardValidatorInfoInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgShardValidatorInfoInterceptorProcessor()
		args.Marshaller = nil

		processor, err := NewShardValidatorInfoInterceptorProcessor(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(processor))
	})
	t.Run("nil peer shard mapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgShardValidatorInfoInterceptorProcessor()
		args.PeerShardMapper = nil

		processor, err := NewShardValidatorInfoInterceptorProcessor(args)
		assert.Equal(t, process.ErrNilPeerShardMapper, err)
		assert.True(t, check.IfNil(processor))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processor, err := NewShardValidatorInfoInterceptorProcessor(createMockArgShardValidatorInfoInterceptorProcessor())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))
	})
}

func Test_shardValidatorInfoInterceptorProcessor_Save(t *testing.T) {
	t.Parallel()

	t.Run("invalid message should error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgShardValidatorInfoInterceptorProcessor()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				wasCalled = true
			},
		}

		processor, err := NewShardValidatorInfoInterceptorProcessor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		// provide heartbeat as intercepted data
		arg := heartbeat.ArgInterceptedHeartbeat{
			ArgBaseInterceptedHeartbeat: heartbeat.ArgBaseInterceptedHeartbeat{
				Marshalizer: &mock.MarshalizerMock{},
			},
			PeerId: "pid",
		}
		arg.DataBuff, _ = arg.Marshalizer.Marshal(heartbeatMessages.HeartbeatV2{})
		ihb, _ := heartbeat.NewInterceptedHeartbeat(arg)

		err = processor.Save(ihb, "", "")
		assert.Equal(t, process.ErrWrongTypeAssertion, err)
		assert.False(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgShardValidatorInfoInterceptorProcessor()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				wasCalled = true
			},
		}

		processor, err := NewShardValidatorInfoInterceptorProcessor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		msg := message.ShardValidatorInfo{
			ShardId: 5,
		}
		dataBuff, _ := args.Marshaller.Marshal(msg)
		arg := p2p.ArgInterceptedShardValidatorInfo{
			Marshaller:  args.Marshaller,
			DataBuff:    dataBuff,
			NumOfShards: 10,
		}
		data, _ := p2p.NewInterceptedShardValidatorInfo(arg)

		err = processor.Save(data, "", "")
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func Test_shardValidatorInfoInterceptorProcessor_DisabledMethods(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	processor, err := NewShardValidatorInfoInterceptorProcessor(createMockArgShardValidatorInfoInterceptorProcessor())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(processor))

	err = processor.Validate(nil, "")
	assert.Nil(t, err)

	processor.RegisterHandler(nil)

}
