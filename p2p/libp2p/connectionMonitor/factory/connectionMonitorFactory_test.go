package factory

import (
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/connectionMonitor"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestConnectionMonitorFactory_CreateNilShouldCreateNilConnectionMonitor(t *testing.T) {
	t.Parallel()

	cmf := NewConnectionMonitorFactory(nil, 1, 1)

	cm, err := cmf.Create()

	assert.Nil(t, err)
	assert.IsType(t, reflect.TypeOf(&connectionMonitor.NilConnectionMonitor{}), reflect.TypeOf(cm))
}

func TestConnectionMonitorFactory_CreateReconnecterPauseResumeShouldCreateConnectionMonitor(t *testing.T) {
	t.Parallel()

	reconn := &mock.ReconnecterWithPauseAndResumeStub{}
	cmf := NewConnectionMonitorFactory(reconn, 1, 1)

	cm, err := cmf.Create()

	assert.Nil(t, err)
	expectedConnMonitor, _ := connectionMonitor.NewLibp2pConnectionMonitor(reconn, 1, 1)
	assert.IsType(t, reflect.TypeOf(expectedConnMonitor), reflect.TypeOf(cm))
}

func TestConnectionMonitorFactory_CreateReconnecterPauseResumeShouldCreateConnectionMonitorWithIntermediateVar(t *testing.T) {
	t.Parallel()

	reconn := &mock.ReconnecterWithPauseAndResumeStub{}
	simpleReconn := p2p.Reconnecter(reconn)
	cmf := NewConnectionMonitorFactory(simpleReconn, 1, 1)

	cm, err := cmf.Create()

	assert.Nil(t, err)
	expectedConnMonitor, _ := connectionMonitor.NewLibp2pConnectionMonitor(reconn, 1, 1)
	assert.IsType(t, reflect.TypeOf(expectedConnMonitor), reflect.TypeOf(cm))
}

func TestConnectionMonitorFactory_CreateReconnecterShouldCreateConnectionMonitor(t *testing.T) {
	t.Parallel()

	reconn := &mock.ReconnecterStub{}
	cmf := NewConnectionMonitorFactory(reconn, 1, 1)

	cm, err := cmf.Create()

	assert.Nil(t, err)
	expectedConnMonitor, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(reconn, 1)
	assert.IsType(t, reflect.TypeOf(expectedConnMonitor), reflect.TypeOf(cm))
}
