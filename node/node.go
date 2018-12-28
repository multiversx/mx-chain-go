package node

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"

	"github.com/pkg/errors"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type topic string

const (
	transactionTopic topic = "tx"
)

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func(*Node) error

// Node is a structure that passes the configuration parameters and initializes
//  required services as requested
type Node struct {
	port                  int
	marshalizer           marshal.Marshalizer
	ctx                   context.Context
	hasher                hashing.Hasher
	maxAllowedPeers       int
	pubSubStrategy        p2p.PubSubStrategy
	initialNodesAddresses []string
	messenger             p2p.Messenger
	accounts			  state.AccountsAdapter
	addrConverter		  state.AddressConverter
	privateKey			  crypto.PrivateKey
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) (*Node, error) {
	node := &Node{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		err := opt(node)
		if err != nil {
			return nil, errors.New("error applying option: " + err.Error())
		}
	}
	return node, nil
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option) error {
	if n.IsRunning() {
		return errors.New("cannot apply options while node is running")
	}
	for _, opt := range opts {
		err := opt(n)
		if err != nil {
			return errors.New("error applying option: " + err.Error())
		}
	}
	return nil
}

// IsRunning will return the current state of the node
func (n *Node) IsRunning() bool {
	return n.messenger != nil
}

// Address returns the first address of the running node
func (n *Node) Address() (string, error) {
	if !n.IsRunning() {
		return "", errors.New("node is not started yet")
	}
	return n.messenger.Addresses()[0], nil
}

// Start will create a new messenger and and set up the Node state as running
func (n *Node) Start() error {
	messenger, err := n.createNetMessenger()
	if err != nil {
		return err
	}
	n.messenger = messenger

	return nil
}

// Stop closes the messenger and undos everything done in Start
func (n *Node) Stop() error {
	if !n.IsRunning() {
		return nil
	}
	err := n.messenger.Close()
	if err != nil {
		return err
	}

	n.messenger = nil
	return nil
}

// ConnectToInitialAddresses connect to the list of peers provided initialAddresses
func (n *Node) ConnectToInitialAddresses() error {
	if !n.IsRunning() {
		return errors.New("node is not started yet")
	}
	if n.initialNodesAddresses == nil {
		return errors.New("no addresses to connect to")
	}
	// Don't try to connect to self
	tmp := n.removeSelfFromList(n.initialNodesAddresses)
	n.messenger.ConnectToAddresses(n.ctx, tmp)
	return nil
}

// ConnectToAddresses will take a slice of addresses and try to connect to all of them.
func (n *Node) ConnectToAddresses(addresses []string) error {
	if !n.IsRunning() {
		return errors.New("node is not started yet")
	}
	n.messenger.ConnectToAddresses(n.ctx, addresses)
	return nil
}

// StartConsensus will start the consesus service for the current node
func (n *Node) StartConsensus() error {
	return nil
}

// GetBalance gets the balance for a specific address
func (n *Node) GetBalance(address string) (*big.Int, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}
	accAddress, err := n.addrConverter.CreateAddressFromHex(address)
	if err != nil {
		return nil, errors.New("invalid address: " + err.Error())
	}
	account, err := n.accounts.GetExistingAccount(accAddress)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}

	if account == nil {
		return big.NewInt(0), nil
	}

	return &account.BaseAccount().Balance, nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(sender string, receiver string, amount big.Int, code string) (*transaction.Transaction, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	if n.privateKey == nil {
		return nil, errors.New("initialize PrivateKey first")
	}

	senderAddress, err := n.addrConverter.CreateAddressFromHex(sender)
	if err != nil {
		return nil, errors.New("could not create sender address from provided param")
	}
	senderAccount, err := n.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}
	newNonce := uint64(0)
	if senderAccount != nil {
		newNonce = senderAccount.BaseAccount().Nonce + 1
	}

	tx := transaction.Transaction{
		Nonce: newNonce,
		Value: amount,
		RcvAddr: []byte(receiver),
		SndAddr: []byte(sender),
	}

	txToByteArray, err := n.marshalizer.Marshal(tx)
	if err != nil {
		return nil, errors.New("could not create byte array representation of the transaction")
	}

	sig, err := n.privateKey.Sign(txToByteArray)
	if err != nil {
		return nil, errors.New("could not sign the transaction")
	}
	tx.Signature = sig

	return &tx, nil
}

