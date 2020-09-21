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
		Type:                    "unknown",
		PeerShardResolver:       &mock.PeerShardResolverStub{},
		Pid:                     "",
		MaxConnectionCount:      5,
		MaxIntraShardValidators: 1,
		MaxCrossShardValidators: 1,
		MaxIntraShardObservers:  1,
		MaxCrossShardObservers:  1,
		MaxFullHistoryObservers: 1,
	}
}

func TestNewSharder_CreateListsSharderShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Type = p2p.ListsSharder
	sharder, err := NewSharder(arg)
	maxPeerCount := 5
	maxValidators := 1
	maxObservers := 1

	expectedSharder, _ := networksharding.NewListsSharder(
		&mock.PeerShardResolverStub{},
		"",
		maxPeerCount,
		maxValidators,
		maxValidators,
		maxObservers,
		maxObservers,
		maxObservers,
	)
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}

func TestNewSharder_CreateOneListSharderShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Type = p2p.OneListSharder
	sharder, err := NewSharder(arg)
	maxPeerCount := 2

	expectedSharder, _ := networksharding.NewOneListSharder("", maxPeerCount)
	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(expectedSharder), reflect.TypeOf(sharder))
}

func TestNewSharder_CreateNilListSharderShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Type = p2p.NilListSharder
	sharder, err := NewSharder(arg)

	expectedSharder := networksharding.NewNilListSharder()
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
