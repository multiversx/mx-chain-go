package libp2p_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
)

var marshalizerOutput = []byte("marshalizer byte output")

func createStubHostForIdentityProvider() (*mock.ConnectableHostStub, network.Stream) {
	newStream := mock.NewStreamMock()

	return &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
		NewStreamCalled: func(ctx context.Context, p peer.ID, pids ...protocol.ID) (stream network.Stream, e error) {
			return newStream, nil
		},
		IDCalled: func() peer.ID {
			return "stub ID"
		},
	}, newStream
}

func createStubConnForIdentityProvider() network.Conn {
	return &mock.ConnStub{
		RemotePeerCalled: func() peer.ID {
			return "remote ID"
		},
	}
}

func createStubMarshalizerForIdentityProvider() p2p.Marshalizer {
	return &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
			return marshalizerOutput, nil
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	}
}

//------- NewIdentityProvider

func TestNewIdentityProvider_NilHostShouldErr(t *testing.T) {
	t.Parallel()

	ip, err := libp2p.NewIdentityProvider(
		nil,
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{},
		&mock.MarshalizerStub{},
		time.Second,
	)

	assert.Nil(t, ip)
	assert.Equal(t, p2p.ErrNilHost, err)
}

func TestNewIdentityProvider_NilShardingCollectorStubShouldErr(t *testing.T) {
	t.Parallel()

	ip, err := libp2p.NewIdentityProvider(
		&mock.ConnectableHostStub{},
		nil,
		&mock.SignerVerifierStub{},
		&mock.MarshalizerStub{},
		time.Second,
	)

	assert.Nil(t, ip)
	assert.Equal(t, p2p.ErrNilNetworkShardingCollector, err)
}

func TestNewIdentityProvider_NilSignerVerifierShouldErr(t *testing.T) {
	t.Parallel()

	ip, err := libp2p.NewIdentityProvider(
		&mock.ConnectableHostStub{},
		&mock.NetworkShardingCollectorStub{},
		nil,
		&mock.MarshalizerStub{},
		time.Second,
	)

	assert.Nil(t, ip)
	assert.Equal(t, p2p.ErrNilSignerVerifier, err)
}

func TestNewIdentityProvider_NilMarshalizerErr(t *testing.T) {
	t.Parallel()

	ip, err := libp2p.NewIdentityProvider(
		&mock.ConnectableHostStub{},
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{},
		nil,
		time.Second,
	)

	assert.Nil(t, ip)
	assert.Equal(t, p2p.ErrNilMarshalizer, err)
}

func TestNewIdentityProvider_ShouldWorkAndSetStreamHandler(t *testing.T) {
	t.Parallel()

	setStreamHandlerCalled := false
	ip, err := libp2p.NewIdentityProvider(
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {
				setStreamHandlerCalled = true
			},
		},
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{},
		&mock.MarshalizerStub{},
		time.Second,
	)

	assert.NotNil(t, ip)
	assert.Nil(t, err)
	assert.True(t, setStreamHandlerCalled)
}

//------- Connected

func TestIdentityProvider_ConnectedMarshalizerFailShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not fail: %v", r))
		}
	}()

	host, _ := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{
			PublicKeyCalled: func() []byte {
				return []byte("pub key")
			},
		},
		&mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
				return nil, errors.New("marshalizer error")
			},
		},
		time.Second,
	)

	ip.Connected(nil, createStubConnForIdentityProvider())

	time.Sleep(time.Millisecond * 100)
}

func TestIdentityProvider_ConnectedSignFailShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not fail: %v", r))
		}
	}()

	host, _ := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{
			PublicKeyCalled: func() []byte {
				return []byte("pub key")
			},
			SignCalled: func(message []byte) (bytes []byte, e error) {
				return nil, errors.New("signing failed")
			},
		},
		createStubMarshalizerForIdentityProvider(),
		time.Second,
	)

	ip.Connected(nil, createStubConnForIdentityProvider())

	time.Sleep(time.Millisecond * 100)
}

