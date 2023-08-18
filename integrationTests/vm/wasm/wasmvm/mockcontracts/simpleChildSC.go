package mockcontracts

import (
	"errors"

	mock "github.com/multiversx/mx-chain-vm-go/mock/context"
	test "github.com/multiversx/mx-chain-vm-go/testcommon"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
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

		_, _ = host.Storage().SetStorage(test.ChildKey, test.ChildData)
		host.Output().Finish(test.ChildFinish)

		return instance
	})
}

func handleChildBehaviorArgument(host vmhost.VMHost, behavior byte) error {
	if behavior == 1 {
		host.Runtime().SignalUserError("child error")
		return errors.New("behavior / child error")
	}

	host.Output().Finish([]byte{behavior})
	return nil
}
