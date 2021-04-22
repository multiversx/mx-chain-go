package integrationTests

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/require"
)

// ShardIdentifier is the numeric index of a shard
type ShardIdentifier = uint32

// Address is a slice of bytes used to identify an account
type Address = []byte

// NodeSlice is a slice of TestProcessorNode instances
type NodeSlice = []*TestProcessorNode

// NodesByShardMap is a map that groups TestProcessorNodes by their shard ID
type NodesByShardMap = map[ShardIdentifier]NodeSlice

// GasScheduleMap is a map containing the predefined gas costs
type GasScheduleMap = map[string]map[string]uint64

// TestNetwork wraps a set of TestProcessorNodes along with a set of test
// Wallets, instantiates them, controls them and provides operations with them;
// designed to be used in integration tests.
// TODO combine TestNetwork with the preexisting TestContext and OneNodeNetwork
// into a single struct containing the functionality of all three
type TestNetwork struct {
	NumShards          int
	NodesPerShard      int
	NodesInMetashard   int
	Nodes              NodeSlice
	NodesSharded       NodesByShardMap
	Wallets            []*TestWalletAccount
	DeploymentAddress  Address
	Proposers          []int
	Round              uint64
	Nonce              uint64
	T                  *testing.T
	DefaultNode        *TestProcessorNode
	DefaultGasPrice    uint64
	MinGasLimit        uint64
	MaxGasLimit        uint64
	DefaultVM          []byte
	DefaultGasSchedule GasScheduleMap
	BypassErrorsOnce   bool
}

// NewTestNetwork creates an unsized TestNetwork; topology must be configured
// afterwards, before starting.
func NewTestNetwork(t *testing.T) *TestNetwork {
	// TODO replace testing.T with testing.TB everywhere in integrationTest
	return &TestNetwork{
		T:                t,
		BypassErrorsOnce: false,
	}
}

// NewTestNetworkSized creates a new TestNetwork containing topology
// information; can be started immediately.
func NewTestNetworkSized(
	t *testing.T,
	numShards int,
	nodesPerShard int,
	nodesInMetashard int,
) *TestNetwork {
	net := NewTestNetwork(t)
	net.NumShards = numShards
	net.NodesPerShard = nodesPerShard
	net.NodesInMetashard = nodesInMetashard

	return net
}

// Start initializes the test network and starts its nodes
func (net *TestNetwork) Start() {
	net.Round = 0
	net.Nonce = 0

	net.createNodes()
	net.indexProposers()
	net.startNodes()
	net.mapNodesByShard()
	net.initDefaults()
}

// Increment only increments the Round and the Nonce, without triggering the
// processing of a block; use Step to process a block as well.
func (net *TestNetwork) Increment() {
	net.Round = IncrementAndPrintRound(net.Round)
	net.Nonce++
}

// Step increments the Round and Nonce and triggers the production and
// synchronization of a single block.
func (net *TestNetwork) Step() {
	net.Round, net.Nonce = ProposeAndSyncOneBlock(
		net.T,
		net.Nodes,
		net.Proposers,
		net.Round,
		net.Nonce)
}

// Steps repeatedly increments the Round and Nonce and processes blocks.
func (net *TestNetwork) Steps(steps int) {
	net.Nonce, net.Round = WaitOperationToBeDone(
		net.T,
		net.Nodes,
		steps,
		net.Nonce,
		net.Round,
		net.Proposers)
}

// Close shuts down the test network.
func (net *TestNetwork) Close() {
	net.closeNodes()
}

// MintNodeAccounts adds the specified value to the accounts owned by the nodes
// of the TestNetwork.
func (net *TestNetwork) MintNodeAccounts(value *big.Int) {
	MintAllNodes(net.Nodes, value)
}

// MintNodeAccountsUint64 adds the specified value to the accounts owned by the
// nodes of the TestNetwork.
func (net *TestNetwork) MintNodeAccountsUint64(value uint64) {
	MintAllNodes(net.Nodes, big.NewInt(0).SetUint64(value))
}

// CreateWallets initializes the internal test wallets
func (net *TestNetwork) CreateWallets(count int) {
	net.Wallets = make([]*TestWalletAccount, count)

	for i := 0; i < count; i++ {
		shardID := ShardIdentifier(i % net.NumShards)
		node := net.firstNodeInShard(shardID)
		net.Wallets[i] = CreateTestWalletAccount(node.ShardCoordinator, shardID)
	}
}

// MintWallets adds the specified value to the test wallets.
func (net *TestNetwork) MintWallets(value *big.Int) {
	// TODO rename Players to Wallets where this function is defined.
	MintAllPlayers(net.Nodes, net.Wallets, value)
}

