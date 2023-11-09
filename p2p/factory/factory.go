package factory

import (
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/crypto"
	"github.com/multiversx/mx-chain-communication-go/p2p/message"
	messagecheck "github.com/multiversx/mx-chain-communication-go/p2p/messageCheck"
	"github.com/multiversx/mx-chain-communication-go/p2p/peersHolder"
	"github.com/multiversx/mx-chain-communication-go/p2p/rating"
	"github.com/multiversx/mx-chain-go/p2p"
)

// ArgsNetworkMessenger defines the options used to create a p2p wrapper
type ArgsNetworkMessenger = libp2p.ArgsNetworkMessenger

// NewNetworkMessenger creates a libP2P messenger by opening a port on the current machine
func NewNetworkMessenger(args ArgsNetworkMessenger) (p2p.Messenger, error) {
	return libp2p.NewNetworkMessenger(args)
}

// LocalSyncTimer uses the local system to provide the current time
type LocalSyncTimer = libp2p.LocalSyncTimer

// Message is a data holder struct
type Message = message.Message

// PeerShard represents the data regarding a new direct connection`s info
type PeerShard = message.PeerShard

// ArgPeersRatingHandler is the DTO used to create a new peers rating handler
type ArgPeersRatingHandler = rating.ArgPeersRatingHandler

// ArgPeersRatingMonitor is the DTO used to create a new peers rating monitor
type ArgPeersRatingMonitor = rating.ArgPeersRatingMonitor

// ArgsMessageVerifier defines the args used to create a message verifier
type ArgsMessageVerifier = messagecheck.ArgsMessageVerifier

// NewPeersRatingHandler returns a new peers rating handler
func NewPeersRatingHandler(args ArgPeersRatingHandler) (p2p.PeersRatingHandler, error) {
	return rating.NewPeersRatingHandler(args)
}

// NewPeersRatingMonitor returns a new peers rating monitor
func NewPeersRatingMonitor(args ArgPeersRatingMonitor) (p2p.PeersRatingMonitor, error) {
	return rating.NewPeersRatingMonitor(args)
}

// NewPeersHolder returns a new instance of peersHolder
func NewPeersHolder(preferredConnectionAddresses []string) (p2p.PreferredPeersHolderHandler, error) {
	return peersHolder.NewPeersHolder(preferredConnectionAddresses)
}

// NewP2PKeyConverter returns a new instance of p2pKeyConverter
func NewP2PKeyConverter() p2p.P2PKeyConverter {
	return crypto.NewP2PKeyConverter()
}

// NewMessageVerifier will return a new instance of messages verifier
func NewMessageVerifier(args ArgsMessageVerifier) (p2p.P2PSigningHandler, error) {
	return messagecheck.NewMessageVerifier(args)
}
