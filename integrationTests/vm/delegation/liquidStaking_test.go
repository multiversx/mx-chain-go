package delegation

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("liquidStaking")

func TestDelegationSystemSCWithLiquidStaking(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, delegationAddress, tokenID, nonce, round := setupNodesDelegationContractInitLiquidStaking(t)
	_ = logger.SetLogLevel("*:TRACE")
	txData := txDataBuilder.NewBuilder().Clear().
		Func("claimDelegatedPosition").
		Bytes(big.NewInt(1).Bytes()).
		Bytes(delegationAddress).
		Bytes(big.NewInt(5000).Bytes()).
		ToString()
	for _, node := range nodes {
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), vm.LiquidStakingSCAddress, txData, core.MinMetaTxExtraGasCost)
	}

	nrRoundsToPropagateMultiShard := 12
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// claim again
	for _, node := range nodes {
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), vm.LiquidStakingSCAddress, txData, core.MinMetaTxExtraGasCost)
	}

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	for i := 1; i < len(nodes); i++ {
		checkLPPosition(t, nodes[i].OwnAccount.Address, nodes, tokenID, uint64(1), big.NewInt(10000))
	}
	// owner is not allowed to get LP position
	checkLPPosition(t, nodes[0].OwnAccount.Address, nodes, tokenID, uint64(1), big.NewInt(0))
	metaNode := getNodeWithShardID(nodes, core.MetachainShardId)
	allDelegatorAddresses := make([][]byte, 0)
	for i := 1; i < len(nodes); i++ {
		allDelegatorAddresses = append(allDelegatorAddresses, nodes[i].OwnAccount.Address)
	}
	verifyDelegatorIsDeleted(t, metaNode, allDelegatorAddresses, delegationAddress)

	oneTransfer := &vmcommon.ESDTTransfer{
		ESDTValue:      big.NewInt(1000),
		ESDTTokenName:  tokenID,
		ESDTTokenType:  uint32(core.NonFungible),
		ESDTTokenNonce: 1,
	}
	esdtTransfers := []*vmcommon.ESDTTransfer{oneTransfer, oneTransfer, oneTransfer, oneTransfer, oneTransfer}
	txBuilder := txDataBuilder.NewBuilder().MultiTransferESDTNFT(esdtTransfers)
	txBuilder.Bytes([]byte("unDelegatePosition"))
	for _, node := range nodes {
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), vm.LiquidStakingSCAddress, txBuilder.ToString(), core.MinMetaTxExtraGasCost)
	}

	txBuilder = txDataBuilder.NewBuilder().MultiTransferESDTNFT(esdtTransfers)
	txBuilder.Bytes([]byte("returnPosition"))
	for _, node := range nodes {
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), vm.LiquidStakingSCAddress, txBuilder.ToString(), core.MinMetaTxExtraGasCost)
	}
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	for _, node := range nodes {
		checkLPPosition(t, node.OwnAccount.Address, nodes, tokenID, uint64(1), big.NewInt(0))
	}

	verifyDelegatorsStake(t, metaNode, "getUserActiveStake", allDelegatorAddresses, delegationAddress, big.NewInt(5000))
	verifyDelegatorsStake(t, metaNode, "getUserUnStakedValue", allDelegatorAddresses, delegationAddress, big.NewInt(5000))
}

func setupNodesDelegationContractInitLiquidStaking(
	t *testing.T,
) ([]*integrationTests.TestProcessorNode, []int, []byte, []byte, uint64, uint64) {
	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	integrationTests.DisplayAndStartNodes(nodes)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	var tokenID []byte
	for _, node := range nodes {
		node.InitDelegationManager()
		tmpTokenID := node.InitLiquidStaking()
		if len(tmpTokenID) != 0 {
			if len(tokenID) == 0 {
				tokenID = tmpTokenID
			}

			if !bytes.Equal(tokenID, tmpTokenID) {
				log.Error("tokenID missmatch", "current", tmpTokenID, "old", tokenID)
			}
		}
	}

	initialVal := big.NewInt(10000000000)
	initialVal.Mul(initialVal, initialVal)
	integrationTests.MintAllNodes(nodes, initialVal)

	delegationAddress := createNewDelegationSystemSC(nodes[0], nodes)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 6
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData := "delegate"
	for _, node := range nodes {
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(10000), delegationAddress, txData, core.MinMetaTxExtraGasCost)
	}

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	return nodes, idxProposers, delegationAddress, tokenID, nonce, round
}

func checkLPPosition(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenID []byte,
	nonce uint64,
	value *big.Int,
) {
	tokenIdentifierPlusNonce := append(tokenID, big.NewInt(0).SetUint64(nonce).Bytes()...)
	esdtData := esdt.GetESDTTokenData(t, address, nodes, string(tokenIdentifierPlusNonce))

	if value.Cmp(big.NewInt(0)) == 0 {
		require.Nil(t, esdtData.TokenMetaData)
		return
	}

	require.NotNil(t, esdtData.TokenMetaData)
	require.Equal(t, vm.LiquidStakingSCAddress, esdtData.TokenMetaData.Creator)
	require.Equal(t, value.Bytes(), esdtData.Value.Bytes())
}
