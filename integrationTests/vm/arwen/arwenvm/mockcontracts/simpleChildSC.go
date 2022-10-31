package mockcontracts

import (
	"errors"

	"github.com/ElrondNetwork/wasm-vm/arwen"
	mock "github.com/ElrondNetwork/wasm-vm/mock/context"
	test "github.com/ElrondNetwork/wasm-vm/testcommon"
)

// SimpleCallChildMock is an exposed mock contract method
func SimpleCallChildMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("simpleChildFunction", func() *mock.InstanceMock {
		testConfig := config.(*test.TestConfig)
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		arguments := host.Runtime().Arguments()
		if len(arguments) != 1 {
			host.Runtime().SignalUserError("wrong num of arguments")
			return instance
		}

		host.Metering().UseGas(testConfig.GasUsedByChild)

		behavior := byte(0)
		if len(arguments[0]) != 0 {
			behavior = arguments[0][0]
		}
		err := handleChildBehaviorArgument(host, behavior)
		if err != nil {
			return instance
		}

		host.Storage().SetStorage(test.ChildKey, test.ChildData)
		host.Output().Finish(test.ChildFinish)

		return instance
	})
}

func handleChildBehaviorArgument(host arwen.VMHost, behavior byte) error {
	if behavior == 1 {
		host.Runtime().SignalUserError("child error")
		return errors.New("behavior / child error")
	}

	host.Output().Finish([]byte{behavior})
	return nil
}
