package discovery

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	protocol "github.com/libp2p/go-libp2p-core/protocol"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	hardCPSLimitLow  = 2
	hardCPSLimitHigh = 100
)

type hostDecorator struct {
	h            host.Host
	acceptedChan chan struct{}
	resumeChan   chan struct{}
	timeout      time.Duration
}

// NewHostDecorator creats a new decorator around an existing host (h) that will olny allow
// cps connections per second and drop the connection request upon timeout expiration
func NewHostDecorator(h host.Host, ctx context.Context, cps uint32, timeout time.Duration) (*hostDecorator, error) {
	if h == nil {
		return nil, p2p.ErrNilHost
	}

	if ctx == nil {
		return nil, p2p.ErrNilContext
	}

	if timeout < 0 {
		return nil, p2p.ErrInvalidDurationProvided
	}

	if cps < hardCPSLimitLow {
		cps = hardCPSLimitLow
	}

	if cps > hardCPSLimitHigh {
		cps = hardCPSLimitHigh
	}

	ret := &hostDecorator{
		h:            h,
		acceptedChan: make(chan struct{}, cps),
		resumeChan:   make(chan struct{}),
		timeout:      timeout,
	}

	go func() {
		defer close(ret.acceptedChan)
		for {
			running := true
			select {
			case ret.acceptedChan <- struct{}{}:
			default:
				running = false
			}

			if running {
				select {
				case <-time.After(time.Second / time.Duration(cps)):
				case <-ctx.Done():
					return
				}
			} else {
				select {
				case <-ret.resumeChan:
				case <-ctx.Done():
					return
				}

			}
		}
	}()

	return ret, nil
}

// Connect tryes to connect connect to pi if the Connetions Per Second (hd.cps)
// allows it.
func (hd *hostDecorator) Connect(ctx context.Context, pi peer.AddrInfo) error {

	select {
	case hd.resumeChan <- struct{}{}:
	default:
	}

	select {
	case _, ok := <-hd.acceptedChan:
		if ok {
			return hd.h.Connect(ctx, pi)
		} else {
			return p2p.ErrContextDone
		}

	case <-ctx.Done():
		return p2p.ErrContextDone
	case <-time.After(hd.timeout):
		return p2p.ErrTimeout
	}
}

// --- just pass to h

// ID returns the (local) peer.ID associated with this Host
func (hd *hostDecorator) ID() peer.ID {
	return hd.h.ID()
}

// Peerstore returns the Host's repository of Peer Addresses and Keys.
func (hd *hostDecorator) Peerstore() peerstore.Peerstore {
	return hd.h.Peerstore()
}

// Returns the listen addresses of the Host
func (hd *hostDecorator) Addrs() []ma.Multiaddr {
	return hd.h.Addrs()
}

// Networks returns the Network interface of the Host
func (hd *hostDecorator) Network() network.Network {
	return hd.h.Network()
}

// Mux returns the Mux multiplexing incoming streams to protocol handlers
func (hd *hostDecorator) Mux() protocol.Switch {
	return hd.h.Mux()
}

// SetStreamHandler sets the protocol handler on the Host's Mux.
// This is equivalent to:
//   host.Mux().SetHandler(proto, handler)
// (Threadsafe)
func (hd *hostDecorator) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	hd.h.SetStreamHandler(pid, handler)
}

// SetStreamHandlerMatch sets the protocol handler on the Host's Mux
// using a matching function for protocol selection.
func (hd *hostDecorator) SetStreamHandlerMatch(pid protocol.ID, match func(string) bool, handler network.StreamHandler) {
	hd.h.SetStreamHandlerMatch(pid, match, handler)
}

// RemoveStreamHandler removes a handler on the mux that was set by
// SetStreamHandler
func (hd *hostDecorator) RemoveStreamHandler(pid protocol.ID) {
	hd.h.RemoveStreamHandler(pid)
}

// NewStream opens a new stream to given peer p, and writes a p2p/protocol
// header with given ProtocolID. If there is no connection to p, attempts
// to create one. If ProtocolID is "", writes no header.
// (Threadsafe)
func (hd *hostDecorator) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	return hd.h.NewStream(ctx, p, pids...)
}

// Close shuts down the host, its Network, and services.
func (hd *hostDecorator) Close() error {
	return hd.h.Close()
}

// ConnManager returns this hosts connection manager
func (hd *hostDecorator) ConnManager() connmgr.ConnManager {
	return hd.h.ConnManager()
}

// EventBus returns the hosts eventbus
func (hd *hostDecorator) EventBus() event.Bus {
	return hd.h.EventBus()
}