func TestIdentityProvider_ConnectedShouldWrite(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not fail: %v", r))
		}
	}()

	host, stream := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{
			PublicKeyCalled: func() []byte {
				return []byte("pub key")
			},
			SignCalled: func(message []byte) (bytes []byte, e error) {
				return []byte("signature"), nil
			},
		},
		createStubMarshalizerForIdentityProvider(),
		time.Second,
	)

	ip.Connected(nil, createStubConnForIdentityProvider())

	time.Sleep(time.Millisecond * 100)

	recovered := make([]byte, len(marshalizerOutput))
	_, _ = stream.Read(recovered)
	assert.Equal(t, marshalizerOutput, recovered)
}

//------- processReceivedData

func TestIdentityProvider_ProcessReceivedDataUnmarshalFailsShouldError(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	host, _ := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{},
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errExpected
			},
		},
		time.Second,
	)

	err := ip.ProcessReceivedData(make([]byte, 0))

	assert.Equal(t, errExpected, err)
}

func TestIdentityProvider_ProcessReceivedDataMarshalFailsShouldError(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	host, _ := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{},
		&mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
				return make([]byte, 0), errExpected
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
		},
		time.Second,
	)

	err := ip.ProcessReceivedData(make([]byte, 0))

	assert.Equal(t, errExpected, err)
}

func TestIdentityProvider_ProcessReceivedDataSignerErrorsShouldError(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	host, _ := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{
			VerifyCalled: func(message []byte, sig []byte, pk []byte) error {
				return errExpected
			},
		},
		createStubMarshalizerForIdentityProvider(),
		time.Second,
	)

	err := ip.ProcessReceivedData(make([]byte, 0))

	assert.Equal(t, errExpected, err)
}

func TestIdentityProvider_ProcessReceivedDataShouldUpdateCollector(t *testing.T) {
	t.Parallel()

	updateWasCalled := false
	host, _ := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{
			UpdatePeerIdPublicKeyCalled: func(pid core.PeerID, pk []byte) {
				updateWasCalled = true
			},
		},
		&mock.SignerVerifierStub{
			VerifyCalled: func(message []byte, sig []byte, pk []byte) error {
				return nil
			},
		},
		createStubMarshalizerForIdentityProvider(),
		time.Second,
	)

	err := ip.ProcessReceivedData(make([]byte, 0))

	assert.Nil(t, err)
	assert.True(t, updateWasCalled)
}

//------- handleStreams

func TestIdentityProvider_HandleStreamsReceivedMessageShouldUpdateCollector(t *testing.T) {
	t.Parallel()

	updateWasCalled := false
	host, stream := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{
			UpdatePeerIdPublicKeyCalled: func(pid core.PeerID, pk []byte) {
				updateWasCalled = true
			},
		},
		&mock.SignerVerifierStub{
			VerifyCalled: func(message []byte, sig []byte, pk []byte) error {
				return nil
			},
		},
		createStubMarshalizerForIdentityProvider(),
		time.Second,
	)
	_, _ = stream.Write([]byte("mock data"))

	ip.HandleStreams(stream)

	assert.True(t, updateWasCalled)
}

func TestIdentityProvider_HandleStreamsTimeoutShouldNotFailOrWrite(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not fail: %v", r))
		}
	}()

	updateWasCalled := false
	host, stream := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{
			UpdatePeerIdPublicKeyCalled: func(pid core.PeerID, pk []byte) {
				updateWasCalled = true
			},
		},
		&mock.SignerVerifierStub{
			VerifyCalled: func(message []byte, sig []byte, pk []byte) error {
				return nil
			},
		},
		createStubMarshalizerForIdentityProvider(),
		time.Second,
	)

	ip.HandleStreams(stream)

	assert.False(t, updateWasCalled)
}

func TestIdentityProvider_HandleStreamsClosedStreamShouldNotFailOrWrite(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not fail: %v", r))
		}
	}()

	updateWasCalled := false
	host, stream := createStubHostForIdentityProvider()
	ip, _ := libp2p.NewIdentityProvider(
		host,
		&mock.NetworkShardingCollectorStub{
			UpdatePeerIdPublicKeyCalled: func(pid core.PeerID, pk []byte) {
				updateWasCalled = true
			},
		},
		&mock.SignerVerifierStub{
			VerifyCalled: func(message []byte, sig []byte, pk []byte) error {
				return nil
			},
		},
		createStubMarshalizerForIdentityProvider(),
		time.Second,
	)
	_ = stream.Close()

	ip.HandleStreams(stream)

	assert.False(t, updateWasCalled)
}
