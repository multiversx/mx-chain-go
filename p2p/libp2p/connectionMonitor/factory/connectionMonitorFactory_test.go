package factory

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/connectionMonitor"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockArg() ArgsConnectionMonitorFactory {
	return ArgsConnectionMonitorFactory{
		Reconnecter:                &mock.ReconnecterStub{},
		Sharder:                    &mock.SharderStub{},
		PreferredPeersHolder:       &p2pmocks.PeersHolderStub{},
		ThresholdMinConnectedPeers: 1,
		TargetCount:                1,
	}
}

func TestNewConnectionMonitor_NilReconnecterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Reconnecter = nil
	cm, err := NewConnectionMonitor(arg)

	assert.True(t, check.IfNil(cm))
	assert.Equal(t, p2p.ErrNilReconnecter, err)
}

func TestNewConnectionMonitor_ListSharderWithReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Sharder = &mock.KadSharderStub{}
	cm, err := NewConnectionMonitor(arg)

	assert.False(t, check.IfNil(cm))
	assert.Nil(t, err)
	cmExpected, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(nil, 0, nil, &p2pmocks.PeersHolderStub{})
	//this works even though cmExpected is nil because it checks only the type
	assert.IsType(t, cmExpected, cm)
}

func TestNewConnectionMonitor_InvalidSharderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Sharder = &mock.SharderStub{}
	cm, err := NewConnectionMonitor(arg)

	assert.True(t, check.IfNil(cm))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewConnectionMonitor_NilPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.PreferredPeersHolder = nil
	cm, err := NewConnectionMonitor(arg)

	assert.True(t, check.IfNil(cm))
	assert.True(t, errors.Is(err, p2p.ErrNilPreferredPeersHolder))
}
