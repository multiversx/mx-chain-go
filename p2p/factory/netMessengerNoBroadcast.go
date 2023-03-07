package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	p2p "github.com/multiversx/mx-chain-p2p-go"
	"github.com/multiversx/mx-chain-p2p-go/libp2p"
)

type netMessengerNoBroadcast struct {
	p2p.Messenger
}

// NewNetMessengerNoBroadcast -
func NewNetMessengerNoBroadcast(args ArgsNetworkMessenger) (*netMessengerNoBroadcast, error) {
	messenger, err := libp2p.NewNetworkMessenger(args)
	if err != nil {
		return nil, err
	}

	return &netMessengerNoBroadcast{
		Messenger: messenger,
	}, nil
}

// BroadcastOnChannelBlocking does nothing and returns nil
func (netMessenger *netMessengerNoBroadcast) BroadcastOnChannelBlocking(_ string, _ string, _ []byte) error {
	return nil
}

// BroadcastOnChannel does nothing
func (netMessenger *netMessengerNoBroadcast) BroadcastOnChannel(_ string, _ string, _ []byte) {
}

// BroadcastUsingPrivateKey does nothing
func (netMessenger *netMessengerNoBroadcast) BroadcastUsingPrivateKey(_ string, _ []byte, _ core.PeerID, _ []byte) {
}

// Broadcast does nothing
func (netMessenger *netMessengerNoBroadcast) Broadcast(_ string, _ []byte) {
}