// SendTransaction will send a new transaction on the topic channel
func (n *Node) SendTransaction(
	nonce uint64,
	sender string,
	receiver string,
	value big.Int,
	transactionData string,
	signature string) (*transaction.Transaction, error) {

		tx := transaction.Transaction{
			Nonce:     nonce,
			Value:     value,
			RcvAddr:   []byte(receiver),
			SndAddr:   []byte(sender),
			Data:      []byte(transactionData),
			Signature: []byte(signature),
		}
		topic := n.messenger.GetTopic(string(transactionTopic))
		err := topic.Broadcast(tx)
		if err != nil {
			return nil, errors.New("could not broadcast transaction: " + err.Error())
		}
		return &tx, nil
}

//GetTransaction gets the transaction
func (n *Node) GetTransaction(hash string) (*transaction.Transaction, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func (n *Node) createNetMessenger() (p2p.Messenger, error) {
	if n.port == 0 {
		return nil, errors.New("Cannot start node on port 0")
	}

	if n.maxAllowedPeers == 0 {
		return nil, errors.New("Cannot start node without providing maxAllowedPeers")
	}

	cp, err := p2p.NewConnectParamsFromPort(n.port)
	if err != nil {
		return nil, err
	}

	nm, err := p2p.NewNetMessenger(n.ctx, n.marshalizer, n.hasher, cp, n.maxAllowedPeers, n.pubSubStrategy)
	if err != nil {
		return nil, err
	}
	return nm, nil
}

func (n *Node) removeSelfFromList(peers []string) []string {
	tmp := peers[:0]
	addr, _ := n.Address()
	for _, p := range peers {
		if addr != p {
			tmp = append(tmp, p)
		}
	}
	return tmp
}

// WithPort sets up the port option for the Node
func WithPort(port int) Option {
	return func(n *Node) error {
		n.port = port
		return nil
	}
}

// WithMarshalizer sets up the marshalizer option for the Node
func WithMarshalizer(marshalizer marshal.Marshalizer) Option {
	return func(n *Node) error {
		if marshalizer == nil {
			return errors.New("trying to set nil marshalizer")
		}
		n.marshalizer = marshalizer
		return nil
	}
}

// WithContext sets up the context option for the Node
func WithContext(ctx context.Context) Option {
	return func(n *Node) error {
		if ctx == nil {
			return errors.New("trying to set nil context")
		}
		n.ctx = ctx
		return nil
	}
}

// WithHasher sets up the hasher option for the Node
func WithHasher(hasher hashing.Hasher) Option {
	return func(n *Node) error {
		if hasher == nil {
			return errors.New("trying to set nil hasher")
		}
		n.hasher = hasher
		return nil
	}
}

// WithMaxAllowedPeers sets up the maxAllowedPeers option for the Node
func WithMaxAllowedPeers(maxAllowedPeers int) Option {
	return func(n *Node) error {
		n.maxAllowedPeers = maxAllowedPeers
		return nil
	}
}

// WithPubSubStrategy sets up the strategy option for the Node
func WithPubSubStrategy(strategy p2p.PubSubStrategy) Option {
	return func(n *Node) error {
		n.pubSubStrategy = strategy
		return nil
	}
}

// WithInitialNodeAddresses sets up the initial node addresses for the Node
func WithInitialNodeAddresses(addresses []string) Option {
	return func(n *Node) error {
		n.initialNodesAddresses = addresses
		return nil
	}
}

// WithAccountsAdapter sets up the accounts adapter for the Node
func WithAccountsAdapter(accounts state.AccountsAdapter) Option {
	return func(n *Node) error {
		if accounts == nil {
			return errors.New("trying to set nil accounts adapter")
		}
		n.accounts = accounts
		return nil
	}
}

// WithAddressConverter sets up the address converter adapter for the Node
func WithAddressConverter(addrConverter state.AddressConverter) Option {
	return func(n *Node) error {
		if addrConverter == nil {
			return errors.New("trying to set nil accounts adapter")
		}
		n.addrConverter = addrConverter
		return nil
	}
}

// WithPrivateKey sets up the private key for the Node
func WithPrivateKey(sk crypto.PrivateKey) Option {
	return func(n *Node) error {
		if sk == nil {
			return errors.New("trying to set nil private key")
		}
		n.privateKey = sk
		return nil
	}
}
