package discovery

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
)

// ConnectableHost is an enhanced Host interface that has the ability to connect to a string address
type ConnectableHost interface {
	host.Host
	ConnectToPeer(ctx context.Context, address string) error
	IsInterfaceNil() bool
}
