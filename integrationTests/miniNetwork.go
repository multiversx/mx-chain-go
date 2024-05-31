package integrationTests

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// MiniNetwork is a mini network, useful for some integration tests
type MiniNetwork struct {
	Round uint64
	Nonce uint64

	Nodes         []*TestProcessorNode
	ShardNode     *TestProcessorNode
	MetachainNode *TestProcessorNode
	Users         map[string]*TestWalletAccount
}

// NewMiniNetwork creates a MiniNetwork
func NewMiniNetwork() *MiniNetwork {
	n := &MiniNetwork{}

	nodes := CreateNodes(
		1,
		1,
		1,
	)

	n.Nodes = nodes
	n.ShardNode = nodes[0]
	n.MetachainNode = nodes[1]
	n.Users = make(map[string]*TestWalletAccount)

	return n
}

// Stop stops the mini network
func (n *MiniNetwork) Stop() {
	n.ShardNode.Close()
	n.MetachainNode.Close()
}

// FundAccount funds an account
func (n *MiniNetwork) FundAccount(address []byte, value *big.Int) {
	shard := n.MetachainNode.ShardCoordinator.ComputeId(address)

	if shard == n.MetachainNode.ShardCoordinator.SelfId() {
		MintAddress(n.MetachainNode.AccntState, address, value)
	} else {
		MintAddress(n.ShardNode.AccntState, address, value)
	}
}

// AddUser adds a user (account) to the mini network
func (n *MiniNetwork) AddUser(balance *big.Int) *TestWalletAccount {
	user := CreateTestWalletAccount(n.ShardNode.ShardCoordinator, 0)
	n.Users[string(user.Address)] = user
	n.FundAccount(user.Address, balance)
	return user
}

// Start starts the mini network
func (n *MiniNetwork) Start() {
	n.Round = 1
	n.Nonce = 1
}

// Continue advances processing with a number of rounds
func (n *MiniNetwork) Continue(t *testing.T, numRounds int) {
	idxProposers := []int{0, 1}

	for i := int64(0); i < int64(numRounds); i++ {
		n.Nonce, n.Round = ProposeAndSyncOneBlock(t, n.Nodes, idxProposers, n.Round, n.Nonce)
	}
}

// SendTransaction sends a transaction
func (n *MiniNetwork) SendTransaction(
	senderPubkey []byte,
	receiverPubkey []byte,
	value *big.Int,
	data string,
	additionalGasLimit uint64,
) (string, error) {
	sender, ok := n.Users[string(senderPubkey)]
	if !ok {
		return "", fmt.Errorf("unknown sender: %s", hex.EncodeToString(senderPubkey))
	}

	tx := &transaction.Transaction{
		Nonce:    sender.Nonce,
		Value:    new(big.Int).Set(value),
		SndAddr:  sender.Address,
		RcvAddr:  receiverPubkey,
		Data:     []byte(data),
		GasPrice: MinTxGasPrice,
		GasLimit: MinTxGasLimit + uint64(len(data)) + additionalGasLimit,
		ChainID:  ChainID,
		Version:  MinTransactionVersion,
	}

	txBuff, _ := tx.GetDataForSigning(TestAddressPubkeyConverter, TestTxSignMarshalizer, TestTxSignHasher)
	tx.Signature, _ = sender.SingleSigner.Sign(sender.SkTxSign, txBuff)
	txHash, err := n.ShardNode.SendTransaction(tx)

	sender.Nonce++

	return txHash, err
}
