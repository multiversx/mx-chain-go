package node

import (
	"context"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"

	"github.com/pkg/errors"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func(*Node)

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
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) *Node {
	node := &Node{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(node)
	}
	return node
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option) error {
	if n.IsRunning() {
		return errors.New("cannot apply options while node is running")
	}
	for _, opt := range opts {
		opt(n)
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

func (n *Node) GetBalance(address string) (*big.Int, error) {
	return big.NewInt(0), nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(sender string, receiver string, amount big.Int, code string) (string, error) {
	return "", nil
}

//GetTransaction gets the transaction
func (n *Node) GetTransaction(hash string) (*transaction.Transaction, error) {
	return nil, nil
}

func (n *Node) createNetMessenger() (p2p.Messenger, error) {
	if n.port == 0 {
		return nil, errors.New("Cannot start node on port 0")
	}
	if n.marshalizer == nil {
		return nil, errors.New("Canot start node without providing a marshalizer")
	}
	if n.hasher == nil {
		return nil, errors.New("Canot start node without providing a hasher")
	}
	if n.maxAllowedPeers == 0 {
		return nil, errors.New("Canot start node without providing maxAllowedPeers")
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
	return func(n *Node) {
		n.port = port
	}
}

// WithMarshalizer sets up the marshalizer option for the Node
func WithMarshalizer(marshalizer marshal.Marshalizer) Option {
	return func(n *Node) {
		n.marshalizer = marshalizer
	}
}

// WithContext sets up the context option for the Node
func WithContext(ctx context.Context) Option {
	return func(n *Node) {
		n.ctx = ctx
	}
}

// WithHasher sets up the hasher option for the Node
func WithHasher(hasher hashing.Hasher) Option {
	return func(n *Node) {
		n.hasher = hasher
	}
}

// WithMaxAllowedPeers sets up the maxAllowedPeers option for the Node
func WithMaxAllowedPeers(maxAllowedPeers int) Option {
	return func(n *Node) {
		n.maxAllowedPeers = maxAllowedPeers
	}
}

// WithMaxAllowedPeers sets up the strategy option for the Node
func WithPubSubStrategy(strategy p2p.PubSubStrategy) Option {
	return func(n *Node) {
		n.pubSubStrategy = strategy
	}
}

func WithInitialNodeAddresses(addresses []string) Option {
	return func(n *Node) {
		n.initialNodesAddresses = addresses
	}
}
