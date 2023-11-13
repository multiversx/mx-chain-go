package components

import (
	"context"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

const routingTableRefreshInterval = time.Second * 300
const retryReconnectionToInitialPeerListInterval = time.Second * 5
const timeToConnect = time.Second * 2

type bootstrapper struct {
	h               host.Host
	initialPeerList []string
	kadDHT          *dht.IpfsDHT
	cancelFunc      func()
	chStart         chan struct{}
}

func newBootstrapper(h host.Host, initialPeerList []string, protocolID protocol.ID) (*bootstrapper, error) {
	instance := &bootstrapper{
		h:               h,
		chStart:         make(chan struct{}, 1),
		initialPeerList: initialPeerList,
	}

	var ctx context.Context
	ctx, instance.cancelFunc = context.WithCancel(context.Background())

	var err error
	instance.kadDHT, err = dht.New(
		ctx,
		h,
		dht.ProtocolPrefix(protocolID),
		dht.RoutingTableRefreshPeriod(routingTableRefreshInterval),
		dht.Mode(dht.ModeServer),
	)
	if err != nil {
		return nil, err
	}
	go instance.tryStart(ctx)

	return instance, nil
}

func (b *bootstrapper) tryStart(ctx context.Context) {
	<-b.chStart

	err := b.kadDHT.Bootstrap(ctx)
	if err != nil {
		log.Error(b.h.ID().String() + ": for kadDHT Bootstrap call: " + err.Error())
	}

	if len(b.initialPeerList) == 0 {
		log.Debug(b.h.ID().String() + ": no initial peer list provided")
		return
	}

	b.connectToInitialPeerList()
	for {
		select {
		case <-ctx.Done():
			log.Debug(b.h.ID().String() + ": stopping bootstrapper")
			return
		case <-time.After(retryReconnectionToInitialPeerListInterval):
			b.connectToInitialPeerList()
		}
	}
}

// connectToInitialPeerList will return the error if there sunt
func (b *bootstrapper) connectToInitialPeerList() {
	for _, address := range b.initialPeerList {
		pi, err := b.getAdrInfoFromP2pAddr(address)
		if err != nil {
			log.Warn("bootstrapper.connectToInitialPeerList - getAdrInfoFromP2pAddr", "error", err)
			continue
		}

		if b.h.Network().Connectedness(pi.ID) == network.Connected {
			continue
		}

		err = b.connectToHost(pi)
		if err != nil {
			log.Warn("can not connect to seeder", "address", address, "error", err.Error())
		} else {
			log.Debug("(re)connected to seeder", "address", address)
		}
	}
}

func (b *bootstrapper) connectToHost(pi peer.AddrInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeToConnect)
	defer cancel()

	return b.h.Connect(ctx, pi)
}

func (b *bootstrapper) getAdrInfoFromP2pAddr(address string) (peer.AddrInfo, error) {
	multiAddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	pi, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	return *pi, nil
}

func (b *bootstrapper) bootstrap() {
	// try a non-blocking write
	select {
	case b.chStart <- struct{}{}:
	default:
	}
}

func (b *bootstrapper) close() {
	b.cancelFunc()
}
