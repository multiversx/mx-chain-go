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

func createMockArg() ArgsSharderFactory {
	return ArgsSharderFactory{
		Reconnecter:        &mock.ReconnecterStub{},
		PeerShardResolver:  &mock.PeerShardResolverStub{},
		PrioBits:           1,
		Pid:                "",
		MaxConnectionCount: 2,
		MaxIntraShard:      1,
		MaxCrossShard:      1,
	}
}

func TestNewSharder_WithNilReconnecterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Reconnecter = nil
	sharder, err := NewSharder(arg)

	assert.Nil(t, sharder)
	assert.True(t, errors.Is(err, p2p.ErrIncompatibleMethodCalled))
}

func TestNewSharder_CreateWithReconnecterWithPauseAndResumeShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Reconnecter = &mock.ReconnecterWithPauseAndResumeStub{}
	sharder, err := NewSharder(arg)

	expectedSharder, _ := networksharding.NewKadSharder(1, &mock.PeerShardResolverStub{})
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}

func TestNewSharder_CreateWithReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	sharder, err := NewSharder(arg)
	maxPeerCount := 2

	expectedSharder, _ := networksharding.NewListKadSharder(&mock.PeerShardResolverStub{}, "", maxPeerCount, maxPeerCount, maxPeerCount)
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}
