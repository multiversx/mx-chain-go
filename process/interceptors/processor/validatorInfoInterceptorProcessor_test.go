package processor

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	heartbeatMessages "github.com/ElrondNetwork/elrond-go/heartbeat"
	heartbeatMocks "github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/p2p"
	"github.com/stretchr/testify/assert"
)

func createMockArgValidatorInfoInterceptorProcessor() ArgValidatorInfoInterceptorProcessor {
	return ArgValidatorInfoInterceptorProcessor{
		PeerShardMapper: &mock.PeerShardMapperStub{},
	}
}

func TestNewValidatorInfoInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil peer shard mapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoInterceptorProcessor()
		args.PeerShardMapper = nil

		processor, err := NewValidatorInfoInterceptorProcessor(args)
		assert.Equal(t, process.ErrNilPeerShardMapper, err)
		assert.True(t, check.IfNil(processor))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processor, err := NewValidatorInfoInterceptorProcessor(createMockArgValidatorInfoInterceptorProcessor())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))
	})
}

func TestValidatorInfoInterceptorProcessor_Save(t *testing.T) {
	t.Parallel()

	t.Run("invalid message should error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgValidatorInfoInterceptorProcessor()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				wasCalled = true
			},
		}

		processor, err := NewValidatorInfoInterceptorProcessor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		// provide heartbeat as intercepted data
		arg := heartbeat.ArgInterceptedHeartbeat{
			ArgBaseInterceptedHeartbeat: heartbeat.ArgBaseInterceptedHeartbeat{
				Marshalizer: &mock.MarshalizerMock{},
			},
			PeerId: "pid",
		}
		arg.DataBuff, _ = arg.Marshalizer.Marshal(&heartbeatMessages.HeartbeatV2{})
		ihb, _ := heartbeat.NewInterceptedHeartbeat(arg)

		err = processor.Save(ihb, "", "")
		assert.Equal(t, process.ErrWrongTypeAssertion, err)
		assert.False(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgValidatorInfoInterceptorProcessor()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				wasCalled = true
			},
		}

		processor, err := NewValidatorInfoInterceptorProcessor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		msg := &message.ShardValidatorInfo{
			ShardId: 5,
		}
		marshaller := heartbeatMocks.MarshallerMock{}
		dataBuff, _ := marshaller.Marshal(msg)
		arg := p2p.ArgInterceptedValidatorInfo{
			Marshaller:  &marshaller,
			DataBuff:    dataBuff,
			NumOfShards: 10,
		}
		data, _ := p2p.NewInterceptedValidatorInfo(arg)

		err = processor.Save(data, "", "")
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestValidatorInfoInterceptorProcessor_DisabledMethod(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	processor, err := NewValidatorInfoInterceptorProcessor(createMockArgValidatorInfoInterceptorProcessor())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(processor))

	err = processor.Validate(nil, "")
	assert.Nil(t, err)

	processor.RegisterHandler(nil)

}
