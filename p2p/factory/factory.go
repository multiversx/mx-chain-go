package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-p2p-go/libp2p"
	p2pCrypto "github.com/multiversx/mx-chain-p2p-go/libp2p/crypto"
	"github.com/multiversx/mx-chain-p2p-go/message"
	"github.com/multiversx/mx-chain-p2p-go/peersHolder"
	"github.com/multiversx/mx-chain-p2p-go/rating"
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

// NewPeersRatingHandler returns a new peers rating handler
func NewPeersRatingHandler(args ArgPeersRatingHandler) (p2p.PeersRatingHandler, error) {
	return rating.NewPeersRatingHandler(args)
}

// NewPeersHolder returns a new instance of peersHolder
func NewPeersHolder(preferredConnectionAddresses []string) (p2p.PreferredPeersHolderHandler, error) {
	return peersHolder.NewPeersHolder(preferredConnectionAddresses)
}

// ConvertPublicKeyToPeerID will convert a public key to core.PeerID
func ConvertPublicKeyToPeerID(pk crypto.PublicKey) (core.PeerID, error) {
	return p2pCrypto.ConvertPublicKeyToPeerID(pk)
}
