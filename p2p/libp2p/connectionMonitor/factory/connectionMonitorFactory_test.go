package factory

import (
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/connectionMonitor"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArg() ArgsConnectionMonitorFactory {
	return ArgsConnectionMonitorFactory{
		Reconnecter:                &mock.ReconnecterWithPauseAndResumeStub{},
		ThresholdMinConnectedPeers: 1,
		TargetCount:                1,
	}
}

func TestNewConnectionMonitor_NilReconnecterShouldCreateNoConnectionMonitor(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	arg.Reconnecter = nil
	cm, err := NewConnectionMonitor(arg)

	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(&connectionMonitor.NilConnectionMonitor{}), reflect.TypeOf(cm))
}

func TestNewConnectionMonitor_ReconnecterPauseResumeShouldCreateConnectionMonitor(t *testing.T) {
	t.Parallel()

	arg := createMockArg()
	cm, err := NewConnectionMonitor(arg)

	assert.Nil(t, err)
	reconn := &mock.ReconnecterWithPauseAndResumeStub{}
	expectedConnMonitor, _ := connectionMonitor.NewLibp2pConnectionMonitor(reconn, 1, 1)
	assert.IsType(t, reflect.TypeOf(expectedConnMonitor), reflect.TypeOf(cm))
}

func TestNewConnectionMonitor_ReconnecterPauseResumeShouldCreateConnectionMonitorWithIntermediateVar(t *testing.T) {
	t.Parallel()

	reconn := &mock.ReconnecterWithPauseAndResumeStub{}
	simpleReconn := p2p.Reconnecter(reconn)
	arg := createMockArg()
	arg.Reconnecter = simpleReconn
	cm, err := NewConnectionMonitor(arg)

	assert.Nil(t, err)
	expectedConnMonitor, _ := connectionMonitor.NewLibp2pConnectionMonitor(reconn, 1, 1)
	assert.IsType(t, reflect.TypeOf(expectedConnMonitor), reflect.TypeOf(cm))
}

func TestNewConnectionMonitor_ReconnecterShouldCreateConnectionMonitor(t *testing.T) {
	t.Parallel()

	reconn := &mock.ReconnecterStub{}
	arg := createMockArg()
	arg.Reconnecter = reconn
	cm, err := NewConnectionMonitor(arg)

	assert.Nil(t, err)
	expectedConnMonitor, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(reconn, 1)
	assert.IsType(t, reflect.TypeOf(expectedConnMonitor), reflect.TypeOf(cm))
}