// MintWalletsUint64 adds the specified value to the test wallets.
func (net *TestNetwork) MintWalletsUint64(value uint64) {
	// TODO rename Players to Wallets where this function is defined.
	MintAllPlayers(net.Nodes, net.Wallets, big.NewInt(0).SetUint64(value))
}

// SendTx submits the provided transaction to the test network; transaction
// must be already signed.
func (net *TestNetwork) SendTx(tx *transaction.Transaction) string {
	node := net.firstNodeInShardOfAddress(tx.SndAddr)
	return net.SendTxFromNode(tx, node)
}

// SignAndSendTx signs then submits the provided transaction to the test
// network, using the provided signer account; use it to send transactions that
// have been modified after being created.
func (net *TestNetwork) SignAndSendTx(signer *TestWalletAccount, tx *transaction.Transaction) string {
	net.SignTx(signer, tx)
	return net.SendTx(tx)
}

// SendTxFromNode submits the provided transaction via the specified node;
// transaction must be already signed.
func (net *TestNetwork) SendTxFromNode(tx *transaction.Transaction, node *TestProcessorNode) string {
	hash, err := node.SendTransaction(tx)
	net.handleOrBypassError(err)

	return hash
}

// CreateSignedTx creates a new transaction from provided information and signs
// it with the `sender` wallet; if modified, a transaction must be sent to the
// network using SignAndSendTx instead of SendTx.
func (net *TestNetwork) CreateSignedTx(
	sender *TestWalletAccount,
	recvAddress Address,
	value *big.Int,
	txData []byte,
) *transaction.Transaction {
	tx := net.CreateTx(sender, recvAddress, value, txData)
	net.SignTx(sender, tx)
	return tx
}

// CreateSignedTxUint64 creates a new transaction from provided information and
// signs it with the `sender` wallet; if modified, a transaction must be sent
// to the network using SignAndSendTx instead of SendTx.
func (net *TestNetwork) CreateSignedTxUint64(
	sender *TestWalletAccount,
	recvAddress Address,
	value uint64,
	txData []byte,
) *transaction.Transaction {
	tx := net.CreateTxUint64(sender, recvAddress, value, txData)
	net.SignTx(sender, tx)
	return tx
}

// CreateTxUint64 creates a new transaction from the provided information; must
// be signed before sending; the nonce of the `sender` wallet is incremented.
func (net *TestNetwork) CreateTxUint64(
	sender *TestWalletAccount,
	recvAddress Address,
	value uint64,
	txData []byte,
) *transaction.Transaction {
	return net.CreateTx(sender, recvAddress, big.NewInt(0).SetUint64(value), txData)
}

// CreateTx creates a new transaction from the provided information; must
// be signed before sending; the nonce of the `sender` wallet is incremented.
func (net *TestNetwork) CreateTx(
	sender *TestWalletAccount,
	recvAddress Address,
	value *big.Int,
	txData []byte,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    sender.Nonce,
		Value:    big.NewInt(0).Set(value),
		RcvAddr:  recvAddress,
		SndAddr:  sender.Address,
		GasPrice: net.DefaultGasPrice,
		GasLimit: net.MinGasLimit,
		Data:     txData,
		ChainID:  ChainID,
		Version:  MinTransactionVersion,
	}

	sender.Nonce++
	return tx
}

// SignTx signs a transaction with the provided `signer` wallet.
func (net *TestNetwork) SignTx(signer *TestWalletAccount, tx *transaction.Transaction) {
	txBuff, err := tx.GetDataForSigning(TestAddressPubkeyConverter, TestTxSignMarshalizer)
	net.handleOrBypassError(err)

	signature, err := signer.SingleSigner.Sign(signer.SkTxSign, txBuff)
	net.handleOrBypassError(err)

	tx.Signature = signature
}

// NewAddress creates a new child address of the provided wallet; used to
// compute the address of newly deployed smart contracts.
func (net *TestNetwork) NewAddress(creator *TestWalletAccount) Address {
	address, err := net.DefaultNode.BlockchainHook.NewAddress(
		creator.Address,
		creator.Nonce,
		net.DefaultVM)
	net.handleOrBypassError(err)

	return address
}

// GetAccountHandler retrieves the `state.UserAccountHandler` instance for the
// specified address by querying a node belonging to the shard of the address.
func (net *TestNetwork) GetAccountHandler(address Address) state.UserAccountHandler {
	node := net.firstNodeInShardOfAddress(address)
	account, err := node.AccntState.GetExistingAccount(address)
	net.handleOrBypassError(err)

	accountHandler := account.(state.UserAccountHandler)
	require.NotNil(net.T, accountHandler)

	return accountHandler
}

