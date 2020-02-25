package factory

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArg() ArgsSharderFactory {
	return ArgsSharderFactory{
		Type:               "unknown",
		PeerShardResolver:  &mock.PeerShardResolverStub{},
		PrioBits:           1,
		Pid:                "",
		MaxConnectionCount: 2,
		MaxIntraShard:      1,
		MaxCrossShard:      1,
	}
}

func TestNewSharder_CreateWithPrioBitsVariantShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Type = p2p.SharderVariantPrioBits
	sharder, err := NewSharder(arg)

	expectedSharder, _ := networksharding.NewKadSharder(1, &mock.PeerShardResolverStub{})
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}

func TestNewSharder_CreateWithListSharderShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Type = p2p.SharderVariantWithLists
	sharder, err := NewSharder(arg)
	maxPeerCount := 2

	expectedSharder, _ := networksharding.NewListKadSharder(&mock.PeerShardResolverStub{}, "", maxPeerCount, maxPeerCount, maxPeerCount)
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}

func TestNewSharder_CreateWithListNoSharderShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Type = p2p.NoSharderWithLists
	sharder, err := NewSharder(arg)
	maxPeerCount := 2

	expectedSharder, _ := networksharding.NewListNoKadSharder("", maxPeerCount)
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}

func TestNewSharder_CreateWithDisabledSharderShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Type = p2p.DisabledSharder
	sharder, err := NewSharder(arg)

	expectedSharder := networksharding.NewDisabledSharder()
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}

func TestNewSharder_CreateWithUnknownVariantShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	sharder, err := NewSharder(arg)

	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
	assert.True(t, check.IfNil(sharder))
}
