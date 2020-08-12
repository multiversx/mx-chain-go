package outport

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/outport/marshaling"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/stretchr/testify/require"
)

func TestNewOutportDriver(t *testing.T) {
	txCoordinator := mock.NewTxCoordinatorMock()
	logsProcessor := mock.NewTxLogsProcessorMock()
	marshalizer := marshaling.CreateMarshalizer(marshaling.JSON)
	sender := NewSender(nil, marshalizer)

	driver, err := newOutportDriver(config.OutportConfig{}, nil, logsProcessor, sender)
	require.Nil(t, driver)
	require.Equal(t, ErrNilTxCoordinator, err)

	driver, err = newOutportDriver(config.OutportConfig{}, txCoordinator, nil, sender)
	require.Nil(t, driver)
	require.Equal(t, ErrNilLogsProcessor, err)

	driver, err = newOutportDriver(config.OutportConfig{}, txCoordinator, logsProcessor, nil)
	require.Nil(t, driver)
	require.Equal(t, ErrNilSender, err)
}

func TestOutportDriver_DigestCommittedBlock(t *testing.T) {

}
