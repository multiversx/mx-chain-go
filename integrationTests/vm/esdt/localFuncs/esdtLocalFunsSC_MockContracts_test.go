//go:build !race
// +build !race

package localFuncs

import (
	"testing"

	test "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/testcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm"
)

func TestESDTLocalMintAndBurnFromSC_MockContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 1, 1, 1)
	net.Start().Step()
	defer net.Close()

	net.CreateUninitializedWallets(1)
	ownerWallet := net.CreateWalletOnShard(0, 0)
	node0shard0 := net.NodesSharded[0][0]

	initialVal := uint64(1000000000)
	net.MintWalletsUint64(initialVal)

	scAddress, _ := arwenvm.GetAddressForNewAccountOnWalletAndNode(t, net, ownerWallet, node0shard0)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	arwenvm.InitializeMockContracts(
		t, net,
		test.CreateMockContractOnShard(scAddress, 0).
			WithOwnerAddress(ownerWallet.Address).
			WithConfig(&test.TestConfig{}).
			WithMethods(
				LocalMintMock,
				LocalBurnMock),
	)

	ESDTLocalMintAndBurnFromSC_RunTestsAndAsserts(t, net.Nodes, ownerWallet, scAddress, net.Proposers, nonce, round)
}
