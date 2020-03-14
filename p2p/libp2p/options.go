package libp2p

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var readTimeout = time.Second * 10

// Option represents a functional configuration parameter that can operate
//  over the networkMessenger struct.
type Option func(*networkMessenger) error

// WithAuthentication sets up the authentication mechanism and peer id - public key collection
func WithAuthentication(
	networkShardingCollector p2p.NetworkShardingCollector,
	signerVerifier p2p.SignerVerifier,
	marshalizer p2p.Marshalizer,
) Option {
	return func(mes *networkMessenger) error {
		if check.IfNil(networkShardingCollector) {
			return p2p.ErrNilNetworkShardingCollector
		}
		if check.IfNil(signerVerifier) {
			return p2p.ErrNilSignerVerifier
		}
		if check.IfNil(marshalizer) {
			return p2p.ErrNilMarshalizer
		}

		var err error
		mes.ip, err = NewIdentityProvider(
			mes.p2pHost,
			networkShardingCollector,
			signerVerifier,
			marshalizer,
			readTimeout,
		)
		if err != nil {
			return err
		}

		mes.p2pHost.Network().Notify(mes.ip)

		return nil
	}
}
