package localFuncs

import (
	mock "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/mock/context"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
)

// MultiTransferViaAsyncMock is an exposed mock contract method
func MultiTransferViaAsyncMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("multi_transfer_via_async", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		scAddress := host.Runtime().GetContextAddress()

		// destAddress + ESDT transfer tripplets (TokenIdentifier + Nonce + Amount)
		args := host.Runtime().Arguments()

		callData := txDataBuilder.NewBuilder()
		callData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
		callData.Bytes(args[0])           // destAddress
		callData.Int((len(args) - 1) / 3) // no of triplets
		for a := 1; a < len(args); a++ {
			callData.Bytes(args[a])
		}
		callData.Str("accept_multi_funds_echo")

		err := arwenvm.RegisterAsyncCallForMockContract(host, config, scAddress, callData)
		if err != nil {
			host.Runtime().SignalUserError(err.Error())
			return instance
		}

		return instance

	})
}

// EmptyCallbackMock is an exposed mock contract method
func EmptyCallbackMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("callBack", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}

// AcceptMultiFundsEcho is an exposed mock contract method
func AcceptMultiFundsEcho(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("accept_multi_funds_echo", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}

// SyncAcceptMultiFundsEcho is an exposed mock contract method
func SyncAcceptMultiFundsEcho(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("forward_sync_accept_funds_multi_transfer", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}
