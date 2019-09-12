package smartContract

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

var gasPrice = 1
var gasLimit = 1000
var initialValueForInternalVariable = uint64(45)

func createScCallsNodes() (p2p.Messenger, []*testNode) {
	advertiser := createMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := createNodes(
		2,
		1,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	return advertiser, nodes
}

func deploySmartContract(t *testing.T, nodeToProcess *testNode, roundNumber uint64, senderAddressBytes []byte, senderNonce uint64) {
	scCode := "aaaa"

	contractTx := createTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		big.NewInt(0),
		scCode+"@"+hex.EncodeToString(factory.InternalTestingVM),
		initialValueForInternalVariable,
	)

	err := nodeToProcess.txProcessor.ProcessTransaction(contractTx, roundNumber)
	assert.Nil(t, err)
	_, err = nodeToProcess.accntState.Commit()
	assert.Nil(t, err)
}

func createTx(
	t *testing.T,
	senderAddressBytes []byte,
	receiverAddressBytes []byte,
	senderNonce uint64,
	value *big.Int,
	scCodeOrFunc string,
	scValue uint64,
) *transaction.Transaction {

	txData := fmt.Sprintf("%s@%X", scCodeOrFunc, scValue)
	tx := &transaction.Transaction{
		Nonce:    senderNonce,
		Value:    value,
		RcvAddr:  receiverAddressBytes,
		SndAddr:  senderAddressBytes,
		GasPrice: uint64(gasPrice),
		GasLimit: uint64(gasLimit),
		Data:     txData,
	}
	assert.NotNil(t, tx)

	return tx
}

func haveTime() time.Duration {
	//fair enough to process a few transactions
	return time.Second * 2
}

// Test within a network of two shards the following situation
// 1. Node in first shard deploys a smart contract -> we also make sure that the resulting smart contract address falls within the same shard
// 2. The same account within the first shard calls the smart contract, we make sure the smart contract is updated and the gas
//  is substracted from the caller's balance
func TestProcessSCCallsInMultiShardArchitecture_FirstShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	generalRoundNumber := uint64(1)
	senderShard := uint32(0)
	senderNonce := uint64(0)
	senderMintingValue := big.NewInt(100000000)

	advertiser, nodes := createScCallsNodes()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.node.Stop()
		}
	}()

	proposerNodeShard1 := nodes[0]

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	senderAddressBytes := []byte("12345678901234567890123456789012")

	// Minting sender account
	createMintingForSenders(nodes, senderShard, [][]byte{senderAddressBytes}, senderMintingValue)

	// should deploy smart contract -> we process a block containing only the sc deployment tx
	deploySmartContract(t, proposerNodeShard1, generalRoundNumber, senderAddressBytes, senderNonce)

	// Test that the gas for deploying the smart contract was substracted from the sender's account
	acc, _ := proposerNodeShard1.node.GetAccount(hex.EncodeToString(senderAddressBytes))
	expectedValue := big.NewInt(0)
	expectedValue.Sub(senderMintingValue, big.NewInt(opGas*1))
	assert.Equal(t, expectedValue, acc.Balance)
	senderNonce++
	assert.Equal(t, senderNonce, acc.Nonce)

	generalRoundNumber++

	// setting the sc deployment address (printed by the transaction processer)
	scDeploymentAdddress, _ := hex.DecodeString("00000000000000000000cca1490e8cd87c767da41cdab632a7a206c5703c3132")

	// Now that the SC is deployed, we test a call from an account located in the first shard
	addValue := uint64(100)
	contractCallTx := createTx(
		t,
		senderAddressBytes,
		scDeploymentAdddress,
		senderNonce,
		big.NewInt(0),
		"Add",
		addValue,
	)

	err := proposerNodeShard1.txProcessor.ProcessTransaction(contractCallTx, generalRoundNumber)
	assert.Nil(t, err)
	_, err = proposerNodeShard1.accntState.Commit()
	assert.Nil(t, err)

	// Test again that the gas for calling the smart contract was substracted from the sender's account
	acc, _ = proposerNodeShard1.node.GetAccount(hex.EncodeToString(senderAddressBytes))
	// Substract the gas price another time
	expectedValue.Sub(expectedValue, big.NewInt(opGas*1))
	assert.Equal(t, expectedValue, acc.Balance)
	senderNonce++
	assert.Equal(t, senderNonce, acc.Nonce)

	generalRoundNumber++
}

