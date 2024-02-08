//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package upgrades

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
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

	scOwner := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	scOwnerNonce := uint64(0)

	network := integrationTests.NewOneNodeNetwork()
	defer network.Stop()

	network.Mint(scOwner, big.NewInt(10000000000000))
	network.GoToRoundOne()

	// Block 0

	scAddress := deploy(network, scOwner, "../testdata/history/output/history.wasm", &scOwnerNonce)
	network.Continue(t, 1)

	// Block 1

	now := queryHistoryGetNow(t, network.Node, scAddress, core.OptionalUint64{})
	snapshotsOfGetNow[1] = now
	network.Continue(t, 1)

	// Block 2

	now = queryHistoryGetNow(t, network.Node, scAddress, core.OptionalUint64{})
	snapshotsOfGetNow[2] = now
	setState(network, scAddress, scOwner, 42, &scOwnerNonce)
	network.Continue(t, 1)

	// Block 3

	state := getState(t, network.Node, scAddress, core.OptionalUint64{})
	snapshotsOfGetState[3] = state
	now = queryHistoryGetNow(t, network.Node, scAddress, core.OptionalUint64{})
	snapshotsOfGetNow[3] = now
	setState(network, scAddress, scOwner, 43, &scOwnerNonce)
	network.Continue(t, 1)

	// Block 4

	state = getState(t, network.Node, scAddress, core.OptionalUint64{})
	snapshotsOfGetState[4] = state
	now = queryHistoryGetNow(t, network.Node, scAddress, core.OptionalUint64{})
	snapshotsOfGetNow[4] = now
	network.Continue(t, 1)

	// Check snapshots
	block1, _ := network.Node.GetShardHeader(1)
	block2, _ := network.Node.GetShardHeader(2)
	block3, _ := network.Node.GetShardHeader(3)
	block4, _ := network.Node.GetShardHeader(4)

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
	historyOfGetState[1] = getState(t, network.Node, scAddress, core.OptionalUint64{HasValue: true, Value: 1})
	historyOfGetNow[1] = queryHistoryGetNow(t, network.Node, scAddress, core.OptionalUint64{HasValue: true, Value: 1})

	historyOfGetState[2] = getState(t, network.Node, scAddress, core.OptionalUint64{HasValue: true, Value: 2})
	historyOfGetNow[2] = queryHistoryGetNow(t, network.Node, scAddress, core.OptionalUint64{HasValue: true, Value: 2})

	historyOfGetState[3] = getState(t, network.Node, scAddress, core.OptionalUint64{HasValue: true, Value: 3})
	historyOfGetNow[3] = queryHistoryGetNow(t, network.Node, scAddress, core.OptionalUint64{HasValue: true, Value: 3})

	historyOfGetState[4] = getState(t, network.Node, scAddress, core.OptionalUint64{HasValue: true, Value: 4})
	historyOfGetNow[4] = queryHistoryGetNow(t, network.Node, scAddress, core.OptionalUint64{HasValue: true, Value: 4})

	require.Equal(t, snapshotsOfGetState[1], historyOfGetState[1])
	require.Equal(t, snapshotsOfGetNow[1].blockNonce, historyOfGetNow[1].blockNonce)
	// This does not seem right!
	require.Equal(t, block4.GetRootHash(), historyOfGetNow[1].stateRootHash)

	require.Equal(t, snapshotsOfGetState[2], historyOfGetState[2])
	require.Equal(t, snapshotsOfGetNow[2].blockNonce, historyOfGetNow[2].blockNonce)
	// This does not seem right!
	require.Equal(t, block4.GetRootHash(), historyOfGetNow[2].stateRootHash)

	require.Equal(t, snapshotsOfGetState[3], historyOfGetState[3])
	require.Equal(t, snapshotsOfGetNow[3].blockNonce, historyOfGetNow[3].blockNonce)
	// This does not seem right!
	require.Equal(t, block4.GetRootHash(), historyOfGetNow[3].stateRootHash)

	require.Equal(t, snapshotsOfGetState[4], historyOfGetState[4])
	require.Equal(t, snapshotsOfGetNow[4].blockNonce, historyOfGetNow[4].blockNonce)
	// This does not seem right!
	require.Equal(t, block4.GetRootHash(), historyOfGetNow[4].stateRootHash)
}

func deploy(network *integrationTests.OneNodeNetwork, sender []byte, codePath string, accountNonce *uint64) []byte {
	code := wasm.GetSCCode(codePath)
	data := fmt.Sprintf("%s@%s@0100", code, hex.EncodeToString(factory.WasmVirtualMachine))

	network.AddTxToPool(&transaction.Transaction{
		Nonce:    *accountNonce,
		Value:    big.NewInt(0),
		RcvAddr:  vm.CreateEmptyAddress(),
		SndAddr:  sender,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte(data),
	})

	*accountNonce++

	scAddress, _ := network.Node.BlockchainHook.NewAddress(sender, 0, factory.WasmVirtualMachine)

	return scAddress
}

func setState(network *integrationTests.OneNodeNetwork, scAddress, sender []byte, value uint64, accountNonce *uint64) {
	data := fmt.Sprintf("setState@%x", value)

	network.AddTxToPool(&transaction.Transaction{
		Nonce:    *accountNonce,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  sender,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte(data),
	})

	*accountNonce++
}

func getState(t *testing.T, node *integrationTests.TestProcessorNode, scAddress []byte, blockNonce core.OptionalUint64) int {
	scQuery := node.SCQueryService
	vmOutput, _, err := scQuery.ExecuteQuery(&process.SCQuery{
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
	scQuery := node.SCQueryService
	vmOutput, _, err := scQuery.ExecuteQuery(&process.SCQuery{
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
