package arwenvm

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/testcommon"
	test "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/testcommon"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm/mockcontracts"
	"github.com/ElrondNetwork/elrond-vm-common/txDataBuilder"
	"github.com/stretchr/testify/require"
)

func TestMockContract_SimpleCall_ChildFail(t *testing.T) {
	net := integrationTests.NewTestNetworkSized(t, 1, 2, 1)
	net.Start()
	net.Step()

	net.CreateUninitializedWallets(2)
	net.Wallets[0] = net.CreateWalletOnShardAndNode(0, 0)
	net.Wallets[1] = net.CreateWalletOnShardAndNode(0, 1)

	net.MintWalletsUint64(100000000000)

	parent := net.Wallets[0]
	child := net.Wallets[1]
	node0 := net.NodesSharded[0][0]
	node1 := net.NodesSharded[0][1]

	parentAddress := GetAddressForNewAccountOnWalletAndNode(t, net, parent, node0)
	childAddress := GetAddressForNewAccountOnWalletAndNode(t, net, child, node1)

	testConfig := &testcommon.TestConfig{
		GasProvided:        1000000,
		GasProvidedToChild: 500000,
		GasUsedByParent:    20000,
		GasUsedByChild:     10000,
		ChildAddress:       childAddress,
	}

	InitializeMockContracts(
		t, net,
		test.CreateMockContract(parentAddress).
			WithBalance(testConfig.ParentBalance).
			WithConfig(testConfig).
			WithMethods(mockcontracts.PerformOnDestCallFailParentMock),
		test.CreateMockContract(childAddress).
			WithBalance(testConfig.ChildBalance).
			WithConfig(testConfig).
			WithMethods(mockcontracts.SimpleCallChildMock),
	)

	_, _ = node0.AccntState.Commit()
	_, _ = node1.AccntState.Commit()

	txData := txDataBuilder.NewBuilder().Func("performOnDestCallFail").ToBytes()
	tx := net.CreateTx(parent, parentAddress, big.NewInt(0), txData)
	tx.GasLimit = testConfig.GasProvided

	txHash := net.SignAndSendTx(parent, tx)

	_ = logger.SetLogLevel("*:TRACE,arwen:TRACE")
	net.Steps(2)

	logHandler, err := node0.TransactionLogProcessor.GetLog([]byte(txHash))
	// node0.TransactionLogProcessor.GetAllCurrentLogs()

	require.Nil(t, err)
	require.NotNil(t, logHandler.GetAddress())
	require.NotNil(t, logHandler.GetLogEvents())
}
