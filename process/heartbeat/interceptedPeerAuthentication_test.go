package heartbeat

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createDefaultInterceptedPeerAuthentication() *heartbeat.PeerAuthentication {
	return &heartbeat.PeerAuthentication{
		Pubkey:           []byte("public key"),
		Signature:        []byte("signature"),
		Pid:              []byte("peer id"),
		Payload:          []byte("payload"),
		PayloadSignature: []byte("payload signature"),
	}
}

func createMockInterceptedPeerAuthenticationArg(interceptedData *heartbeat.PeerAuthentication) ArgInterceptedPeerAuthentication {
	arg := ArgInterceptedPeerAuthentication{}
	arg.Marshalizer = &mock.MarshalizerMock{}
	arg.Hasher = &hashingMocks.HasherMock{}
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
	t.Run("nil hasher should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
		arg.Hasher = nil

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.Nil(t, ipa)
		assert.Equal(t, process.ErrNilHasher, err)
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
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())

		ipa, err := NewInterceptedPeerAuthentication(arg)
		assert.False(t, ipa.IsInterfaceNil())
		assert.Nil(t, err)
	})
}

func Test_interceptedPeerAuthentication_CheckValidity(t *testing.T) {
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

func Test_interceptedPeerAuthentication_Hash(t *testing.T) {
	t.Parallel()

	arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
	ipa, _ := NewInterceptedPeerAuthentication(arg)
	hash := ipa.Hash()
	expectedHash := arg.Hasher.Compute(string(arg.DataBuff))
	assert.Equal(t, expectedHash, hash)
}

func Test_interceptedPeerAuthentication_Getters(t *testing.T) {
	t.Parallel()

	arg := createMockInterceptedPeerAuthenticationArg(createDefaultInterceptedPeerAuthentication())
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

	identifiers := ipa.Identifiers()
	assert.Equal(t, 2, len(identifiers))
	assert.Equal(t, expectedPeerAuthentication.Pubkey, identifiers[0])
	assert.Equal(t, expectedPeerAuthentication.Pid, identifiers[1])
}
