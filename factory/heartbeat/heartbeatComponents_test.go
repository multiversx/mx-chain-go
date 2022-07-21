package heartbeat_test

import (
	"testing"

	heartbeatComp "github.com/ElrondNetwork/elrond-go/factory/heartbeat"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

// ------------ Test HeartbeatComponents --------------------
func TestHeartbeatComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	heartbeatArgs := componentsMock.GetHeartbeatFactoryArgs(shardCoordinator)
	hcf, err := heartbeatComp.NewHeartbeatComponentsFactory(heartbeatArgs)
	require.Nil(t, err)
	cc, err := hcf.Create()
	require.Nil(t, err)

	err = cc.Close()
	require.NoError(t, err)
}
