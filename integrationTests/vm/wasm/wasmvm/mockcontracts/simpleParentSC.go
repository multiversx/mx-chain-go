package mockcontracts

import (
	"math/big"

	mock "github.com/multiversx/mx-chain-vm-go/mock/context"
	test "github.com/multiversx/mx-chain-vm-go/testcommon"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/multiversx/mx-chain-vm-go/vmhost/vmhooks"
	"github.com/stretchr/testify/require"
)

var failBehavior = []byte{1}

// PerformOnDestCallFailParentMock is an exposed mock contract method
func PerformOnDestCallFailParentMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("performOnDestCallFail", func() *mock.InstanceMock {
		testConfig := config.(*test.TestConfig)
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		t := instance.T

		err := host.Metering().UseGasBounded(testConfig.GasUsedByParent)
		if err != nil {
			host.Runtime().SetRuntimeBreakpointValue(vmhost.BreakpointOutOfGas)
			return instance
		}

		_, _ = host.Storage().SetStorage(test.ParentKeyA, test.ParentDataA)
		host.Output().Finish(test.ParentFinishA)

		retVal := vmhooks.ExecuteOnDestContextWithTypedArgs(
			host,
			int64(testConfig.GasProvidedToChild),
			big.NewInt(0),
			[]byte("simpleChildFunction"),
			testConfig.ChildAddress,
			[][]byte{failBehavior})
		require.Equal(t, retVal, int32(1))

		return instance
	})
}
