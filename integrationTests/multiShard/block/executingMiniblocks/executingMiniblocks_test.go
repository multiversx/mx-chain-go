package executingMiniblocks

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldProcessBlocksInMultiShardArchitecture(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	fmt.Println("Setup nodes...")
	numOfShards := 6
	nodesPerShard := 3
	numMetachainNodes := 1

	idxProposers := []int{0, 3, 6, 9, 12, 15, 18}
	senderShard := uint32(0)
	recvShards := []uint32{1, 2}
	round := uint64(0)
	nonce := uint64(0)

	valMinting := big.NewInt(10000000)
	valToTransferPerTx := big.NewInt(2)

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	fmt.Println("Generating private keys for senders and receivers...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	txToGenerateInEachMiniBlock := 3

	proposerNode := nodes[0]

	//sender shard keys, receivers  keys
	sendersPrivateKeys := make([]crypto.PrivateKey, 3)
	receiversPublicKeys := make(map[uint32][]crypto.PublicKey)
	for i := 0; i < txToGenerateInEachMiniBlock; i++ {
		sendersPrivateKeys[i], _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
		//receivers in same shard with the sender
		_, pk, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
		receiversPublicKeys[senderShard] = append(receiversPublicKeys[senderShard], pk)
		//receivers in other shards
		for _, shardId := range recvShards {
			_, pk, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, shardId)
			receiversPublicKeys[shardId] = append(receiversPublicKeys[shardId], pk)
		}
	}

	fmt.Println("Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	fmt.Println("Generating transactions...")
	integrationTests.GenerateAndDisseminateTxs(
		proposerNode,
		sendersPrivateKeys,
		receiversPublicKeys,
		valToTransferPerTx,
		integrationTests.MinTxGasPrice,
		integrationTests.MinTxGasLimit,
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)
	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	}

	gasPricePerTxBigInt := big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice)
	gasLimitPerTxBigInt := big.NewInt(0).SetUint64(integrationTests.MinTxGasLimit)
	gasValue := big.NewInt(0).Mul(gasPricePerTxBigInt, gasLimitPerTxBigInt)
	totalValuePerTx := big.NewInt(0).Add(gasValue, valToTransferPerTx)
	fmt.Println("Test nodes from proposer shard to have the correct balances...")
	for _, n := range nodes {
		isNodeInSenderShard := n.ShardCoordinator.SelfId() == senderShard
		if !isNodeInSenderShard {
			continue
		}

		//test sender balances
		for _, sk := range sendersPrivateKeys {
			valTransferred := big.NewInt(0).Mul(totalValuePerTx, big.NewInt(int64(len(receiversPublicKeys))))
			valRemaining := big.NewInt(0).Sub(valMinting, valTransferred)
			integrationTests.TestPrivateKeyHasBalance(t, n, sk, valRemaining)
		}
		//test receiver balances from same shard
		for _, pk := range receiversPublicKeys[proposerNode.ShardCoordinator.SelfId()] {
			integrationTests.TestPublicKeyHasBalance(t, n, pk, valToTransferPerTx)
		}
	}

	fmt.Println("Test nodes from receiver shards to have the correct balances...")
	for _, n := range nodes {
		isNodeInReceiverShardAndNotProposer := false
		for _, shardId := range recvShards {
			if n.ShardCoordinator.SelfId() == shardId {
				isNodeInReceiverShardAndNotProposer = true
				break
			}
		}
		if !isNodeInReceiverShardAndNotProposer {
			continue
		}

		//test receiver balances from same shard
		for _, pk := range receiversPublicKeys[n.ShardCoordinator.SelfId()] {
			integrationTests.TestPublicKeyHasBalance(t, n, pk, valToTransferPerTx)
		}
	}
}

func TestSimpleTransactionsWithMoreGasWhichYieldInReceiptsInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	minGasLimit := uint64(10000)
	for _, node := range nodes {
		node.EconomicsData.SetMinGasLimit(minGasLimit)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(1000000000)
	sendValue := big.NewInt(5)
	integrationTests.MintAllNodes(nodes, initialVal)
	receiverAddress := []byte("12345678901234567890123456789012")

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	gasLimit := minGasLimit * 2
	time.Sleep(time.Second)
	nrRoundsToTest := 10
	for i := 0; i <= nrRoundsToTest; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		for _, node := range nodes {
			integrationTests.PlayerSendsTransaction(
				nodes,
				node.OwnAccount,
				receiverAddress,
				sendValue,
				"",
				gasLimit,
			)
		}

		time.Sleep(2 * time.Second)
	}

	time.Sleep(time.Second)

	txGasNeed := nodes[0].EconomicsData.GetMinGasLimit()
	txGasPrice := nodes[0].EconomicsData.GetMinGasPrice()

	oneTxCost := big.NewInt(0).Add(sendValue, big.NewInt(0).SetUint64(txGasNeed*txGasPrice))
	txTotalCost := big.NewInt(0).Mul(oneTxCost, big.NewInt(int64(nrRoundsToTest)))

	expectedBalance := big.NewInt(0).Sub(initialVal, txTotalCost)
	for _, verifierNode := range nodes {
		for _, node := range nodes {
			accWrp, err := verifierNode.AccntState.GetExistingAccount(node.OwnAccount.Address)
			if err != nil {
				continue
			}

			account, _ := accWrp.(state.UserAccountHandler)
			assert.Equal(t, expectedBalance, account.GetBalance())
		}
	}
}

