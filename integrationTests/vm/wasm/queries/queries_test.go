package queries

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/stretchr/testify/require"
)

type now struct {
	blockNonce    uint64
	stateRootHash []byte
}

func TestQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	snapshotsOfGetNow := make(map[uint64]now)
	snapshotsOfGetState := make(map[uint64]int)
	historyOfGetNow := make(map[uint64]now)
	historyOfGetState := make(map[uint64]int)

	network := integrationTests.NewMiniNetwork()
	defer network.Stop()

	scOwner := network.AddUser(big.NewInt(10000000000000))

	network.Start()

	// Block 1

	scAddress := deploy(t, network, scOwner.Address, "../testdata/history/output/history.wasm")
	network.Continue(t, 1)

	// Block 2

	now := queryHistoryGetNow(t, network.ShardNode, scAddress, core.OptionalUint64{})
	snapshotsOfGetNow[1] = now
	network.Continue(t, 1)

	// Block 3

	now = queryHistoryGetNow(t, network.ShardNode, scAddress, core.OptionalUint64{})
	snapshotsOfGetNow[2] = now
	setState(t, network, scAddress, scOwner.Address, 42)
	network.Continue(t, 1)

	// Block 4

	state := getState(t, network.ShardNode, scAddress, core.OptionalUint64{})
	snapshotsOfGetState[3] = state
	now = queryHistoryGetNow(t, network.ShardNode, scAddress, core.OptionalUint64{})
	snapshotsOfGetNow[3] = now
	setState(t, network, scAddress, scOwner.Address, 43)
	network.Continue(t, 1)

	// Block 4

	state = getState(t, network.ShardNode, scAddress, core.OptionalUint64{})
	snapshotsOfGetState[4] = state
	now = queryHistoryGetNow(t, network.ShardNode, scAddress, core.OptionalUint64{})
	snapshotsOfGetNow[4] = now
	network.Continue(t, 1)

	// Check snapshots
	block1, _ := network.ShardNode.GetShardHeader(1)
	block2, _ := network.ShardNode.GetShardHeader(2)
	block3, _ := network.ShardNode.GetShardHeader(3)

	require.Equal(t, uint64(1), snapshotsOfGetNow[1].blockNonce)
	require.Equal(t, uint64(2), snapshotsOfGetNow[2].blockNonce)
	require.Equal(t, uint64(3), snapshotsOfGetNow[3].blockNonce)
	require.Equal(t, uint64(4), snapshotsOfGetNow[4].blockNonce)

	require.Equal(t, block1.GetRootHash(), snapshotsOfGetNow[1].stateRootHash)
	require.Equal(t, block1.GetRootHash(), snapshotsOfGetNow[2].stateRootHash)
	require.NotEqual(t, block2.GetRootHash(), snapshotsOfGetNow[3].stateRootHash)
	require.NotEqual(t, block3.GetRootHash(), snapshotsOfGetNow[4].stateRootHash)

	require.Equal(t, 42, snapshotsOfGetState[3])
	require.Equal(t, 43, snapshotsOfGetState[4])

	// Check history
	historyOfGetState[1] = getState(t, network.ShardNode, scAddress, core.OptionalUint64{HasValue: true, Value: 1})
	historyOfGetNow[1] = queryHistoryGetNow(t, network.ShardNode, scAddress, core.OptionalUint64{HasValue: true, Value: 1})

	historyOfGetState[2] = getState(t, network.ShardNode, scAddress, core.OptionalUint64{HasValue: true, Value: 2})
	historyOfGetNow[2] = queryHistoryGetNow(t, network.ShardNode, scAddress, core.OptionalUint64{HasValue: true, Value: 2})

	historyOfGetState[3] = getState(t, network.ShardNode, scAddress, core.OptionalUint64{HasValue: true, Value: 3})
	historyOfGetNow[3] = queryHistoryGetNow(t, network.ShardNode, scAddress, core.OptionalUint64{HasValue: true, Value: 3})

	historyOfGetState[4] = getState(t, network.ShardNode, scAddress, core.OptionalUint64{HasValue: true, Value: 4})
	historyOfGetNow[4] = queryHistoryGetNow(t, network.ShardNode, scAddress, core.OptionalUint64{HasValue: true, Value: 4})

	require.Equal(t, snapshotsOfGetState[1], historyOfGetState[1])
	require.Equal(t, snapshotsOfGetNow[1].blockNonce, historyOfGetNow[1].blockNonce)

	require.Equal(t, snapshotsOfGetState[2], historyOfGetState[2])
	require.Equal(t, snapshotsOfGetNow[2].blockNonce, historyOfGetNow[2].blockNonce)

	require.Equal(t, snapshotsOfGetState[3], historyOfGetState[3])
	require.Equal(t, snapshotsOfGetNow[3].blockNonce, historyOfGetNow[3].blockNonce)

	require.Equal(t, snapshotsOfGetState[4], historyOfGetState[4])
	require.Equal(t, snapshotsOfGetNow[4].blockNonce, historyOfGetNow[4].blockNonce)
}

