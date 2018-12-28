package node_test

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNewNode(t *testing.T) {
	t.Parallel()
	n := node.NewNode()
	assert.NotNil(t, n)
}

func TestNewNodeNotRunning(t *testing.T) {
	t.Parallel()
	n := node.NewNode()
	assert.False(t, n.IsRunning())
}

func TestStartNoPort(t *testing.T) {
	t.Parallel()
	n := node.NewNode()
	err := n.Start()
	assert.NotNil(t, err)
}

func TestStartNoMarshalizer(t *testing.T) {
	t.Parallel()
	n := node.NewNode(node.WithPort(4000))
	err := n.Start()
	assert.NotNil(t, err)
}

func TestStartNoHasher(t *testing.T) {
	t.Parallel()
	n := node.NewNode(node.WithPort(4000), node.WithMarshalizer(mock.Marshalizer{}))
	err := n.Start()
	assert.NotNil(t, err)
}

func TestStartNoMaxAllowedPeers(t *testing.T) {
	t.Parallel()
	n := node.NewNode(node.WithPort(4000), node.WithMarshalizer(mock.Marshalizer{}), node.WithHasher(mock.Hasher{}))
	err := n.Start()
	assert.NotNil(t, err)
}

func TestStartCorrectParams(t *testing.T) {
	t.Parallel()
	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
	)
	err := n.Start()
	assert.Nil(t, err)
	assert.True(t, n.IsRunning())
}

func TestStartCorrectParamsApplyingOptions(t *testing.T) {
	t.Parallel()
	n := node.NewNode()
	err := n.ApplyOptions(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)

	logError(err)

	err = n.Start()
	assert.Nil(t, err)
	assert.True(t, n.IsRunning())
}

func TestNode_ApplyOptions_NodeStarted(t *testing.T) {
	t.Parallel()
	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n.Start()
	logError(err)

	err = n.ApplyOptions(
		node.WithMaxAllowedPeers(4),
	)

	assert.NotNil(t, err)
	assert.True(t, n.IsRunning())
}

func TestNode_Stop_NotStartedYet(t *testing.T) {
	t.Parallel()
	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
	)
	err := n.Start()
	logError(err)
	err = n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestNode_Stop(t *testing.T) {
	t.Parallel()
	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
	)

	err := n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestNode_ConnectToInitialAddressesNodeNotStarted(t *testing.T) {
	t.Parallel()
	n2 := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithInitialNodeAddresses([]string{addr}),
	)

	err = n.ConnectToInitialAddresses()
	assert.NotNil(t, err)
}

func TestNode_ConnectToNilInitialAddresses(t *testing.T) {
	t.Parallel()

	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n.Start()
	assert.Nil(t, err)
	err = n.ConnectToInitialAddresses()
	assert.NotNil(t, err)
}

func TestNode_ConnectToInitialAddresses(t *testing.T) {
	t.Parallel()
	n2 := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithInitialNodeAddresses([]string{addr}),
	)
	err = n.Start()
	assert.Nil(t, err)

	err = n.ConnectToInitialAddresses()
	assert.Nil(t, err)
}

func TestNode_ConnectToAddressesNodeNotStarted(t *testing.T) {
	t.Parallel()
	n2 := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)

	err = n.ConnectToAddresses([]string{addr})
	assert.NotNil(t, err)
}

func TestNode_ConnectToAddresses(t *testing.T) {
	t.Parallel()
	n2 := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err = n.Start()
	assert.Nil(t, err)

	err = n.ConnectToAddresses([]string{addr})
	assert.Nil(t, err)
}

func TestNode_AddressNodeNotStarted(t *testing.T) {
	t.Parallel()
	n := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	_, err := n.Address()
	assert.NotNil(t, err)
}
