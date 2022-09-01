package localFuncs

import (
	"math/big"

	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/arwen/elrondapi"
	mock "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/mock/context"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
)

// LocalMintMock is an exposed mock contract method
func LocalMintMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("localMint", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		scAddress := host.Runtime().GetContextAddress()
		args := host.Runtime().Arguments()

		callData := txDataBuilder.NewBuilder()
		callData.LocalMintESDT(
			string(args[0]),
			big.NewInt(0).SetBytes(args[1]).Int64())

		elrondapi.ExecuteOnDestContextWithTypedArgs(
			host,
			1_000_000,
			big.NewInt(0),
			[]byte(callData.Function()),
			scAddress,
			callData.ElementsAsBytes())

		return instance
	})
}

// LocalBurnMock is an exposed mock contract method
func LocalBurnMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("localBurn", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		scAddress := host.Runtime().GetContextAddress()
		args := host.Runtime().Arguments()

		callData := txDataBuilder.NewBuilder()
		callData.LocalBurnESDT(
			string(args[0]),
			big.NewInt(0).SetBytes(args[1]).Int64())

		elrondapi.ExecuteOnDestContextWithTypedArgs(
			host,
			1_000_000,
			big.NewInt(0),
			[]byte(callData.Function()),
			scAddress,
			callData.ElementsAsBytes())

		return instance
	})
}