// Test within a network of two shards the following situation
// 1. Node in first shard deploys a smart contract -> we also make sure that the resulting smart contract address falls within the same shard
// 2. Another account within the second shard calls the smart contract, we make sure the smart contract is updated and the gas
//  is substracted from the caller's balance at the end
func TestProcessSCCallsInMultiShardArchitecture_FirstShardReceivesCallFromSecondShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	generalRoundNumber := uint64(1)
	senderShard := uint32(0)
	receiverShard := uint32(1)
	senderNonce := uint64(0)
	mintingValue := big.NewInt(100000000)
	receiverNonce := uint64(0)

	advertiser, nodes := createScCallsNodes()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.node.Stop()
		}
	}()

	proposerNodeShard1 := nodes[0]
	proposerNodeShard2 := nodes[1]

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	secondShardAddressBytes := []byte("12345678901234567890123456789011")

	// Minting sender account
	createMintingForSenders(nodes, senderShard, [][]byte{senderAddressBytes}, mintingValue)
	createMintingForSenders(nodes, receiverShard, [][]byte{secondShardAddressBytes}, mintingValue)

	// should deploy smart contract -> we process a block containing only the sc deployment tx
	deploySmartContract(t, proposerNodeShard1, generalRoundNumber, senderAddressBytes, senderNonce)

	// Test that the gas for deploying the smart contract was substracted from the sender's account
	acc, _ := proposerNodeShard1.node.GetAccount(hex.EncodeToString(senderAddressBytes))
	expectedValue := big.NewInt(0)
	expectedValue.Sub(mintingValue, big.NewInt(opGas*1))
	assert.Equal(t, expectedValue, acc.Balance)

	senderNonce++
	assert.Equal(t, senderNonce, acc.Nonce)

	generalRoundNumber++

	// setting the sc deployment address (printed by the transaction processer)
	scDeploymentAdddress, _ := hex.DecodeString("00000000000000000000cca1490e8cd87c767da41cdab632a7a206c5703c3132")

	// Now that the SC is deployed, we test a call from an account located in the second shard
	addValue := uint64(100)
	contractCallTx := createTx(
		t,
		secondShardAddressBytes,
		scDeploymentAdddress,
		receiverNonce,
		big.NewInt(0),
		"Add",
		addValue,
	)

	// The second shard should process this tx as MoveBalance
	err := proposerNodeShard2.txProcessor.ProcessTransaction(contractCallTx, generalRoundNumber)
	assert.Nil(t, err)
	_, err = proposerNodeShard2.accntState.Commit()
	assert.Nil(t, err)

	// Test again that the gas for calling the smart contract was substracted from the sender's account
	acc, _ = proposerNodeShard2.node.GetAccount(hex.EncodeToString(secondShardAddressBytes))
	// TODO: Afrer fees are implemented, from mintingValue we should substract gasLimit + fees until the other shard executes
	//  the smart contract and a refund can be made with the remaining value the following rounds
	assert.Equal(t, mintingValue, acc.Balance)

	receiverNonce++
	assert.Equal(t, receiverNonce, acc.Nonce)

	generalRoundNumber++

	// After second shard processed the transaction, tx should get into the first shard where the SC resides
	err = proposerNodeShard1.txProcessor.ProcessTransaction(contractCallTx, generalRoundNumber)
	assert.Nil(t, err)
	_, err = proposerNodeShard1.accntState.Commit()
	assert.Nil(t, err)

	scAccount, _ := proposerNodeShard1.node.GetAccount(hex.EncodeToString(scDeploymentAdddress))
	storedVal, _ := scAccount.DataTrieTracker().RetrieveValue([]byte("a"))
	storedValBI := big.NewInt(0).SetBytes(storedVal)

	assert.Equal(t, big.NewInt(int64(initialValueForInternalVariable+addValue)), storedValBI)
}

