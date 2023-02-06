package esdt

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm/wasmvm"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	mock "github.com/multiversx/mx-chain-vm-go/mock/context"
	test "github.com/multiversx/mx-chain-vm-go/testcommon"
	vmhooks "github.com/multiversx/mx-chain-vm-go/vmhost/vmhooks"
)

// MultiTransferViaAsyncMock is an exposed mock contract method
func MultiTransferViaAsyncMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("multi_transfer_via_async", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		testConfig := config.(*test.TestConfig)

		scAddress := host.Runtime().GetContextAddress()

		// destAddress + ESDT transfer tripplets (TokenIdentifier + Nonce + Amount)
		args := host.Runtime().Arguments()
		destAddress := args[0]
		transfers := args[1:]

		callData := txDataBuilder.NewBuilder()
		callData.
			TransferMultiESDT(destAddress, transfers).
			Str("accept_multi_funds_echo")

		value := big.NewInt(testConfig.TransferFromParentToChild).Bytes()
		err := wasmvm.RegisterAsyncCallForMockContract(host, config, scAddress, value, callData)
		if err != nil {
			host.Runtime().SignalUserError(err.Error())
			return instance
		}

		return instance

	})
}

// SyncMultiTransferMock is an exposed mock contract method
func SyncMultiTransferMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("forward_sync_accept_funds_multi_transfer", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		scAddress := host.Runtime().GetContextAddress()

		// destAddress + ESDT transfer tripplets (TokenIdentifier + Nonce + Amount)
		args := host.Runtime().Arguments()
		destAddress := args[0]
		transfers := args[1:]

		callData := txDataBuilder.NewBuilder()
		callData.
			TransferMultiESDT(destAddress, transfers).
			Str("accept_funds_echo")

		vmhooks.ExecuteOnDestContextWithTypedArgs(
			host,
			1_000_000,
			big.NewInt(0),
			[]byte(callData.Function()),
			scAddress,
			callData.ElementsAsBytes())

		return instance
	})
}

// MultiTransferExecuteMock is an exposed mock contract method
func MultiTransferExecuteMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("forward_transf_exec_accept_funds_multi_transfer", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		// destAddress + ESDT transfer tripplets (TokenIdentifier + Nonce + Amount)
		args := host.Runtime().Arguments()
		destAddress := args[0]
		numOfTransfers := (len(args) - 1) / 3

		transfers := make([]*vmcommon.ESDTTransfer, numOfTransfers)
		for i := 0; i < numOfTransfers; i++ {
			tokenStartIndex := 1 + i*parsers.ArgsPerTransfer
			transfer := createEsdtTransferFromArgs(args, tokenStartIndex)
			transfers[i] = transfer
		}

		vmhooks.TransferESDTNFTExecuteWithTypedArgs(
			host,
			destAddress,
			transfers,
			1_000_000,
			[]byte("accept_multi_funds_echo"),
			[][]byte{})

		return instance
	})
}

func createEsdtTransferFromArgs(args [][]byte, transferTripletStartIndex int) *vmcommon.ESDTTransfer {
	transfer := &vmcommon.ESDTTransfer{
		ESDTTokenName:  args[transferTripletStartIndex],
		ESDTTokenNonce: big.NewInt(0).SetBytes(args[transferTripletStartIndex+1]).Uint64(),
		ESDTValue:      big.NewInt(0).SetBytes(args[transferTripletStartIndex+2]),
		ESDTTokenType:  uint32(core.Fungible),
	}
	if transfer.ESDTTokenNonce > 0 {
		transfer.ESDTTokenType = uint32(core.NonFungible)
	}
	return transfer
}

// EmptyCallbackMock is an exposed mock contract method
func EmptyCallbackMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("callBack", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}

// AcceptMultiFundsEchoMock is an exposed mock contract method
func AcceptMultiFundsEchoMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("accept_multi_funds_echo", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}

// AcceptFundsEchoMock is an exposed mock contract method
func AcceptFundsEchoMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("accept_funds_echo", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}

// DoAsyncCallMock is an exposed mock contract method
func DoAsyncCallMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("doAsyncCall", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		args := host.Runtime().Arguments()
		destAddress := args[0]
		egldValue := args[1]
		function := string(args[2])

		callData := txDataBuilder.NewBuilder()
		callData.Func(function)
		for a := 2; a < len(args); a++ {
			callData.Bytes(args[a])
		}

		err := wasmvm.RegisterAsyncCallForMockContract(host, config, destAddress, egldValue, callData)
		if err != nil {
			host.Runtime().SignalUserError(err.Error())
			return instance
		}

		return instance
	})
}
