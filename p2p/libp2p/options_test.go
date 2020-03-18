package libp2p

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
)

func createStubMessengerForDefineOptions(notifeeCalled func(), setStreamHandlerCalled func()) *networkMessenger {
	stubHost := &mock.ConnectableHostStub{
		EventBusCalled:  nil,
		IDCalled:        nil,
		PeerstoreCalled: nil,
		AddrsCalled:     nil,
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{
				NotifyCalled: func(notifiee network.Notifiee) {
					notifeeCalled()
				},
			}
		},
		MuxCalled:     nil,
		ConnectCalled: nil,
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {
			setStreamHandlerCalled()
		},
		SetStreamHandlerMatchCalled: nil,
		RemoveStreamHandlerCalled:   nil,
		NewStreamCalled:             nil,
		CloseCalled:                 nil,
		ConnManagerCalled:           nil,
		ConnectToPeerCalled:         nil,
	}

	mes := &networkMessenger{
		p2pHost: stubHost,
		ctx:     context.Background(),
	}

	return mes
}

func createStubMessengerFailingIfTriggered(t *testing.T) *networkMessenger {
	notifeeCalled := func() {
		assert.Fail(t, "should have not called Notify")
	}
	setStreamHandlerCalled := func() {
		assert.Fail(t, "should have not called SetStreamHandler")
	}

	return createStubMessengerForDefineOptions(notifeeCalled, setStreamHandlerCalled)
}

//------- WithAuthentication

func TestWithAuthentication_NilNetworkShardingCollectorShouldErr(t *testing.T) {
	t.Parallel()

	mes := createStubMessengerFailingIfTriggered(t)
	opt := WithAuthentication(
		nil,
		&mock.SignerVerifierStub{},
		&mock.MarshalizerStub{},
	)

	err := opt(mes)

	assert.Equal(t, p2p.ErrNilNetworkShardingCollector, err)
}

func TestWithAuthentication_NilSignerVerifierShouldErr(t *testing.T) {
	t.Parallel()

	mes := createStubMessengerFailingIfTriggered(t)
	opt := WithAuthentication(
		&mock.NetworkShardingCollectorStub{},
		nil,
		&mock.MarshalizerStub{},
	)

	err := opt(mes)

	assert.Equal(t, p2p.ErrNilSignerVerifier, err)
}

func TestWithAuthentication_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	mes := createStubMessengerFailingIfTriggered(t)
	opt := WithAuthentication(
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{},
		nil,
	)

	err := opt(mes)

	assert.Equal(t, p2p.ErrNilMarshalizer, err)
}

func TestWithAuthentication_ShouldWork(t *testing.T) {
	t.Parallel()

	notifeeCalled := false
	setStreamHandlerCalled := false
	mes := createStubMessengerForDefineOptions(
		func() {
			notifeeCalled = true
		},
		func() {
			setStreamHandlerCalled = true
		},
	)
	opt := WithAuthentication(
		&mock.NetworkShardingCollectorStub{},
		&mock.SignerVerifierStub{},
		&mock.MarshalizerStub{},
	)

	err := opt(mes)

	assert.Nil(t, err)
	assert.NotNil(t, mes.ip)
	assert.True(t, notifeeCalled)
	assert.True(t, setStreamHandlerCalled)
}