// Test within a network of two shards the following situation
// 1. Node in first shard deploys a smart contract -> we also make sure that the resulting smart contract address falls within the same shard
// 2. Another account within the second shard calls the smart contract, we make sure the smart contract is updated and the gas
//  is substracted from the caller's balance at the end
func TestProcessSCCallsInMultiShardArchitecture_FirstShardReceivesCallFromSecondShardWithSCResults(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	generalRoundNumber := uint64(1)
	scShard := uint32(0)
	accShard := uint32(1)
	accNonce := uint64(0)
	mintingValue := big.NewInt(100000000)
	scNonce := uint64(0)

	advertiser, nodes := createScCallsNodes()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.node.Stop()
		}
	}()

	proposerNodeShardSC := nodes[0]
	proposerNodeShardAccount := nodes[1]

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	scAccountAddressBytes := []byte("12345678901234567890123456789012")
	accountShardAddressBytes := []byte("12345678901234567890123456789011")

	// Minting sender account
	createMintingForSenders(nodes, scShard, [][]byte{scAccountAddressBytes}, mintingValue)
	createMintingForSenders(nodes, accShard, [][]byte{accountShardAddressBytes}, mintingValue)

	// should deploy smart contract -> we process a block containing only the sc deployment tx
	deploySmartContract(t, proposerNodeShardSC, generalRoundNumber, scAccountAddressBytes, accNonce)

	// Test that the gas for deploying the smart contract was substracted from the sender's account
	acc, _ := proposerNodeShardSC.node.GetAccount(hex.EncodeToString(scAccountAddressBytes))
	expectedValue := big.NewInt(0)
	expectedValue.Sub(mintingValue, big.NewInt(opGas*1))
	assert.Equal(t, expectedValue, acc.Balance)
	accNonce++
	assert.Equal(t, accNonce, acc.Nonce)

	generalRoundNumber++

	// setting the sc deployment address (printed by the transaction processer)
	scDeploymentAdddress, _ := hex.DecodeString("00000000000000000000cca1490e8cd87c767da41cdab632a7a206c5703c3132")

	// Update the SC account balance so we can call withdraw function
	createMintingForSenders(nodes, scShard, [][]byte{scDeploymentAdddress}, mintingValue)

	// Now that the SC is deployed, we test a call from an account located in the second shard
	withdrawValue := uint64(100)
	contractCallTx := createTx(
		t,
		accountShardAddressBytes,
		scDeploymentAdddress,
		scNonce,
		big.NewInt(0),
		"Withdraw",
		withdrawValue,
	)

	// The account shard should process this tx as MoveBalance
	scNonce++
	processAndTestSmartContractCallInSender(
		t,
		contractCallTx,
		proposerNodeShardAccount,
		accountShardAddressBytes,
		generalRoundNumber,
		mintingValue,
		scNonce,
	)
	generalRoundNumber++

	// After second shard processed the transaction, tx should get into the first shard where the SC resides
	// We use processBlock here since we want the scResults saved into the storage
	processAndTestSmartContractCallInDestination(
		t,
		contractCallTx,
		proposerNodeShardSC,
		scDeploymentAdddress,
		scShard,
		accShard,
		scNonce,
	)

	// Now the shard containing the SC should have the intermediate results with a sc result that should update the caller's balance
	processAndTestIntermediateResults(
		t,
		proposerNodeShardSC,
		proposerNodeShardAccount,
		accountShardAddressBytes,
		accShard,
		generalRoundNumber,
		mintingValue,
		withdrawValue,
	)
}