func TestSimpleTransactionsWithMoreValueThanBalanceYieldReceiptsInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	minGasLimit := uint64(10000)
	for _, node := range nodes {
		node.EconomicsData.SetMinGasLimit(minGasLimit)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	nrTxsToSend := uint64(10)
	initialVal := big.NewInt(0).SetUint64(nrTxsToSend * minGasLimit * integrationTests.MinTxGasPrice)
	halfInitVal := big.NewInt(0).Div(initialVal, big.NewInt(2))
	integrationTests.MintAllNodes(nodes, initialVal)
	receiverAddress := []byte("12345678901234567890123456789012")

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	for _, node := range nodes {
		for j := uint64(0); j < nrTxsToSend; j++ {
			integrationTests.PlayerSendsTransaction(
				nodes,
				node.OwnAccount,
				receiverAddress,
				halfInitVal,
				"",
				minGasLimit,
			)
		}
	}

	time.Sleep(2 * time.Second)

	integrationTests.UpdateRound(nodes, round)
	integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
	integrationTests.SyncBlock(t, nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == core.MetachainShardId {
			continue
		}

		header := node.BlockChain.GetCurrentBlockHeader()
		shardHdr, ok := header.(*block.Header)
		numInvalid := 0
		require.True(t, ok)
		for _, mb := range shardHdr.MiniBlockHeaders {
			if mb.Type == block.InvalidBlock {
				numInvalid++
			}
		}
		assert.Equal(t, 1, numInvalid)
	}

	time.Sleep(time.Second)
	numRoundsToTest := 6
	for i := 0; i < numRoundsToTest; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)

	expectedReceiverValue := big.NewInt(0).Mul(big.NewInt(int64(len(nodes))), halfInitVal)
	for _, verifierNode := range nodes {
		for _, node := range nodes {
			accWrp, err := verifierNode.AccntState.GetExistingAccount(node.OwnAccount.Address)
			if err != nil {
				continue
			}

			account, _ := accWrp.(state.UserAccountHandler)
			assert.Equal(t, big.NewInt(0), account.GetBalance())
		}

		accWrp, err := verifierNode.AccntState.GetExistingAccount(receiverAddress)
		if err != nil {
			continue
		}

		account, _ := accWrp.(state.UserAccountHandler)
		assert.Equal(t, expectedReceiverValue, account.GetBalance())
	}
}

