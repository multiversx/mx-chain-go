package components

import (
	"context"
	"fmt"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
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
		chStart:         make(chan struct{}),
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
	defer func() {
		err := b.kadDHT.Bootstrap(ctx)
		if err != nil {
			log.Error(b.h.ID().String() + ": for kadDHT Bootstrap call: " + err.Error())
		}
	}()

	if len(b.initialPeerList) == 0 {
		log.Debug(b.h.ID().String() + ": no initial peer list provided")
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Info(b.h.ID().String() + ": bootstrapper try start ended early")
			return
		case <-b.chStart:
			err := b.connectToInitialPeerList()
			if err == nil {
				log.Info(b.h.ID().String() + ": CONNECTED to the network")
				return
			}

			time.Sleep(retryReconnectionToInitialPeerListInterval)
		}
	}
}

// connectToInitialPeerList will return the error if there sunt
func (b *bootstrapper) connectToInitialPeerList() error {
	for _, address := range b.initialPeerList {
		err := b.connectToHost(address)
		if err != nil {
			log.Debug(b.h.ID().String() + ": while attempting initial connection to " + address + ": " + err.Error())
			continue
		}

		return nil
	}

	return fmt.Errorf("unable to connect to initial peers")
}

func (b *bootstrapper) connectToHost(address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeToConnect)
	defer cancel()

	multiAddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return err
	}

	pi, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		return err
	}

	return b.h.Connect(ctx, *pi)
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