func processAndTestSmartContractCallInSender(
	t *testing.T,
	contractCallTx *transaction.Transaction,
	proposerNodeShardAccount *testNode,
	accountShardAddressBytes []byte,
	generalRoundNumber uint64,
	mintingValue *big.Int,
	scNonce uint64,
) {
	err := proposerNodeShardAccount.txProcessor.ProcessTransaction(contractCallTx, generalRoundNumber)
	assert.Nil(t, err)
	_, err = proposerNodeShardAccount.accntState.Commit()
	assert.Nil(t, err)

	// Test again that the gas for calling the smart contract was substracted from the sender's account
	acc, _ := proposerNodeShardAccount.node.GetAccount(hex.EncodeToString(accountShardAddressBytes))
	// TODO: Afrer fees are implemented, from mintingValue we should substract gasLimit + fees until the other shard executes
	//  the smart contract and a refund can be made with the remaining value the following rounds
	assert.Equal(t, mintingValue, acc.Balance)
	assert.Equal(t, scNonce, acc.Nonce)
}

func processAndTestSmartContractCallInDestination(t *testing.T, contractCallTx *transaction.Transaction,
	proposerNodeShardSC *testNode, scDeploymentAdddress []byte, scShard, accShard uint32, scNonce uint64) {
	txBytes, _ := testMarshalizer.Marshal(contractCallTx)
	txHash := testHasher.Compute(string(txBytes))
	blockBody := block.Body{
		&block.MiniBlock{
			TxHashes:        [][]byte{txHash},
			ReceiverShardID: scShard,
			SenderShardID:   accShard,
			Type:            block.TxBlock,
		},
	}
	// Before processing make sure to add the tx into the pool of the scShard
	strCache := process.ShardCacherIdentifier(accShard, scShard)
	proposerNodeShardSC.dPool.Transactions().ShardDataStore(strCache).Put(txHash, contractCallTx)
	proposerNodeShardSC.txCoordinator.RequestBlockTransactions(blockBody)
	_ = proposerNodeShardSC.txCoordinator.ProcessBlockTransaction(blockBody, scNonce, haveTime)
	_ = proposerNodeShardSC.txCoordinator.SaveBlockDataToStorage(blockBody)

	_, err := proposerNodeShardSC.accntState.Commit()
	assert.Nil(t, err)

	scAccount, _ := proposerNodeShardSC.node.GetAccount(hex.EncodeToString(scDeploymentAdddress))
	storedVal, _ := scAccount.DataTrieTracker().RetrieveValue([]byte("a"))
	storedValBI := big.NewInt(0).SetBytes(storedVal)

	assert.Equal(t, big.NewInt(int64(initialValueForInternalVariable)), storedValBI)
}

func processAndTestIntermediateResults(t *testing.T, proposerNodeShardSC *testNode, proposerNodeShardAccount *testNode,
	accountShardAddressBytes []byte, accShard uint32, generalRoundNumber uint64, mintingValue *big.Int, withdrawValue uint64) {
	mbs := proposerNodeShardSC.scrForwarder.CreateAllInterMiniBlocks()
	mb, _ := mbs[accShard]
	assert.NotNil(t, mb)
	// Should have two transactions - one for the withdraw value, the second one for the gas return
	assert.Equal(t, 2, len(mb.TxHashes))

	for _, hash := range mb.TxHashes {
		txBytes, _ := proposerNodeShardSC.store.Get(dataRetriever.UnsignedTransactionUnit, hash)
		assert.NotNil(t, txBytes)

		tx := &transaction.Transaction{}
		_ = testMarshalizer.Unmarshal(tx, txBytes)

		// Now execute transaction back into the account shard
		_ = proposerNodeShardAccount.txProcessor.ProcessTransaction(tx, generalRoundNumber)
		generalRoundNumber++
	}
	_, err := proposerNodeShardAccount.accntState.Commit()
	assert.Nil(t, err)

	// After execution, the first account that started the interaction with the smart contract should have:
	//  - Initial balance + withdraw value - fees
	// TODO: Fees and gas should be taken into consideration when the fees are implemented - now we have extra money
	//  from the gas returned since the gas was not substracted in the first place
	finalValue := big.NewInt(0).Add(mintingValue, big.NewInt(int64(withdrawValue+uint64(gasLimit-1*gasPrice))))
	acc, _ := proposerNodeShardAccount.node.GetAccount(hex.EncodeToString(accountShardAddressBytes))
	assert.Equal(t, finalValue, acc.Balance)
}
