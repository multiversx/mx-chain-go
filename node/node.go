package node

import (
	"context"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/pkg/errors"
)

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func (*Node)

// Node is a structure that passes the configuration parameters and initializes
//  required services as requested
type Node struct {
	port int
	marshalizer marshal.Marshalizer
	ctx context.Context
	hasher hashing.Hasher
	maxAllowedPeers int
	pubSubStrategy p2p.PubSubStrategy
	initialNodesAddresses []string
	messenger p2p.Messenger
	isRunning bool
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) *Node {
	node := &Node{
		isRunning: false,
		ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(node)
	}
	return node
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option)  {
	for _, opt := range opts {
		opt(n)
	}
}

// IsRunning will return the current state of the node
func (n *Node) IsRunning() bool {
	return n.isRunning
}

// Address returns the first address of the running node
func (n *Node) Address() (string, error) {
	if !n.isRunning {
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
	n.isRunning = true

	return nil
}

// ConnectToAddresses will take a slice of addresses and try to connect to all of them.
func (n *Node) ConnectToAddresses(peers []string) error {
	if n.messenger == nil {
		return errors.New("node is not started yet")
	}
	// Don't try to connect to self
	tmp := peers[:0]
	for _, p := range peers {
		if n.messenger.Addresses()[0] != p {
			tmp = append(tmp, p)
		}
	}
	n.messenger.ConnectToAddresses(n.ctx, tmp)
	return nil
}

// StartConsensus will start the consesus service for the current node
func (n *Node) StartConsensus() error {
	return nil
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

	nm, err := p2p.NewNetMessenger(n.ctx, n.marshalizer, n.hasher, cp,  n.maxAllowedPeers, n.pubSubStrategy)
	return nm, nil
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

