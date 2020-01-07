package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// WithPeerBlackList defines the option of setting a peer black list handler
func WithPeerBlackList(blacklistHandler p2p.BlacklistHandler) p2p.Option {
	return func(cfg *p2p.Config) error {
		if cfg == nil {
			return p2p.ErrNilConfigVariable
		}
		if check.IfNil(blacklistHandler) {
			return p2p.ErrNilPeerBlacklistHandler
		}

		cfg.BlacklistHandler = blacklistHandler

		return nil
	}
}
