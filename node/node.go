package node

import (
	"context"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/pkg/errors"
)

type Option func (*Node)
type Node struct {
	port int
	marshalizer marshal.Marshalizer
	ctx context.Context
	hasher hashing.Hasher
	maxAllowedPeers int
	pubSubStrategy p2p.PubSubStrategy
}

func NewNode(opts ...Option) *Node {
	node := &Node{}
	for _, opt := range opts {
		opt(node)
	}
	return node
}

func (n *Node) ApplyOptions(opts ...Option)  {
	for _, opt := range opts {
		opt(n)
	}
}

func (n *Node) Start() error {
	messenger, err := n.createNetMessenger()
	if err != nil {
		return err
	}



	return nil
}

func (n *Node) createNetMessenger() (p2p.Messenger, error) {
	if n.port == 0 {
		return nil, errors.New("Cannot start node on port 0")
	}
	cp, err := p2p.NewConnectParamsFromPort(n.port)
	if err != nil {
		return nil, err
	}

	nm, err := p2p.NewNetMessenger(n.ctx, n.marshalizer, n.hasher, cp,  n.maxAllowedPeers, n.pubSubStrategy)
	return nm, nil
}

func WithPort(port int) Option {
	return func(n *Node) {
		n.port = port
	}
}

func WithMarshalizer(marshalizer marshal.Marshalizer) Option {
	return func(n *Node) {
		n.marshalizer = marshalizer
	}
}

func WithContext(ctx context.Context) Option {
	return func(n *Node) {
		n.ctx = ctx
	}
}

func WithHasher(hasher hashing.Hasher) Option {
	return func(n *Node) {
		n.hasher = hasher
	}
}


