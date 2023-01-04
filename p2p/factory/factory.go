package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p"
	p2pCrypto "github.com/ElrondNetwork/elrond-go-p2p/libp2p/crypto"
	"github.com/ElrondNetwork/elrond-go-p2p/message"
	"github.com/ElrondNetwork/elrond-go-p2p/peersHolder"
	"github.com/ElrondNetwork/elrond-go-p2p/rating"
	"github.com/ElrondNetwork/elrond-go/p2p"
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
