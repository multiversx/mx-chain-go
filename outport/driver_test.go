package outport

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/stretchr/testify/require"
)

func TestNewOutportDriver(t *testing.T) {
	driver, err := NewOutportDriver(config.OutportConfig{}, nil, mock.NewTxLogsProcessorMock())
	require.Nil(t, driver)
	require.Equal(t, ErrNilTxCoordinator, err)

	driver, err = NewOutportDriver(config.OutportConfig{}, mock.NewTxCoordinatorMock(), nil)
	require.Nil(t, driver)
	require.Equal(t, ErrNilLogsProcessor, err)
}

func TestOutportDriver_DigestCommittedBlock(t *testing.T) {

}
