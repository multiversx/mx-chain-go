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

	driver, err := NewOutportDriver(config.OutportConfig{}, nil, logsProcessor, marshalizer)
	require.Nil(t, driver)
	require.Equal(t, ErrNilTxCoordinator, err)

	driver, err = NewOutportDriver(config.OutportConfig{}, txCoordinator, nil, marshalizer)
	require.Nil(t, driver)
	require.Equal(t, ErrNilLogsProcessor, err)

	driver, err = NewOutportDriver(config.OutportConfig{}, txCoordinator, logsProcessor, nil)
	require.Nil(t, driver)
	require.Equal(t, ErrNilMarshalizer, err)
}

func TestOutportDriver_DigestCommittedBlock(t *testing.T) {

}
