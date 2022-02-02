package factory

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedPeerAuthenticationDataFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil arg should error", func(t *testing.T) {
		t.Parallel()

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(nil)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilArgumentStruct, err)
	})
	t.Run("nil CoreComponents should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.CoreComponents = nil

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilCoreComponentsHolder, err)
	})
	t.Run("nil InternalMarshalizer should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.IntMarsh = nil
		arg := createMockArgument(coreComp, cryptoComp)

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.NodesCoordinator = nil

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilNodesCoordinator, err)
	})
	t.Run("nil SignaturesHandler should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.SignaturesHandler = nil

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilSignaturesHandler, err)
	})
	t.Run("nil PeerSignatureHandler should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.PeerSignatureHandler = nil

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilPeerSignatureHandler, err)
	})
	t.Run("invalid expiry timespan should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.HeartbeatExpiryTimespanInSec = 1

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrInvalidExpiryTimespan, err)
	})
	t.Run("should work and create", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(arg)
		assert.False(t, ipadf.IsInterfaceNil())
		assert.Nil(t, err)

		payload := &heartbeat.Payload{
			Timestamp:       time.Now().Unix(),
			HardforkMessage: "hardfork message",
		}
		marshalizer := mock.MarshalizerMock{}
		payloadBytes, err := marshalizer.Marshal(payload)
		assert.Nil(t, err)

		peerAuthentication := &heartbeat.PeerAuthentication{
			Pubkey:           []byte("public key"),
			Signature:        []byte("signature"),
			Pid:              []byte("peer id"),
			Payload:          payloadBytes,
			PayloadSignature: []byte("payload signature"),
		}
		marshalizedPAMessage, err := marshalizer.Marshal(peerAuthentication)
		assert.Nil(t, err)

		interceptedData, err := ipadf.Create(marshalizedPAMessage)
		assert.NotNil(t, interceptedData)
		assert.Nil(t, err)
		assert.True(t, strings.Contains(fmt.Sprintf("%T", interceptedData), "*heartbeat.interceptedPeerAuthentication"))
	})
}
