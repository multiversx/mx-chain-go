package processor_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	heartbeatMessages "github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/heartbeat"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
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
	marshaller := mock.MarshalizerMock{}
	payloadBytes, _ := marshaller.Marshal(payload)

	return &heartbeatMessages.HeartbeatV2{
		Payload:         payloadBytes,
		VersionNumber:   "version number",
		NodeDisplayName: "node display name",
		Identity:        "identity",
		Nonce:           123,
		PeerSubType:     uint32(core.RegularPeer),
		Pubkey:          []byte("public key"),
	}
}

func createMockInterceptedHeartbeat() process.InterceptedData {
	arg := heartbeat.ArgBaseInterceptedHeartbeat{
		Marshaller: &mock.MarshalizerMock{},
	}
	arg.DataBuff, _ = arg.Marshaller.Marshal(createInterceptedHeartbeat())
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
		wasPutPeerIdShardIdCalled := false
		wasPutPeerIdSubTypeCalled := false
		args := createHeartbeatInterceptorProcessArg()
		args.PeerShardMapper = &p2pmocks.NetworkShardingCollectorStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				wasPutPeerIdShardIdCalled = true
			},
			PutPeerIdSubTypeCalled: func(pid core.PeerID, peerSubType core.P2PPeerSubType) {
				wasPutPeerIdSubTypeCalled = true
			},
		}

		paip, err := processor.NewHeartbeatInterceptorProcessor(args)
		assert.Nil(t, err)
		assert.False(t, paip.IsInterfaceNil())
		assert.Equal(t, process.ErrWrongTypeAssertion, paip.Save(providedData, "", ""))
		assert.False(t, wasPutPeerIdShardIdCalled)
		assert.False(t, wasPutPeerIdSubTypeCalled)
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
				ihb := value.(*heartbeatMessages.HeartbeatV2)
				providedHbHandler := providedHb.(interceptedDataHandler)
				providedHbMessage := providedHbHandler.Message().(*heartbeatMessages.HeartbeatV2)
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
		wasPutPeerIdShardIdCalled := false
		wasPutPeerIdSubTypeCalled := false
		arg.PeerShardMapper = &p2pmocks.NetworkShardingCollectorStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				wasPutPeerIdShardIdCalled = true
				assert.Equal(t, providedPid, pid)
			},
			PutPeerIdSubTypeCalled: func(pid core.PeerID, peerSubType core.P2PPeerSubType) {
				wasPutPeerIdSubTypeCalled = true
				assert.Equal(t, providedPid, pid)
			},
		}

		hip, err := processor.NewHeartbeatInterceptorProcessor(arg)
		assert.Nil(t, err)
		assert.False(t, hip.IsInterfaceNil())

		err = hip.Save(providedHb, providedPid, "")
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		assert.True(t, wasPutPeerIdShardIdCalled)
		assert.True(t, wasPutPeerIdSubTypeCalled)
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
