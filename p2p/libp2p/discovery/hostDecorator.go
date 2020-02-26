package discovery

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	hardCPSLimitLow  = 2
	hardCPSLimitHigh = 100
)

type hostDecorator struct {
	host.Host
	acceptedChan chan struct{}
	timeout      time.Duration
}

// NewHostDecorator creates a new decorator around an existing host (h) that will only allow
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
	if cps < hardCPSLimitLow || cps > hardCPSLimitHigh {
		return nil, p2p.ErrInvalidValue
	}

	ret := &hostDecorator{
		Host:         h,
		acceptedChan: make(chan struct{}, cps),
		timeout:      timeout,
	}

	go ret.generate(ctx)

	return ret, nil
}

// generate new connect tokens
func (hd *hostDecorator) generate(ctx context.Context) {
	defer close(hd.acceptedChan)
	newTockenWait := time.Second / time.Duration(cap(hd.acceptedChan))
	for {
		select {
		case hd.acceptedChan <- struct{}{}:
		case <-ctx.Done():
			return
		}

		select {
		case <-time.After(newTockenWait):
		case <-ctx.Done():
			return
		}
	}
}

// Connect tries to connect connect to pi if the Connections Per Second (hd.cps)
// allows it.
func (hd *hostDecorator) Connect(ctx context.Context, pi peer.AddrInfo) error {

	select {
	case _, ok := <-hd.acceptedChan:
		if ok {
			return hd.Host.Connect(ctx, pi)
		}
		return p2p.ErrContextDone
	case <-ctx.Done():
		return p2p.ErrContextDone
	case <-time.After(hd.timeout):
		return p2p.ErrTimeout
	}
}
