package heartbeat

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	processMocks "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createDefaultInterceptedPeerAuthentication() *heartbeat.PeerAuthentication {
	payload := &heartbeat.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: "hardfork message",
	}
	marshalizer := testscommon.MarshalizerMock{}
	payloadBytes, err := marshalizer.Marshal(payload)
	if err != nil {
		return nil
	}

	return &heartbeat.PeerAuthentication{
		Pubkey:           []byte("public key"),
		Signature:        []byte("signature"),
		Pid:              []byte("peer id"),
		Payload:          payloadBytes,
		PayloadSignature: []byte("payload signature"),
	}
}

func getSizeOfPA(pa *heartbeat.PeerAuthentication) int {
	return len(pa.Pubkey) + len(pa.Pid) +
		len(pa.Signature) + len(pa.Payload) +
		len(pa.PayloadSignature)
}

func createMockInterceptedPeerAuthenticationArg(interceptedData *heartbeat.PeerAuthentication) ArgInterceptedPeerAuthentication {
	arg := ArgInterceptedPeerAuthentication{
		ArgBaseInterceptedHeartbeat: ArgBaseInterceptedHeartbeat{
			Marshalizer: &testscommon.MarshalizerMock{},
		},
		NodesCoordinator:     &shardingMocks.NodesCoordinatorStub{},
		SignaturesHandler:    &processMocks.SignaturesHandlerStub{},
		PeerSignatureHandler: &cryptoMocks.PeerSignatureHandlerStub{},
		ExpiryTimespanInSec:  30,
	}
	arg.DataBuff, _ = arg.Marshalizer.Marshal(interceptedData)

	return arg
}

func TestNewInterceptedPeerAuthentication(t *testing.T) {
	t.Parallel()

	t.Run("nil data buff should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.DataBuff = nil

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ipa)
		assert.Equal(t, process.ErrNilBuffer, err)
	})
	t.Run("nil marshalizer should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.Marshalizer = nil

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ipa)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.NodesCoordinator = nil

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ipa)
		assert.Equal(t, process.ErrNilNodesCoordinator, err)
	})
	t.Run("nil signatures handler should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.SignaturesHandler = nil

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ipa)
		assert.Equal(t, process.ErrNilSignaturesHandler, err)
	})
	t.Run("invalid expiry timespan should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.ExpiryTimespanInSec = 1

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ipa)
		assert.Equal(t, process.ErrInvalidExpiryTimespan, err)
	})
	t.Run("nil peer signature handler should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.PeerSignatureHandler = nil

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ipa)
		assert.Equal(t, process.ErrNilPeerSignatureHandler, err)
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.Marshalizer = &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ipa)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("unmarshalable payload returns error", func(t *testing.T) {
		t.Parallel()

		interceptedData := createDefaultInterceptedPeerAuthentication()
		interceptedData.Payload = []byte("invalid data")
		arg := createMockInterceptedPeerAuthenticationArg(interceptedData)

		ihb, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ihb)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.False(t, ipa.IsInterfaceNil())
		assert.Nil(t, err)
	})
}

