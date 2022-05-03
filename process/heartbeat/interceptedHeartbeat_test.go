package heartbeat

import (
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func createDefaultInterceptedHeartbeat() *heartbeat.HeartbeatV2 {
	payload := &heartbeat.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: "hardfork message",
	}
	marshalizer := mock.MarshalizerMock{}
	payloadBytes, err := marshalizer.Marshal(payload)
	if err != nil {
		return nil
	}

	return &heartbeat.HeartbeatV2{
		Payload:         payloadBytes,
		VersionNumber:   "version number",
		NodeDisplayName: "node display name",
		Identity:        "identity",
		Nonce:           123,
		PeerSubType:     uint32(core.RegularPeer),
	}
}

func getSizeOfHeartbeat(hb *heartbeat.HeartbeatV2) int {
	return len(hb.Payload) + len(hb.VersionNumber) +
		len(hb.NodeDisplayName) + len(hb.Identity) +
		uint64Size + uint32Size
}

func createMockInterceptedHeartbeatArg(interceptedData *heartbeat.HeartbeatV2) ArgInterceptedHeartbeat {
	arg := ArgInterceptedHeartbeat{}
	arg.Marshalizer = &mock.MarshalizerMock{}
	arg.DataBuff, _ = arg.Marshalizer.Marshal(interceptedData)
	arg.PeerId = "pid"

	return arg
}

func TestNewInterceptedHeartbeat(t *testing.T) {
	t.Parallel()

	t.Run("nil data buff should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedHeartbeatArg(createDefaultInterceptedHeartbeat())
		arg.DataBuff = nil

		ihb, err := NewInterceptedHeartbeat(arg)
		assert.Nil(t, ihb)
		assert.Equal(t, process.ErrNilBuffer, err)
	})
	t.Run("nil marshalizer should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedHeartbeatArg(createDefaultInterceptedHeartbeat())
		arg.Marshalizer = nil

		ihb, err := NewInterceptedHeartbeat(arg)
		assert.Nil(t, ihb)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("empty pid should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedHeartbeatArg(createDefaultInterceptedHeartbeat())
		arg.PeerId = ""

		ihb, err := NewInterceptedHeartbeat(arg)
		assert.Nil(t, ihb)
		assert.Equal(t, process.ErrEmptyPeerID, err)
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedHeartbeatArg(createDefaultInterceptedHeartbeat())
		arg.Marshalizer = &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}

		ihb, err := NewInterceptedHeartbeat(arg)
		assert.Nil(t, ihb)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("unmarshalable payload returns error", func(t *testing.T) {
		t.Parallel()

		interceptedData := createDefaultInterceptedHeartbeat()
		interceptedData.Payload = []byte("invalid data")
		arg := createMockInterceptedHeartbeatArg(interceptedData)

		ihb, err := NewInterceptedHeartbeat(arg)
		assert.Nil(t, ihb)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedHeartbeatArg(createDefaultInterceptedHeartbeat())

		ihb, err := NewInterceptedHeartbeat(arg)
		assert.False(t, ihb.IsInterfaceNil())
		assert.Nil(t, err)
	})
}

func TestInterceptedHeartbeat_CheckValidity(t *testing.T) {
	t.Parallel()
	t.Run("payloadProperty too short", testInterceptedHeartbeatPropertyLen(payloadProperty, false))
	t.Run("payloadProperty too long", testInterceptedHeartbeatPropertyLen(payloadProperty, true))

	t.Run("versionNumberProperty too short", testInterceptedHeartbeatPropertyLen(versionNumberProperty, false))
	t.Run("versionNumberProperty too long", testInterceptedHeartbeatPropertyLen(versionNumberProperty, true))

	t.Run("nodeDisplayNameProperty too long", testInterceptedHeartbeatPropertyLen(nodeDisplayNameProperty, true))

	t.Run("identityProperty too long", testInterceptedHeartbeatPropertyLen(identityProperty, true))

	t.Run("invalid peer subtype should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedHeartbeatArg(createDefaultInterceptedHeartbeat())
		ihb, _ := NewInterceptedHeartbeat(arg)
		ihb.heartbeat.PeerSubType = 123
		err := ihb.CheckValidity()
		assert.Equal(t, process.ErrInvalidPeerSubType, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockInterceptedHeartbeatArg(createDefaultInterceptedHeartbeat())
		ihb, _ := NewInterceptedHeartbeat(arg)
		err := ihb.CheckValidity()
		assert.Nil(t, err)
	})
}

func testInterceptedHeartbeatPropertyLen(property string, tooLong bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		value := []byte("")
		expectedError := process.ErrPropertyTooShort
		if tooLong {
			value = make([]byte, 130)
			expectedError = process.ErrPropertyTooLong
		}

		arg := createMockInterceptedHeartbeatArg(createDefaultInterceptedHeartbeat())
		ihb, _ := NewInterceptedHeartbeat(arg)
		switch property {
		case payloadProperty:
			ihb.heartbeat.Payload = value
		case versionNumberProperty:
			ihb.heartbeat.VersionNumber = string(value)
		case nodeDisplayNameProperty:
			ihb.heartbeat.NodeDisplayName = string(value)
		case identityProperty:
			ihb.heartbeat.Identity = string(value)
		default:
			assert.True(t, false)
		}

		err := ihb.CheckValidity()
		assert.True(t, strings.Contains(err.Error(), expectedError.Error()))
	}
}

func TestInterceptedHeartbeat_Getters(t *testing.T) {
	t.Parallel()

	providedHB := createDefaultInterceptedHeartbeat()
	arg := createMockInterceptedHeartbeatArg(providedHB)
	ihb, _ := NewInterceptedHeartbeat(arg)
	expectedHeartbeat := &heartbeat.HeartbeatV2{}
	err := arg.Marshalizer.Unmarshal(expectedHeartbeat, arg.DataBuff)
	assert.Nil(t, err)
	assert.True(t, ihb.IsForCurrentShard())
	assert.Equal(t, interceptedHeartbeatType, ihb.Type())
	assert.Equal(t, []byte(""), ihb.Hash())
	assert.Equal(t, arg.PeerId.Bytes(), ihb.Identifiers()[0])
	providedHBSize := getSizeOfHeartbeat(providedHB)
	assert.Equal(t, providedHBSize, ihb.SizeInBytes())
}
