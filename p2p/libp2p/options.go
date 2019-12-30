package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// Option represents a functional configuration parameter that can operate
//  over the networkMessenger struct.
type Option func(*networkMessenger) error

// WithPeerBlackList defines the option of setting a peer black list handler
func WithPeerBlackList(blacklistHandler p2p.BlacklistHandler) Option {
	return func(mes *networkMessenger) error {
		return mes.ctxProvider.SetPeerBlacklist(blacklistHandler)
	}
}
