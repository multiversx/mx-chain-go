package validatorToDelegation

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorToDelegationManagerWithNewContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	stakingWalletAccount := integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, nodes[0].ShardCoordinator.SelfId())

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)
	integrationTests.MintAllPlayers(nodes, []*integrationTests.TestWalletAccount{stakingWalletAccount}, initialVal)
	integrationTests.SaveDelegationManagerConfig(nodes)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send stake tx and check sender's balance
	genesisBlock := nodes[0].GenesisBlocks[core.MetachainShardId]
	metaBlock := genesisBlock.(*block.MetaBlock)
	nodePrice := big.NewInt(0).Set(metaBlock.EpochStart.Economics.NodePrice)

	frontendBLSPubkey, err := hex.DecodeString("309befb6387288380edda61ce174b12d42ad161d19361dfcf7e61e6a4e812caf07e45a5a1c5c1e6e1f2f4d84d794dc16d9c9db0088397d85002194b773c30a8b7839324b3b80d9b8510fe53385ba7b7383c96a4c07810db31d84b0feefafbd03")
	require.Nil(t, err)
	frontendHexSignature := "17b1f945404c0c98d2e69a576f3635f4ebe77cd396561566afb969333b0da053e7485b61ef10311f512e3ec2f351ee95"

	nonce, round = generateSendAndWaitToExecuteStakeTransaction(
		t,
		nodes,
		stakingWalletAccount,
		idxProposers,
		nodePrice,
		frontendBLSPubkey,
		frontendHexSignature,
		nonce,
		round,
	)

	time.Sleep(time.Second)

	integrationTests.SaveDelegationContractsList(nodes)

	nonce, round = generateSendAndWaitToExecuteTransaction(
		t,
		nodes,
		stakingWalletAccount,
		idxProposers,
		"makeNewContractFromValidatorData",
		big.NewInt(0),
		[]byte{10},
		nonce,
		round)

	time.Sleep(time.Second)
	scAddressBytes, _ := hex.DecodeString("0000000000000000000100000000000000000000000000000000000002ffffff")
	testBLSKeyOwnerIsAddress(t, nodes, scAddressBytes, frontendBLSPubkey)
}

func TestValidatorToDelegationManagerWithMerge(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	stakingWalletAccount := integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, nodes[0].ShardCoordinator.SelfId())

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)
	integrationTests.MintAllPlayers(nodes, []*integrationTests.TestWalletAccount{stakingWalletAccount}, initialVal)
	integrationTests.SaveDelegationManagerConfig(nodes)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send stake tx and check sender's balance
	var txData string
	genesisBlock := nodes[0].GenesisBlocks[core.MetachainShardId]
	metaBlock := genesisBlock.(*block.MetaBlock)
	nodePrice := big.NewInt(0).Set(metaBlock.EpochStart.Economics.NodePrice)

	frontendBLSPubkey, err := hex.DecodeString("309befb6387288380edda61ce174b12d42ad161d19361dfcf7e61e6a4e812caf07e45a5a1c5c1e6e1f2f4d84d794dc16d9c9db0088397d85002194b773c30a8b7839324b3b80d9b8510fe53385ba7b7383c96a4c07810db31d84b0feefafbd03")
	require.Nil(t, err)
	frontendHexSignature := "17b1f945404c0c98d2e69a576f3635f4ebe77cd396561566afb969333b0da053e7485b61ef10311f512e3ec2f351ee95"

	nonce, round = generateSendAndWaitToExecuteStakeTransaction(
		t,
		nodes,
		stakingWalletAccount,
		idxProposers,
		nodePrice,
		frontendBLSPubkey,
		frontendHexSignature,
		nonce,
		round,
	)

	time.Sleep(time.Second)

	integrationTests.SaveDelegationContractsList(nodes)

	nonce, round = generateSendAndWaitToExecuteTransaction(
		t,
		nodes,
		stakingWalletAccount,
		idxProposers,
		"createNewDelegationContract",
		big.NewInt(10000),
		[]byte{0},
		nonce,
		round,
	)

	scAddressBytes, _ := hex.DecodeString("0000000000000000000100000000000000000000000000000000000002ffffff")
	txData = txDataBuilder.NewBuilder().Clear().
		Func("mergeValidatorDataToContract").
		Bytes(scAddressBytes).
		ToString()
	integrationTests.PlayerSendsTransaction(
		nodes,
		stakingWalletAccount,
		vm.DelegationManagerSCAddress,
		big.NewInt(0),
		txData,
		integrationTests.MinTxGasLimit+uint64(len(txData))+1+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)

	time.Sleep(time.Second)
	testBLSKeyOwnerIsAddress(t, nodes, scAddressBytes, frontendBLSPubkey)
}

func testBLSKeyOwnerIsAddress(t *testing.T, nodes []*integrationTests.TestProcessorNode, address []byte, blsKey []byte) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != core.MetachainShardId {
			continue
		}

		acnt, _ := n.AccntState.GetExistingAccount(vm.StakingSCAddress)
		userAcc, _ := acnt.(state.UserAccountHandler)

		marshaledData, _ := userAcc.DataTrieTracker().RetrieveValue(blsKey)
		stakingData := &systemSmartContracts.StakedDataV2_0{}
		_ = integrationTests.TestMarshalizer.Unmarshal(stakingData, marshaledData)
		assert.Equal(t, stakingData.OwnerAddress, address)
	}
}

func generateSendAndWaitToExecuteStakeTransaction(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	stakingWalletAccount *integrationTests.TestWalletAccount,
	idxProposers []int,
	nodePrice *big.Int,
	frontendBLSPubkey []byte,
	frontendHexSignature string,
	nonce, round uint64,
) (uint64, uint64) {
	oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())
	pubKey := hex.EncodeToString(frontendBLSPubkey)
	txData := "stake" + "@" + oneEncoded + "@" + pubKey + "@" + frontendHexSignature
	integrationTests.PlayerSendsTransaction(
		nodes,
		stakingWalletAccount,
		vm.ValidatorSCAddress,
		nodePrice,
		txData,
		integrationTests.MinTxGasLimit+uint64(len(txData))+1+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 6
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	return nonce, round
}

func generateSendAndWaitToExecuteTransaction(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	stakingWalletAccount *integrationTests.TestWalletAccount,
	idxProposers []int,
	function string,
	value *big.Int,
	serviceFee []byte,
	nonce, round uint64,
) (uint64, uint64) {
	maxDelegationCap := []byte{0}
	txData := txDataBuilder.NewBuilder().Clear().
		Func(function).
		Bytes(maxDelegationCap).
		Bytes(serviceFee).
		ToString()

	integrationTests.PlayerSendsTransaction(
		nodes,
		stakingWalletAccount,
		vm.DelegationManagerSCAddress,
		value,
		txData,
		integrationTests.MinTxGasLimit+uint64(len(txData))+1+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)

	return nonce, round
}
