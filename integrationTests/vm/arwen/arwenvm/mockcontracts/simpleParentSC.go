package mockcontracts

import (
	"math/big"

	"github.com/ElrondNetwork/wasm-vm/arwen"
	"github.com/ElrondNetwork/wasm-vm/arwen/elrondapi"
	mock "github.com/ElrondNetwork/wasm-vm/mock/context"
	test "github.com/ElrondNetwork/wasm-vm/testcommon"
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
			host.Runtime().SetRuntimeBreakpointValue(arwen.BreakpointOutOfGas)
			return instance
		}

		host.Storage().SetStorage(test.ParentKeyA, test.ParentDataA)
		host.Output().Finish(test.ParentFinishA)

		retVal := elrondapi.ExecuteOnDestContextWithTypedArgs(
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