func deploy(t *testing.T, network *integrationTests.MiniNetwork, sender []byte, codePath string) []byte {
	code := wasm.GetSCCode(codePath)
	data := fmt.Sprintf("%s@%s@0100", code, hex.EncodeToString(factory.WasmVirtualMachine))

	_, err := network.SendTransaction(
		sender,
		make([]byte, 32),
		big.NewInt(0),
		data,
		1000,
	)
	require.NoError(t, err)

	scAddress, _ := network.ShardNode.BlockchainHook.NewAddress(sender, 0, factory.WasmVirtualMachine)
	return scAddress
}

func setState(t *testing.T, network *integrationTests.MiniNetwork, scAddress []byte, sender []byte, value uint64) {
	data := fmt.Sprintf("setState@%x", value)

	_, err := network.SendTransaction(
		sender,
		scAddress,
		big.NewInt(0),
		data,
		1000,
	)

	require.NoError(t, err)
}

func getState(t *testing.T, node *integrationTests.TestProcessorNode, scAddress []byte, blockNonce core.OptionalUint64) int {
	vmOutput, _, err := node.SCQueryService.ExecuteQuery(&process.SCQuery{
		ScAddress:  scAddress,
		FuncName:   "getState",
		Arguments:  [][]byte{},
		BlockNonce: blockNonce,
	})

	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	data := vmOutput.ReturnData

	return int(big.NewInt(0).SetBytes(data[0]).Uint64())
}

func queryHistoryGetNow(t *testing.T, node *integrationTests.TestProcessorNode, scAddress []byte, blockNonce core.OptionalUint64) now {
	vmOutput, _, err := node.SCQueryService.ExecuteQuery(&process.SCQuery{
		ScAddress:  scAddress,
		FuncName:   "getNow",
		Arguments:  [][]byte{},
		BlockNonce: blockNonce,
	})

	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	data := vmOutput.ReturnData

	return now{
		blockNonce:    big.NewInt(0).SetBytes(data[0]).Uint64(),
		stateRootHash: data[1],
	}
}

func TestQueries_Metachain(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	network := integrationTests.NewMiniNetwork()
	defer network.Stop()

	network.Start()

	alice := network.AddUser(big.NewInt(10000000000000))

	// Issue fungible token
	issueCost := big.NewInt(1000)
	tokenNameHex := hex.EncodeToString([]byte("Test"))
	tokenTickerHex := hex.EncodeToString([]byte("TEST"))
	txData := fmt.Sprintf("issue@%s@%s@64@00", tokenNameHex, tokenTickerHex)

	_, err := network.SendTransaction(
		alice.Address,
		vm.ESDTSCAddress,
		issueCost,
		txData,
		core.MinMetaTxExtraGasCost,
	)

	require.NoError(t, err)
	network.Continue(t, 5)

	tokens, err := network.MetachainNode.Node.GetAllIssuedESDTs(core.FungibleESDT, context.Background())
	require.NoError(t, err)
	require.Len(t, tokens, 1)

	// Query token on older block (should fail)
	vmOutput, _, err := network.MetachainNode.SCQueryService.ExecuteQuery(&process.SCQuery{
		ScAddress:  vm.ESDTSCAddress,
		FuncName:   "getTokenProperties",
		Arguments:  [][]byte{[]byte(tokens[0])},
		BlockNonce: core.OptionalUint64{HasValue: true, Value: 2},
	})

	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, vmOutput.ReturnCode)
	require.Equal(t, "no ticker with given name", vmOutput.ReturnMessage)

	// Query token on newer block (should succeed)
	vmOutput, _, err = network.MetachainNode.SCQueryService.ExecuteQuery(&process.SCQuery{
		ScAddress:  vm.ESDTSCAddress,
		FuncName:   "getTokenProperties",
		Arguments:  [][]byte{[]byte(tokens[0])},
		BlockNonce: core.OptionalUint64{HasValue: true, Value: 4},
	})

	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	require.Equal(t, "Test", string(vmOutput.ReturnData[0]))
}