// ShardOfAddress returns the shard ID of the specified address.
func (net *TestNetwork) ShardOfAddress(address Address) ShardIdentifier {
	return net.DefaultNode.ShardCoordinator.ComputeId(address)
}

// ComputeTxFee calculates the cost of the provided transaction, smart contract
// execution or built-in function calls notwithstanding.
func (net *TestNetwork) ComputeTxFee(tx *transaction.Transaction) *big.Int {
	return net.DefaultNode.EconomicsData.ComputeTxFee(tx)
}

// ComputeTxFeeUint64 calculates the cost of the provided transaction, smart contract
// execution or built-in function calls notwithstanding.
func (net *TestNetwork) ComputeTxFeeUint64(tx *transaction.Transaction) uint64 {
	return net.DefaultNode.EconomicsData.ComputeTxFee(tx).Uint64()
}

// ComputeGasLimit calculates the base gas limit of the provided
// transaction, smart contract execution or built-in function calls
// notwithstanding.
func (net *TestNetwork) ComputeGasLimit(tx *transaction.Transaction) uint64 {
	return net.DefaultNode.EconomicsData.ComputeGasLimit(tx)
}

// RequireWalletNoncesInSyncWithState asserts that the nonces of all wallets
// managed by the test network are in sync with the actual nonces in the
// blockchain state.
func (net *TestNetwork) RequireWalletNoncesInSyncWithState() {
	for i, wallet := range net.Wallets {
		account := net.GetAccountHandler(wallet.Address)
		require.Equal(net.T, wallet.Nonce, account.GetNonce(),
			fmt.Sprintf("wallet %d has nonce out of sync", i))
	}
}

// TODO cannot define SCQueryInt() because it requires vm.GetIntValueFromSC()
// which causes an import cycle.
// func (net *TestNetwork) SCQueryInt(contract Address, function string, args ...[]byte) *big.Int {
// 	shardID := net.ShardOfAddress(contract)
// 	firstNodeInShard := net.NodesSharded[shardID][0]

// 	return vm.GetIntValueFromSC(
// 		net.DefaultGasSchedule,
// 		firstNodeInShard.AccntState,
// 		contract,
// 		function,
// 		args...)
// }

func (net *TestNetwork) createNodes() {
	net.Nodes = CreateNodes(
		net.NumShards,
		net.NodesPerShard,
		net.NodesInMetashard)
}

func (net *TestNetwork) indexProposers() {
	net.Proposers = make([]int, net.NumShards+1)
	for i := 0; i < net.NumShards; i++ {
		net.Proposers[i] = i * net.NodesPerShard
	}
	net.Proposers[net.NumShards] = net.NumShards * net.NodesPerShard
}

func (net *TestNetwork) mapNodesByShard() {
	net.NodesSharded = make(NodesByShardMap)
	for _, node := range net.Nodes {
		shardID := node.ShardCoordinator.SelfId()
		net.NodesSharded[shardID] = append(net.NodesSharded[shardID], node)
	}
}

func (net *TestNetwork) startNodes() {
	DisplayAndStartNodes(net.Nodes)
}

func (net *TestNetwork) initDefaults() {
	net.DeploymentAddress = make(Address, 32)
	net.DefaultNode = net.Nodes[0]
	net.DefaultGasPrice = MinTxGasPrice
	net.DefaultGasSchedule = nil
	net.DefaultVM = factory.ArwenVirtualMachine

	defaultNodeShardID := net.DefaultNode.ShardCoordinator.SelfId()
	net.MinGasLimit = MinTxGasLimit
	net.MaxGasLimit = net.DefaultNode.EconomicsData.MaxGasLimitPerBlock(defaultNodeShardID) - 1
}

func (net *TestNetwork) closeNodes() {
	for _, node := range net.Nodes {
		err := node.Messenger.Close()
		net.handleOrBypassError(err)
	}
}

func (net *TestNetwork) firstNodeInShard(shardID ShardIdentifier) *TestProcessorNode {
	firstNodeInShard := net.NodesSharded[shardID][0]
	require.NotNil(net.T, firstNodeInShard)
	return firstNodeInShard
}

func (net *TestNetwork) firstNodeInShardOfAddress(address Address) *TestProcessorNode {
	shardID := net.ShardOfAddress(address)
	firstNodeInShard := net.NodesSharded[shardID][0]
	require.NotNil(net.T, firstNodeInShard)
	return firstNodeInShard
}

func (net *TestNetwork) handleOrBypassError(err error) {
	if net.BypassErrorsOnce {
		net.BypassErrorsOnce = false
		return
	}

	require.Nil(net.T, err)
}
