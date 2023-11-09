package factory

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedHeartbeatDataFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil InternalMarshalizer should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.IntMarsh = nil
		arg := createMockArgument(coreComp, cryptoComp)

		ihdf, err := NewInterceptedHeartbeatDataFactory(*arg)
		assert.Nil(t, ihdf)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("empty peer id should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.PeerID = ""

		ihdf, err := NewInterceptedHeartbeatDataFactory(*arg)
		assert.Nil(t, ihdf)
		assert.Equal(t, process.ErrEmptyPeerID, err)
	})
	t.Run("should work and create", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)

		ihdf, err := NewInterceptedHeartbeatDataFactory(*arg)
		assert.False(t, ihdf.IsInterfaceNil())
		assert.Nil(t, err)

		payload := &heartbeat.Payload{
			Timestamp:       time.Now().Unix(),
			HardforkMessage: "hardfork message",
		}
		marshaller := mock.MarshalizerMock{}
		payloadBytes, err := marshaller.Marshal(payload)
		assert.Nil(t, err)

		hb := &heartbeat.HeartbeatV2{
			Payload:         payloadBytes,
			VersionNumber:   "version number",
			NodeDisplayName: "node display name",
			Identity:        "identity",
			Nonce:           10,
			PeerSubType:     0,
			Pubkey:          []byte("public key"),
		}
		marshaledHeartbeat, err := marshaller.Marshal(hb)
		assert.Nil(t, err)

		interceptedData, err := ihdf.Create(marshaledHeartbeat)
		assert.NotNil(t, interceptedData)
		assert.Nil(t, err)
		assert.True(t, strings.Contains(fmt.Sprintf("%T", interceptedData), "*heartbeat.interceptedHeartbeat"))
	})
}
