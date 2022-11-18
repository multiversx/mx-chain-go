package arwenvm

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-vm-common/txDataBuilder"
	"github.com/ElrondNetwork/wasm-vm/mock/contracts"
	"github.com/ElrondNetwork/wasm-vm/testcommon"
	test "github.com/ElrondNetwork/wasm-vm/testcommon"
	"github.com/stretchr/testify/require"
)

func Test_ExecuteOnDestCtx_BlockchainHook(t *testing.T) {

	net := integrationTests.NewTestNetworkSized(t, 1, 1, 1)
	net.Start()
	net.Step()

	net.CreateWallets(2)
	net.MintWalletsUint64(100000000000)
	ownerOfParent := net.Wallets[0]

	fakeVMType, _ := hex.DecodeString("abab")

	node := net.NodesSharded[0][0]
	parentAddress, _ := GetAddressForNewAccount(t, net, node)
	childAddress, _ := GetAddressForNewAccountWithVM(t, net, node, fakeVMType)

	testConfig := &testcommon.TestConfig{
		ParentBalance: 20,
		ChildBalance:  10,

		GasProvided:        2_000_000,
		GasProvidedToChild: 1_000_000,
		GasUsedByParent:    400,
	}

	defaultVM, _ := node.VMContainer.Get(factory.ArwenVirtualMachine)
	err := node.VMContainer.Add(fakeVMType, defaultVM)
	require.Nil(t, err)

	InitializeMockContractsWithVMContainer(
		t, net,
		node.VMContainer,
		test.CreateMockContract(parentAddress).
			WithBalance(testConfig.ParentBalance).
			WithConfig(testConfig).
			WithMethods(contracts.ExecOnDestCtxParentMock),
		test.CreateMockContract(childAddress).
			WithBalance(testConfig.ChildBalance).
			WithConfig(testConfig).
			WithMethods(contracts.SimpleChildSetStorageMock),
	)

	txData := txDataBuilder.
		NewBuilder().
		Func("execOnDestCtx").
		Bytes(childAddress).
		Bytes([]byte("simpleChildSetStorage")).
		Bytes(big.NewInt(0).SetUint64(1).Bytes()).
		ToBytes()
	tx := net.CreateTx(ownerOfParent, parentAddress, big.NewInt(0), txData)
	tx.GasLimit = testConfig.GasProvided

	_ = net.SignAndSendTx(ownerOfParent, tx)

	net.Steps(16)

	childHandler, err := net.NodesSharded[0][0].BlockchainHook.GetUserAccount(childAddress)
	require.Nil(t, err)

	childValue, _, err := childHandler.AccountDataHandler().RetrieveValue(test.ChildKey)
	require.Nil(t, err)
	require.Equal(t, test.ChildData, childValue)
}
