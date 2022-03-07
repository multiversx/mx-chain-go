package processor_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	heartbeatMessages "github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createHeartbeatInterceptorProcessArg() processor.ArgHeartbeatInterceptorProcessor {
	return processor.ArgHeartbeatInterceptorProcessor{
		HeartbeatCacher:  testscommon.NewCacherStub(),
		ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
		PeerShardMapper:  &p2pmocks.NetworkShardingCollectorStub{},
	}
}

func createInterceptedHeartbeat() *heartbeatMessages.HeartbeatV2 {
	payload := &heartbeatMessages.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: "hardfork message",
	}
	marshalizer := mock.MarshalizerMock{}
	payloadBytes, _ := marshalizer.Marshal(payload)

	return &heartbeatMessages.HeartbeatV2{
		Payload:         payloadBytes,
		VersionNumber:   "version number",
		NodeDisplayName: "node display name",
		Identity:        "identity",
		Nonce:           123,
		PeerSubType:     uint32(core.RegularPeer),
	}
}

func createMockInterceptedHeartbeat() process.InterceptedData {
	arg := heartbeat.ArgInterceptedHeartbeat{
		ArgBaseInterceptedHeartbeat: heartbeat.ArgBaseInterceptedHeartbeat{
			Marshalizer: &mock.MarshalizerMock{},
		},
		PeerId: "pid",
	}
	arg.DataBuff, _ = arg.Marshalizer.Marshal(createInterceptedHeartbeat())
	ihb, _ := heartbeat.NewInterceptedHeartbeat(arg)

	return ihb
}

func TestNewHeartbeatInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil cacher should error", func(t *testing.T) {
		t.Parallel()

		arg := createHeartbeatInterceptorProcessArg()
		arg.HeartbeatCacher = nil
		hip, err := processor.NewHeartbeatInterceptorProcessor(arg)
		assert.Equal(t, process.ErrNilHeartbeatCacher, err)
		assert.Nil(t, hip)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		arg := createHeartbeatInterceptorProcessArg()
		arg.ShardCoordinator = nil
		hip, err := processor.NewHeartbeatInterceptorProcessor(arg)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.Nil(t, hip)
	})
	t.Run("nil peer shard mapper should error", func(t *testing.T) {
		t.Parallel()

		arg := createHeartbeatInterceptorProcessArg()
		arg.PeerShardMapper = nil
		hip, err := processor.NewHeartbeatInterceptorProcessor(arg)
		assert.Equal(t, process.ErrNilPeerShardMapper, err)
		assert.Nil(t, hip)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hip, err := processor.NewHeartbeatInterceptorProcessor(createHeartbeatInterceptorProcessArg())
		assert.Nil(t, err)
		assert.False(t, hip.IsInterfaceNil())
	})
}

func TestHeartbeatInterceptorProcessor_Save(t *testing.T) {
	t.Parallel()

	t.Run("invalid data should error", func(t *testing.T) {
		t.Parallel()

		hip, err := processor.NewHeartbeatInterceptorProcessor(createHeartbeatInterceptorProcessArg())
		assert.Nil(t, err)
		assert.False(t, hip.IsInterfaceNil())
		assert.Equal(t, process.ErrWrongTypeAssertion, hip.Save(nil, "", ""))
	})
	t.Run("invalid heartbeat data should error", func(t *testing.T) {
		t.Parallel()

		providedData := createMockInterceptedPeerAuthentication() // unable to cast to intercepted heartbeat
		wasUpdatePeerIdShardIdCalled := false
		wasUpdatePeerIdSubTypeCalled := false
		args := createHeartbeatInterceptorProcessArg()
		args.PeerShardMapper = &p2pmocks.NetworkShardingCollectorStub{
			UpdatePeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				wasUpdatePeerIdShardIdCalled = true
			},
			UpdatePeerIdSubTypeCalled: func(pid core.PeerID, peerSubType core.P2PPeerSubType) {
				wasUpdatePeerIdSubTypeCalled = true
			},
		}

		paip, err := processor.NewHeartbeatInterceptorProcessor(args)
		assert.Nil(t, err)
		assert.False(t, paip.IsInterfaceNil())
		assert.Equal(t, process.ErrWrongTypeAssertion, paip.Save(providedData, "", ""))
		assert.False(t, wasUpdatePeerIdShardIdCalled)
		assert.False(t, wasUpdatePeerIdSubTypeCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHb := createMockInterceptedHeartbeat()
		wasCalled := false
		providedPid := core.PeerID("pid")
		arg := createHeartbeatInterceptorProcessArg()
		arg.HeartbeatCacher = &testscommon.CacherStub{
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				assert.True(t, bytes.Equal(providedPid.Bytes(), key))
				ihb := value.(heartbeatMessages.HeartbeatV2)
				providedHbHandler := providedHb.(interceptedDataHandler)
				providedHbMessage := providedHbHandler.Message().(heartbeatMessages.HeartbeatV2)
				assert.Equal(t, providedHbMessage.Identity, ihb.Identity)
				assert.Equal(t, providedHbMessage.Payload, ihb.Payload)
				assert.Equal(t, providedHbMessage.NodeDisplayName, ihb.NodeDisplayName)
				assert.Equal(t, providedHbMessage.PeerSubType, ihb.PeerSubType)
				assert.Equal(t, providedHbMessage.VersionNumber, ihb.VersionNumber)
				assert.Equal(t, providedHbMessage.Nonce, ihb.Nonce)
				wasCalled = true
				return false
			},
		}
		wasUpdatePeerIdShardIdCalled := false
		wasUpdatePeerIdSubTypeCalled := false
		arg.PeerShardMapper = &p2pmocks.NetworkShardingCollectorStub{
			UpdatePeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				wasUpdatePeerIdShardIdCalled = true
				assert.Equal(t, providedPid, pid)
			},
			UpdatePeerIdSubTypeCalled: func(pid core.PeerID, peerSubType core.P2PPeerSubType) {
				wasUpdatePeerIdSubTypeCalled = true
				assert.Equal(t, providedPid, pid)
			},
		}

		hip, err := processor.NewHeartbeatInterceptorProcessor(arg)
		assert.Nil(t, err)
		assert.False(t, hip.IsInterfaceNil())

		err = hip.Save(providedHb, providedPid, "")
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		assert.True(t, wasUpdatePeerIdShardIdCalled)
		assert.True(t, wasUpdatePeerIdSubTypeCalled)
	})
}

func TestHeartbeatInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	hip, err := processor.NewHeartbeatInterceptorProcessor(createHeartbeatInterceptorProcessArg())
	assert.Nil(t, err)
	assert.False(t, hip.IsInterfaceNil())
	assert.Nil(t, hip.Validate(nil, ""))
}

func TestHeartbeatInterceptorProcessor_RegisterHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	hip, err := processor.NewHeartbeatInterceptorProcessor(createHeartbeatInterceptorProcessArg())
	assert.Nil(t, err)
	assert.False(t, hip.IsInterfaceNil())
	hip.RegisterHandler(nil)
}
