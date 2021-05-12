package heartbeat

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultInterceptedData() *heartbeat.PeerHeartbeat {
	return &heartbeat.PeerHeartbeat{
		Payload:   []byte("payload"),
		Pubkey:    []byte("public key"),
		Signature: []byte("signature"),
		ShardID:   3,
		Pid:       []byte("peer id"),
	}
}

func createMockInterceptedPeerHeartbeatArg(interceptedData *heartbeat.PeerHeartbeat) ArgInterceptedPeerHeartbeat {
	arg := ArgInterceptedPeerHeartbeat{
		Marshalizer:          &mock.MarshalizerMock{},
		PeerSignatureHandler: &mock.PeerSignatureHandlerStub{},
		Hasher:               &mock.HasherMock{},
	}
	arg.DataBuff, _ = arg.Marshalizer.Marshal(interceptedData)

	return arg
}

func TestNewInterceptedPeerHeartbeat_EmptyBufferShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockInterceptedPeerHeartbeatArg(createDefaultInterceptedData())
	arg.DataBuff = nil
	iph, err := NewInterceptedPeerHeartbeat(arg)

	assert.Equal(t, process.ErrNilBuffer, err)
	assert.True(t, check.IfNil(iph))
}

func TestNewInterceptedPeerHeartbeat_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockInterceptedPeerHeartbeatArg(createDefaultInterceptedData())
	arg.Marshalizer = nil
	iph, err := NewInterceptedPeerHeartbeat(arg)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.True(t, check.IfNil(iph))
}

func TestNewInterceptedPeerHeartbeat_NilPeerSignatureHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockInterceptedPeerHeartbeatArg(createDefaultInterceptedData())
	arg.PeerSignatureHandler = nil
	iph, err := NewInterceptedPeerHeartbeat(arg)

	assert.Equal(t, process.ErrNilPeerSignatureHandler, err)
	assert.True(t, check.IfNil(iph))
}

func TestNewInterceptedPeerHeartbeat_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockInterceptedPeerHeartbeatArg(createDefaultInterceptedData())
	arg.Hasher = nil
	iph, err := NewInterceptedPeerHeartbeat(arg)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.True(t, check.IfNil(iph))
}

func TestNewInterceptedPeerHeartbeat_NilInvalidDataBuffShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockInterceptedPeerHeartbeatArg(createDefaultInterceptedData())
	arg.DataBuff = []byte("not a valid buffer")
	iph, err := NewInterceptedPeerHeartbeat(arg)

	assert.NotNil(t, err)
	assert.True(t, check.IfNil(iph))
}

func TestNewInterceptedPeerHeartbeat_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockInterceptedPeerHeartbeatArg(createDefaultInterceptedData())
	iph, err := NewInterceptedPeerHeartbeat(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(iph))
}

func TestInterceptedPeerHeartbeat_GettersAndSetters(t *testing.T) {
	t.Parallel()

	pkBytes := []byte("public key")
	pid := []byte("pid")
	hash := []byte("hash")
	interceptedData := createDefaultInterceptedData()
	interceptedData.Pubkey = pkBytes
	interceptedData.Pid = pid
	arg := createMockInterceptedPeerHeartbeatArg(interceptedData)
	arg.Hasher = &mock.HasherStub{
		ComputeCalled: func(s string) []byte {
			return hash
		},
	}
	iph, err := NewInterceptedPeerHeartbeat(arg)
	require.Nil(t, err)
	iph.SetComputedShardID(2)

	assert.Equal(t, pkBytes, iph.PublicKey())
	assert.True(t, iph.IsForCurrentShard())
	assert.Equal(t, interceptedType, iph.Type())
	assert.Equal(t, hash, iph.Hash())
	assert.Equal(t, [][]byte{pkBytes, pid}, iph.Identifiers())
	expectedString := "pk=7075626c6963206b6579, pid=ekxT, sig=7369676e6174757265, payload=7061796c6f6164, received shardID=3, computed shardID=2"
	assert.Equal(t, expectedString, iph.String())
}

func TestInterceptedPeerHeartbeat_CheckValidity(t *testing.T) {
	t.Parallel()

	interceptedData := createDefaultInterceptedData()
	interceptedData.Pubkey = bytes.Repeat([]byte{1}, maxSizeInBytes+1)
	iph, _ := NewInterceptedPeerHeartbeat(createMockInterceptedPeerHeartbeatArg(interceptedData))
	err := iph.CheckValidity()
	assert.True(t, errors.Is(err, process.ErrPropertyTooLong) && strings.Contains(err.Error(), publicKeyProperty))

	interceptedData = createDefaultInterceptedData()
	interceptedData.Pubkey = make([]byte, 0)
	iph, _ = NewInterceptedPeerHeartbeat(createMockInterceptedPeerHeartbeatArg(interceptedData))
	err = iph.CheckValidity()
	assert.True(t, errors.Is(err, process.ErrPropertyTooShort) && strings.Contains(err.Error(), publicKeyProperty))

	interceptedData = createDefaultInterceptedData()
	interceptedData.Signature = bytes.Repeat([]byte{1}, maxSizeInBytes+1)
	iph, _ = NewInterceptedPeerHeartbeat(createMockInterceptedPeerHeartbeatArg(interceptedData))
	err = iph.CheckValidity()
	assert.True(t, errors.Is(err, process.ErrPropertyTooLong) && strings.Contains(err.Error(), signatureProperty))

	interceptedData = createDefaultInterceptedData()
	interceptedData.Payload = bytes.Repeat([]byte{1}, maxSizeInBytes+1)
	iph, _ = NewInterceptedPeerHeartbeat(createMockInterceptedPeerHeartbeatArg(interceptedData))
	err = iph.CheckValidity()
	assert.True(t, errors.Is(err, process.ErrPropertyTooLong) && strings.Contains(err.Error(), payloadProperty))

	interceptedData = createDefaultInterceptedData()
	interceptedData.Pid = bytes.Repeat([]byte{1}, maxSizeInBytes+1)
	iph, _ = NewInterceptedPeerHeartbeat(createMockInterceptedPeerHeartbeatArg(interceptedData))
	err = iph.CheckValidity()
	assert.True(t, errors.Is(err, process.ErrPropertyTooLong) && strings.Contains(err.Error(), peerIdProperty))

	interceptedData = createDefaultInterceptedData()
	iph, _ = NewInterceptedPeerHeartbeat(createMockInterceptedPeerHeartbeatArg(interceptedData))
	err = iph.CheckValidity()
	assert.Nil(t, err)
}
