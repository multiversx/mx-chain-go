package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/require"
)

func TestCreateHostDriver(t *testing.T) {
	t.Parallel()

	args := ArgsHostDriverFactory{
		HostConfig: config.HostDriverConfig{
			URL:                "localhost",
			RetryDurationInSec: 1,
			MarshallerType:     "json",
			Mode:               data.ModeClient,
		},
		Marshaller: &marshallerMock.MarshalizerStub{},
	}

	driver, err := CreateHostDriver(args)
	require.Nil(t, err)
	require.NotNil(t, driver)
	require.Equal(t, "*host.hostDriver", fmt.Sprintf("%T", driver))
}
