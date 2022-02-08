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
	"github.com/stretchr/testify/assert"
)

type interceptedDataSizeHandler interface {
	SizeInBytes() int
}

func createHeartbeatInterceptorProcessArg() processor.ArgHeartbeatInterceptorProcessor {
	return processor.ArgHeartbeatInterceptorProcessor{
		HeartbeatCacher: testscommon.NewCacherStub(),
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
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHb := createMockInterceptedHeartbeat()
		wasCalled := false
		providedPid := core.PeerID("pid")
		arg := createHeartbeatInterceptorProcessArg()
		arg.HeartbeatCacher = &testscommon.CacherStub{
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				assert.True(t, bytes.Equal(providedPid.Bytes(), key))
				ihb := value.(process.InterceptedData)
				assert.True(t, bytes.Equal(providedHb.Identifiers()[0], ihb.Identifiers()[0]))
				ihbSizeHandler := value.(interceptedDataSizeHandler)
				providedHbSizeHandler := providedHb.(interceptedDataSizeHandler)
				assert.Equal(t, providedHbSizeHandler.SizeInBytes(), ihbSizeHandler.SizeInBytes())
				wasCalled = true
				return false
			},
		}
		hip, err := processor.NewHeartbeatInterceptorProcessor(arg)
		assert.Nil(t, err)
		assert.False(t, hip.IsInterfaceNil())

		err = hip.Save(providedHb, providedPid, "")
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestHeartbeatInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	hip, err := processor.NewHeartbeatInterceptorProcessor(createHeartbeatInterceptorProcessArg())
	assert.Nil(t, err)
	assert.False(t, hip.IsInterfaceNil())
	assert.Nil(t, hip.Validate(nil, ""))
	hip.RegisterHandler(nil) // for coverage only, method only logs
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
