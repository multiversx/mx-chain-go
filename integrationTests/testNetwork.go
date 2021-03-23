package integrationTests

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/require"
)

type ShardIdentifier = uint32
type Address = []byte
type NodeSlice = []*TestProcessorNode
type NodesByShardMap = map[ShardIdentifier]NodeSlice
type GasScheduleMap = map[string]map[string]uint64

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
	Advertiser         p2p.Messenger
	AdvertiserAddress  string
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

func NewTestNetwork(t *testing.T) *TestNetwork {
	// TODO replace testing.T with testing.TB everywhere in integrationTest
	return &TestNetwork{
		T:                t,
		BypassErrorsOnce: false,
	}
}

func NewTestNetworkSized(t *testing.T, numShards int, nodesPerShard int, nodesInMetashard int) *TestNetwork {
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

	net.createAdvertiser()
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
	net.closeAdvertiser()
	net.closeNodes()
}

// MintNodeAccounts adds the specified value to the accounts owned by the nodes
// of the TestNetwork.
func (net *TestNetwork) MintNodeAccounts(value *big.Int) {
	MintAllNodes(net.Nodes, value)
}

// MintNodeAccountsInt64 adds the specified value to the accounts owned by the
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
	MintAllPlayers(net.Nodes, net.Wallets, value)
}

// MintWalletsInt64 adds the specified value to the test wallets.
func (net *TestNetwork) MintWalletsUint64(value uint64) {
	MintAllPlayers(net.Nodes, net.Wallets, big.NewInt(0).SetUint64(value))
}

func (net *TestNetwork) SendTx(tx *transaction.Transaction) string {
	node := net.firstNodeInShardOfAddress(tx.SndAddr)
	return net.SendTxFromNode(tx, node)
}

func (net *TestNetwork) SignAndSendTx(sender *TestWalletAccount, tx *transaction.Transaction) string {
	net.SignTx(sender, tx)
	return net.SendTx(tx)
}

func (net *TestNetwork) SendTxFromNode(tx *transaction.Transaction, node *TestProcessorNode) string {
	hash, err := node.SendTransaction(tx)
	net.handleOrBypassError(err)

	return hash
}

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

func (net *TestNetwork) CreateTxUint64(
	sender *TestWalletAccount,
	recvAddress Address,
	value uint64,
	txData []byte,
) *transaction.Transaction {
	return net.CreateTx(sender, recvAddress, big.NewInt(0).SetUint64(value), txData)
}

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

func (net *TestNetwork) SignTx(signer *TestWalletAccount, tx *transaction.Transaction) {
	txBuff, err := tx.GetDataForSigning(TestAddressPubkeyConverter, TestTxSignMarshalizer)
	net.handleOrBypassError(err)

	signature, err := signer.SingleSigner.Sign(signer.SkTxSign, txBuff)
	net.handleOrBypassError(err)

	tx.Signature = signature
}

func (net *TestNetwork) NewAddress(creator *TestWalletAccount) Address {
	address, err := net.DefaultNode.BlockchainHook.NewAddress(
		creator.Address,
		creator.Nonce,
		net.DefaultVM)
	net.handleOrBypassError(err)

	return address
}

func (net *TestNetwork) GetAccountHandler(address Address) state.UserAccountHandler {
	node := net.firstNodeInShardOfAddress(address)
	account, err := node.AccntState.GetExistingAccount(address)
	net.handleOrBypassError(err)

	accountHandler := account.(state.UserAccountHandler)
	require.NotNil(net.T, accountHandler)

	return accountHandler
}

func (net *TestNetwork) ShardOfAddress(address Address) ShardIdentifier {
	return net.DefaultNode.ShardCoordinator.ComputeId(address)
}

func (net *TestNetwork) ComputeTxFee(tx *transaction.Transaction) *big.Int {
	return net.DefaultNode.EconomicsData.ComputeTxFee(tx)
}

func (net *TestNetwork) ComputeTxFeeUint64(tx *transaction.Transaction) uint64 {
	return net.DefaultNode.EconomicsData.ComputeTxFee(tx).Uint64()
}

func (net *TestNetwork) ComputeGasLimit(tx *transaction.Transaction) uint64 {
	return net.DefaultNode.EconomicsData.ComputeGasLimit(tx)
}

func (net *TestNetwork) RequireWalletNoncesInSyncWithState() {
	for _, wallet := range net.Wallets {
		account := net.GetAccountHandler(wallet.Address)
		require.Equal(net.T, wallet.Nonce, account.GetNonce(), "wallet nonce out of sync")
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

func (net *TestNetwork) createAdvertiser() {
	net.Advertiser = CreateMessengerWithKadDht(net.AdvertiserAddress)
	err := net.Advertiser.Bootstrap(0)
	net.handleOrBypassError(err)
}

func (net *TestNetwork) createNodes() {
	net.Nodes = CreateNodes(
		net.NumShards,
		net.NodesPerShard,
		net.NodesInMetashard,
		GetConnectableAddress(net.Advertiser))
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
		_, exists := net.NodesSharded[shardID]
		if !exists {
			net.NodesSharded[shardID] = make(NodeSlice, 0)
		}
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

func (net *TestNetwork) closeAdvertiser() {
	err := net.Advertiser.Close()
	net.handleOrBypassError(err)
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
	if net.BypassErrorsOnce == true {
		net.BypassErrorsOnce = false
		return
	}

	require.Nil(net.T, err)
}
