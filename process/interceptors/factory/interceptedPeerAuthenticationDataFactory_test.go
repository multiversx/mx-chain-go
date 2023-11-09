package factory

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedPeerAuthenticationDataFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil CoreComponents should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.CoreComponents = nil

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(*arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilCoreComponentsHolder, err)
	})
	t.Run("nil InternalMarshalizer should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.IntMarsh = nil
		arg := createMockArgument(coreComp, cryptoComp)

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(*arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.NodesCoordinator = nil

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(*arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilNodesCoordinator, err)
	})
	t.Run("nil SignaturesHandler should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.SignaturesHandler = nil

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(*arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilSignaturesHandler, err)
	})
	t.Run("nil PeerSignatureHandler should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.PeerSignatureHandler = nil

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(*arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrNilPeerSignatureHandler, err)
	})
	t.Run("invalid expiry timespan should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.HeartbeatExpiryTimespanInSec = 1

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(*arg)
		assert.Nil(t, ipadf)
		assert.Equal(t, process.ErrInvalidExpiryTimespan, err)
	})
	t.Run("invalid hardfork pub key should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.HardforkTriggerPubKeyField = make([]byte, 0)
		arg := createMockArgument(coreComp, cryptoComp)

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(*arg)
		assert.Nil(t, ipadf)
		assert.True(t, errors.Is(err, process.ErrInvalidValue))
	})
	t.Run("should work and create", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)

		ipadf, err := NewInterceptedPeerAuthenticationDataFactory(*arg)
		assert.False(t, ipadf.IsInterfaceNil())
		assert.Nil(t, err)

		payload := &heartbeat.Payload{
			Timestamp:       time.Now().Unix(),
			HardforkMessage: "hardfork message",
		}
		marshaller := mock.MarshalizerMock{}
		payloadBytes, err := marshaller.Marshal(payload)
		assert.Nil(t, err)

		peerAuthentication := &heartbeat.PeerAuthentication{
			Pubkey:           []byte("public key"),
			Signature:        []byte("signature"),
			Pid:              []byte("peer id"),
			Payload:          payloadBytes,
			PayloadSignature: []byte("payload signature"),
		}
		marshaledPeerAuthentication, err := marshaller.Marshal(peerAuthentication)
		assert.Nil(t, err)

		interceptedData, err := ipadf.Create(marshaledPeerAuthentication)
		assert.NotNil(t, interceptedData)
		assert.Nil(t, err)
		assert.True(t, strings.Contains(fmt.Sprintf("%T", interceptedData), "*heartbeat.interceptedPeerAuthentication"))
	})
}
