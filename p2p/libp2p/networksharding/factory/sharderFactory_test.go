package factory

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestSharderFactory_CreateWithNilShouldErr(t *testing.T) {
	t.Parallel()

	sf := NewSharderFactory(nil, &mock.PeerShardResolverStub{}, 1, "", 1)

	sharder, err := sf.Create()

	assert.Nil(t, sharder)
	assert.True(t, errors.Is(err, p2p.ErrIncompatibleMethodCalled))
}

func TestSharderFactory_CreateWithReconnecterWithPasueAndResumeShouldWork(t *testing.T) {
	t.Parallel()

	reconn := &mock.ReconnecterWithPauseAndResumeStub{}
	peerResolver := &mock.PeerShardResolverStub{}
	sf := NewSharderFactory(reconn, peerResolver, 1, "", 1)

	sharder, err := sf.Create()

	expectedSharder, _ := networksharding.NewKadSharder(1, peerResolver)
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}

func TestSharderFactory_CreateWithReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	maxPeerCount := 2
	reconn := &mock.ReconnecterStub{}
	peerResolver := &mock.PeerShardResolverStub{}
	sf := NewSharderFactory(reconn, peerResolver, 1, "", maxPeerCount)

	sharder, err := sf.Create()

	expectedSharder, _ := networksharding.NewListKadSharder(peerResolver, "", maxPeerCount, maxPeerCount, maxPeerCount)
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}
