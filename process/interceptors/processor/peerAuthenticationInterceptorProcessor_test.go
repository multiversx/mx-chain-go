package processor

import (
	"bytes"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	heartbeatMessages "github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createPeerAuthenticationInterceptorProcessArg() ArgPeerAuthenticationInterceptorProcessor {
	return ArgPeerAuthenticationInterceptorProcessor{
		PeerAuthenticationCacher: testscommon.NewCacherStub(),
	}
}

func createInterceptedPeerAuthentication() *heartbeatMessages.PeerAuthentication {
	payload := &heartbeatMessages.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: "hardfork message",
	}
	marshalizer := mock.MarshalizerMock{}
	payloadBytes, err := marshalizer.Marshal(payload)
	if err != nil {
		return nil
	}

	return &heartbeatMessages.PeerAuthentication{
		Pubkey:           []byte("public key"),
		Signature:        []byte("signature"),
		Pid:              []byte("peer id"),
		Payload:          payloadBytes,
		PayloadSignature: []byte("payload signature"),
	}
}

func createMockInterceptedPeerAuthentication() *heartbeat.InterceptedPeerAuthentication {
	arg := heartbeat.ArgInterceptedPeerAuthentication{
		ArgBaseInterceptedHeartbeat: heartbeat.ArgBaseInterceptedHeartbeat{
			Marshalizer: &mock.MarshalizerMock{},
		},
		NodesCoordinator:     &mock.NodesCoordinatorStub{},
		SignaturesHandler:    &mock.SignaturesHandlerStub{},
		PeerSignatureHandler: &mock.PeerSignatureHandlerStub{},
		ExpiryTimespanInSec:  30,
	}
	arg.DataBuff, _ = arg.Marshalizer.Marshal(createInterceptedPeerAuthentication())
	ipa, _ := heartbeat.NewInterceptedPeerAuthentication(arg)

	return ipa
}

func TestNewPeerAuthenticationInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil cacher should error", func(t *testing.T) {
		t.Parallel()

		arg := createPeerAuthenticationInterceptorProcessArg()
		arg.PeerAuthenticationCacher = nil
		paip, err := NewPeerAuthenticationInterceptorProcessor(arg)
		assert.Equal(t, process.ErrNilPeerAuthenticationCacher, err)
		assert.Nil(t, paip)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		paip, err := NewPeerAuthenticationInterceptorProcessor(createPeerAuthenticationInterceptorProcessArg())
		assert.Nil(t, err)
		assert.False(t, paip.IsInterfaceNil())
	})
}

func TestPeerAuthenticationInterceptorProcessor_Save(t *testing.T) {
	t.Parallel()

	t.Run("invalid data should error", func(t *testing.T) {
		t.Parallel()

		paip, err := NewPeerAuthenticationInterceptorProcessor(createPeerAuthenticationInterceptorProcessArg())
		assert.Nil(t, err)
		assert.False(t, paip.IsInterfaceNil())
		assert.Equal(t, process.ErrWrongTypeAssertion, paip.Save(nil, "", ""))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedIPA := createMockInterceptedPeerAuthentication()
		wasCalled := false
		providedPid := core.PeerID("pid")
		arg := createPeerAuthenticationInterceptorProcessArg()
		arg.PeerAuthenticationCacher = &testscommon.CacherStub{
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				assert.True(t, bytes.Equal(providedPid.Bytes(), key))
				ipa := value.(*heartbeat.InterceptedPeerAuthentication)
				assert.Equal(t, providedIPA.PeerID(), ipa.PeerID())
				assert.Equal(t, providedIPA.Payload(), ipa.Payload())
				assert.Equal(t, providedIPA.Signature(), ipa.Signature())
				assert.Equal(t, providedIPA.PayloadSignature(), ipa.PayloadSignature())
				assert.Equal(t, providedIPA.SizeInBytes(), ipa.SizeInBytes())
				wasCalled = true
				return false
			},
		}
		paip, err := NewPeerAuthenticationInterceptorProcessor(arg)
		assert.Nil(t, err)
		assert.False(t, paip.IsInterfaceNil())

		err = paip.Save(providedIPA, providedPid, "")
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestPeerAuthenticationInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	paip, err := NewPeerAuthenticationInterceptorProcessor(createPeerAuthenticationInterceptorProcessArg())
	assert.Nil(t, err)
	assert.False(t, paip.IsInterfaceNil())
	assert.Nil(t, paip.Validate(nil, ""))
	paip.RegisterHandler(nil) // for coverage only, method only logs
}
