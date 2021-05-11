package factory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInterceptedPeerHeartbeatDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedPeerHeartbeatDataFactory(nil)

	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedPeerHeartbeatDataFactory_NilCoreComponentsShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)
	arg.CoreComponents = nil

	imh, err := NewInterceptedPeerHeartbeatDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilCoreComponentsHolder, err)
}

func TestNewInterceptedPeerHeartbeatDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.IntMarsh = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedPeerHeartbeatDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedPeerHeartbeatDataFactory_NilPeerSignatureHandlerShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)
	arg.PeerSignatureHandler = nil

	imh, err := NewInterceptedPeerHeartbeatDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilPeerSignatureHandler, err)
}

func TestNewInterceptedPeerHeartbeatDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.Hash = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedPeerHeartbeatDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestInterceptedPeerHeartbeatDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedPeerHeartbeatDataFactory(arg)
	assert.False(t, check.IfNil(imh))
	assert.Nil(t, err)

	mockPeerHeartbeat := &heartbeat.PeerHeartbeat{
		Payload:   []byte("payload"),
		Pubkey:    []byte("pk"),
		Signature: []byte("sig"),
		ShardID:   0,
		Pid:       []byte("pid"),
	}
	buff, err := arg.CoreComponents.InternalMarshalizer().Marshal(mockPeerHeartbeat)
	require.Nil(t, err)

	interceptedData, err := imh.Create(buff)
	assert.Nil(t, err)

	assert.True(t, strings.Contains(fmt.Sprintf("%T", interceptedData), "*heartbeat.interceptedPeerHeartbeat"))
}