func TestExecuteBlocksWithGapsBetweenBlocks(t *testing.T) {
	//TODO fix this test
	t.Skip("TODO fix this test")
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodesPerShard := 2
	shardConsensusGroupSize := 2
	nbMetaNodes := 400
	nbShards := 1
	consensusGroupSize := 400

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	cacheMut := &sync.Mutex{}

	putCounter := 0
	cacheMap := make(map[string]interface{})

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorWithCacher(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		shardConsensusGroupSize,
		consensusGroupSize,
		seedAddress,
	)

	roundsPerEpoch := uint64(1000)
	maxGasLimitPerBlock := uint64(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(100)
	for _, nodes := range nodesMap {
		integrationTests.SetEconomicsParameters(nodes, maxGasLimitPerBlock, gasPrice, gasLimit)
		integrationTests.DisplayAndStartNodes(nodes[0:1])

		for _, node := range nodes {
			node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
		}
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	round := uint64(1)
	roundDifference := 10
	nonce := uint64(1)

	firstNodeOnMeta := nodesMap[core.MetachainShardId][0]
	body, header, _ := firstNodeOnMeta.ProposeBlock(round, nonce)

	// set bitmap for all consensus nodes signing
	bitmap := make([]byte, consensusGroupSize/8+1)
	for i := range bitmap {
		bitmap[i] = 0xFF
	}

	bitmap[consensusGroupSize/8] >>= uint8(8 - (consensusGroupSize % 8))
	header.SetPubKeysBitmap(bitmap)

	firstNodeOnMeta.CommitBlock(body, header)

	round += uint64(roundDifference)
	nonce++
	putCounter = 0

	cacheMut.Lock()
	for k := range cacheMap {
		delete(cacheMap, k)
	}
	cacheMut.Unlock()

	firstNodeOnMeta.ProposeBlock(round, nonce)

	assert.Equal(t, roundDifference, putCounter)
}

// TestShouldSubtractTheCorrectTxFee uses the mock VM as it's gas model is predictable
// The test checks the tx fee subtraction from the sender account when deploying a SC
// It also checks the fee obtained by the leader is correct
// Test parameters: 2 shards + meta, each with 2 nodes
func TestShouldSubtractTheCorrectTxFee(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := 2
	consensusGroupSize := 2
	nodesPerShard := 2
	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	// create map of shards - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nodesPerShard,
		maxShards,
		consensusGroupSize,
		consensusGroupSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
		integrationTests.SetEconomicsParameters(nodes, integrationTests.MaxGasLimitPerBlock, integrationTests.MinTxGasPrice, integrationTests.MinTxGasLimit)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(integrationTests.StepDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	initialVal := big.NewInt(10000000)
	senders := integrationTests.CreateSendersWithInitialBalances(nodesMap, initialVal)

	deployValue := big.NewInt(0)
	nodeShard0 := nodesMap[0][0]
	txData := "DEADBEEF@" + hex.EncodeToString(factory.InternalTestingVM) + "@00"
	dummyTx := &transaction.Transaction{
		Data: []byte(txData),
	}
	gasLimit := nodeShard0.EconomicsData.ComputeGasLimit(dummyTx)
	gasLimit += integrationTests.OpGasValueForMockVm
	gasPrice := integrationTests.MinTxGasPrice
	txNonce := uint64(0)
	owner := senders[0][0]
	ownerPk, _ := owner.GeneratePublic().ToByteArray()
	integrationTests.ScCallTxWithParams(
		nodeShard0,
		owner,
		txNonce,
		txData,
		deployValue,
		gasLimit,
		gasPrice,
	)

	_, _, consensusNodes := integrationTests.AllShardsProposeBlock(round, nonce, nodesMap)
	shardId0 := uint32(0)

	_ = integrationTests.IncrementAndPrintRound(round)

	// test sender account decreased its balance with gasPrice * gasLimit
	accnt, err := consensusNodes[shardId0][0].AccntState.GetExistingAccount(ownerPk)
	assert.Nil(t, err)
	ownerAccnt := accnt.(state.UserAccountHandler)
	expectedBalance := big.NewInt(0).Set(initialVal)
	tx := &transaction.Transaction{GasPrice: gasPrice, GasLimit: gasLimit, Data: []byte(txData)}
	txCost := consensusNodes[shardId0][0].EconomicsData.ComputeTxFee(tx)
	expectedBalance.Sub(expectedBalance, txCost)
	assert.Equal(t, expectedBalance, ownerAccnt.GetBalance())

	printContainingTxs(consensusNodes[shardId0][0], consensusNodes[shardId0][0].BlockChain.GetCurrentBlockHeader().(*block.Header))
}

func printContainingTxs(tpn *integrationTests.TestProcessorNode, hdr *block.Header) {
	for _, miniblockHdr := range hdr.MiniBlockHeaders {
		miniblockBytes, err := tpn.Storage.Get(dataRetriever.MiniBlockUnit, miniblockHdr.Hash)
		if err != nil {
			fmt.Println("miniblock " + base64.StdEncoding.EncodeToString(miniblockHdr.Hash) + "not found")
			continue
		}

		miniblock := &block.MiniBlock{}
		err = integrationTests.TestMarshalizer.Unmarshal(miniblock, miniblockBytes)
		if err != nil {
			fmt.Println("can not unmarshal miniblock " + base64.StdEncoding.EncodeToString(miniblockHdr.Hash))
			continue
		}

		for _, txHash := range miniblock.TxHashes {
			txBytes := []byte("not found")

			mbType := miniblockHdr.Type
			switch mbType {
			case block.TxBlock:
				txBytes, err = tpn.Storage.Get(dataRetriever.TransactionUnit, txHash)
				if err != nil {
					fmt.Println("tx hash " + base64.StdEncoding.EncodeToString(txHash) + " not found")
					continue
				}
			case block.SmartContractResultBlock:
				txBytes, err = tpn.Storage.Get(dataRetriever.UnsignedTransactionUnit, txHash)
				if err != nil {
					fmt.Println("scr hash " + base64.StdEncoding.EncodeToString(txHash) + " not found")
					continue
				}
			case block.RewardsBlock:
				txBytes, err = tpn.Storage.Get(dataRetriever.RewardTransactionUnit, txHash)
				if err != nil {
					fmt.Println("reward hash " + base64.StdEncoding.EncodeToString(txHash) + " not found")
					continue
				}
			}

			fmt.Println(string(txBytes))
		}
	}
}