func TestInterceptedPeerAuthentication_CheckValidity(t *testing.T) {
	t.Parallel()
	t.Run("publicKeyProperty too short", testInterceptedPeerAuthenticationPropertyLen(publicKeyProperty, false))
	t.Run("publicKeyProperty too short", testInterceptedPeerAuthenticationPropertyLen(publicKeyProperty, true))

	t.Run("signatureProperty too short", testInterceptedPeerAuthenticationPropertyLen(signatureProperty, false))
	t.Run("signatureProperty too short", testInterceptedPeerAuthenticationPropertyLen(signatureProperty, true))

	t.Run("peerIdProperty too short", testInterceptedPeerAuthenticationPropertyLen(peerIdProperty, false))
	t.Run("peerIdProperty too short", testInterceptedPeerAuthenticationPropertyLen(peerIdProperty, true))

	t.Run("payloadProperty too short", testInterceptedPeerAuthenticationPropertyLen(payloadProperty, false))
	t.Run("payloadProperty too short", testInterceptedPeerAuthenticationPropertyLen(payloadProperty, true))

	t.Run("payloadSignatureProperty too short", testInterceptedPeerAuthenticationPropertyLen(payloadSignatureProperty, false))
	t.Run("payloadSignatureProperty too short", testInterceptedPeerAuthenticationPropertyLen(payloadSignatureProperty, true))

	t.Run("nodesCoordinator.GetValidatorWithPublicKey returns error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.NodesCoordinator = &processMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, expectedErr
			},
		}
		ipa, _ := NewInterceptedPeerAuthentication(arg)
		err := ipa.CheckValidity()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("signaturesHandler.Verify returns error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.SignaturesHandler = &processMocks.SignaturesHandlerStub{
			VerifyCalled: func(payload []byte, pid core.PeerID, signature []byte) error {
				return expectedErr
			},
		}
		ipa, _ := NewInterceptedPeerAuthentication(arg)
		err := ipa.CheckValidity()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("peerSignatureHandler.VerifyPeerSignature returns error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.PeerSignatureHandler = &processMocks.PeerSignatureHandlerStub{
			VerifyPeerSignatureCalled: func(pk []byte, pid core.PeerID, signature []byte) error {
				return expectedErr
			},
		}
		ipa, _ := NewInterceptedPeerAuthentication(arg)
		err := ipa.CheckValidity()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("message is expired", func(t *testing.T) {
		t.Parallel()

		marshalizer := testscommon.MarshalizerMock{}
		expiryTimespanInSec := int64(30)
		interceptedData := createDefaultInterceptedPeerAuthentication()
		expiredTimestamp := time.Now().Unix() - expiryTimespanInSec - 1
		payload := &heartbeat.Payload{
			Timestamp: expiredTimestamp,
		}
		payloadBytes, err := marshalizer.Marshal(payload)
		assert.Nil(t, err)

		interceptedData.Payload = payloadBytes
		arg := createMockInterceptedPeerAuthenticationArg(interceptedData)
		arg.Marshalizer = &marshalizer
		arg.ExpiryTimespanInSec = expiryTimespanInSec

		ipa, _ := NewInterceptedPeerAuthentication(arg)

		err = ipa.CheckValidity()
		assert.Equal(t, process.ErrMessageExpired, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		ipa, _ := NewInterceptedPeerAuthentication(arg)
		err := ipa.CheckValidity()
		assert.Nil(t, err)
	})
}

func testInterceptedPeerAuthenticationPropertyLen(property string, tooLong bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		value := []byte("")
		expectedError := process.ErrPropertyTooShort
		if tooLong {
			value = make([]byte, 130)
			expectedError = process.ErrPropertyTooLong
		}

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		ipa, _ := NewInterceptedPeerAuthentication(arg)
		switch property {
		case publicKeyProperty:
			ipa.peerAuthentication.Pubkey = value
		case signatureProperty:
			ipa.peerAuthentication.Signature = value
		case peerIdProperty:
			ipa.peerId = core.PeerID(value)
		case payloadProperty:
			ipa.peerAuthentication.Payload = value
		case payloadSignatureProperty:
			ipa.peerAuthentication.PayloadSignature = value
		default:
			assert.True(t, false)
		}

		err := ipa.CheckValidity()
		assert.True(t, strings.Contains(err.Error(), expectedError.Error()))
	}
}

func TestInterceptedPeerAuthentication_Getters(t *testing.T) {
	t.Parallel()

	providedPA := createDefaultInterceptedPeerAuthentication()
	arg := createMockInterceptedPeerAuthenticationArg(providedPA)
	ipa, _ := NewInterceptedPeerAuthentication(arg)
	expectedPeerAuthentication := &heartbeat.PeerAuthentication{}
	err := arg.Marshalizer.Unmarshal(expectedPeerAuthentication, arg.DataBuff)
	assert.Nil(t, err)
	assert.True(t, ipa.IsForCurrentShard())
	assert.Equal(t, interceptedPeerAuthenticationType, ipa.Type())
	assert.Equal(t, expectedPeerAuthentication.Pid, []byte(ipa.PeerID()))
	assert.Equal(t, expectedPeerAuthentication.Signature, ipa.Signature())
	assert.Equal(t, expectedPeerAuthentication.Payload, ipa.Payload())
	assert.Equal(t, expectedPeerAuthentication.PayloadSignature, ipa.PayloadSignature())
	assert.Equal(t, []byte(""), ipa.Hash())

	identifiers := ipa.Identifiers()
	assert.Equal(t, 2, len(identifiers))
	assert.Equal(t, expectedPeerAuthentication.Pubkey, identifiers[0])
	assert.Equal(t, expectedPeerAuthentication.Pid, identifiers[1])
	providedPASize := getSizeOfPA(providedPA)
	assert.Equal(t, providedPASize, ipa.SizeInBytes())
}
