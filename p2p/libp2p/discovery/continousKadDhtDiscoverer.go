package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
)

// ContinousKadDhtDiscoverer is the kad-dht discovery type implementation
// This implementation does not support pausing and resuming of the discovery process
type ContinousKadDhtDiscoverer struct {
	mutKadDht     sync.RWMutex
	kadDHT        *dht.IpfsDHT
	refreshCancel context.CancelFunc

	contextProvider *libp2p.Libp2pContext

	refreshInterval  time.Duration
	randezVous       string
	initialPeersList []string
}

// NewContinousKadDhtDiscoverer creates a new kad-dht discovery type implementation
// initialPeersList can be nil or empty, no initial connection will be attempted, a warning message will appear
func NewContinousKadDhtDiscoverer(
	refreshInterval time.Duration,
	randezVous string,
	initialPeersList []string,
) *ContinousKadDhtDiscoverer {

	isListNilOrEmpty := len(initialPeersList) == 0
	if isListNilOrEmpty {
		log.Warn("nil or empty initial peers list provided to kad dht implementation. " +
			"No initial connection will be done")
	}

	return &ContinousKadDhtDiscoverer{
		refreshInterval:  refreshInterval,
		randezVous:       randezVous,
		initialPeersList: initialPeersList,
	}
}

// Bootstrap will start the bootstrapping new peers process
func (ckdd *ContinousKadDhtDiscoverer) Bootstrap() error {
	ckdd.mutKadDht.Lock()
	defer ckdd.mutKadDht.Unlock()

	if ckdd.kadDHT != nil {
		return p2p.ErrPeerDiscoveryProcessAlreadyStarted
	}
	if ckdd.contextProvider == nil {
		return p2p.ErrNilContextProvider
	}

	return ckdd.startDHT()
}

// UpdateRandezVous change the randezVous string, and restart the discovery with the new protocols
func (ckdd *ContinousKadDhtDiscoverer) UpdateRandezVous(s string) error {
	ckdd.mutKadDht.Lock()
	defer ckdd.mutKadDht.Unlock()

	if s == ckdd.randezVous {
		return nil
	}

	err := ckdd.stopDHT()
	if err != nil {
		log.Debug("error wile stopping kad-dht discovery, skip", "error", err)
	}

	ckdd.randezVous = s
	return ckdd.startDHT()
}

func (ckdd *ContinousKadDhtDiscoverer) protocols() []protocol.ID {
	return []protocol.ID{
		protocol.ID(fmt.Sprintf("%s/erd_%s", opts.ProtocolDHT, ckdd.randezVous)),
		protocol.ID(fmt.Sprintf("%s/erd", opts.ProtocolDHT)),
		//TODO: to be removed once the seed is updated
		opts.ProtocolDHT,
	}
}

func (ckdd *ContinousKadDhtDiscoverer) startDHT() error {
	ctx := ckdd.contextProvider.Context()
	h := ckdd.contextProvider.Host()

	ctxrun, cancel := context.WithCancel(ctx)
	hd, err := NewHostDecorator(h, ctxrun, 3, time.Second)
	if err != nil {
		cancel()
		return err
	}

	kademliaDHT, err := dht.New(ctx, hd, opts.Protocols(ckdd.protocols()...))
	if err != nil {
		cancel()
		return err
	}

	go ckdd.connectToInitialAndBootstrap(ctxrun)

	ckdd.kadDHT = kademliaDHT
	ckdd.refreshCancel = cancel
	return nil
}

func (ckdd *ContinousKadDhtDiscoverer) stopDHT() error {
	if ckdd.refreshCancel == nil {
		return nil
	}

	ckdd.refreshCancel()
	ckdd.refreshCancel = nil

	h := ckdd.contextProvider.Host()

	for _, p := range ckdd.protocols() {
		h.RemoveStreamHandler(p)
	}

	err := ckdd.kadDHT.Close()

	ckdd.kadDHT = nil

	return err
}

func (ckdd *ContinousKadDhtDiscoverer) connectToInitialAndBootstrap(ctx context.Context) {
	chanStartBootstrap := ckdd.connectToOnePeerFromInitialPeersList(
		ckdd.refreshInterval,
		ckdd.initialPeersList)

	//TODO(iulian) remove one nested go routine. Refactor the whole function
	go func() {
		<-chanStartBootstrap

		go func() {
			for {
				ckdd.mutKadDht.RLock()
				kadDht := ckdd.kadDHT
				ckdd.mutKadDht.RUnlock()

				var err = error(nil)
				if kadDht != nil {
					err = kadDht.Bootstrap(ctx)
				}
				if err == kbucket.ErrLookupFailure {
					<-ckdd.ReconnectToNetwork()
				}
				select {
				case <-time.After(ckdd.refreshInterval):
				case <-ctx.Done():
					return
				}
			}

		}()
	}()
}

func (ckdd *ContinousKadDhtDiscoverer) connectToOnePeerFromInitialPeersList(
	intervalBetweenAttempts time.Duration,
	initialPeersList []string,
) <-chan struct{} {

	chanDone := make(chan struct{}, 1)

	if len(initialPeersList) == 0 {
		chanDone <- struct{}{}
		return chanDone
	}

	go ckdd.tryConnectToSeeder(intervalBetweenAttempts, initialPeersList, chanDone)

	return chanDone
}

func (ckdd *ContinousKadDhtDiscoverer) tryConnectToSeeder(
	intervalBetweenAttempts time.Duration,
	initialPeersList []string,
	chanDone chan struct{},
) {

	startIndex := 0
	h := ckdd.contextProvider.Host()
	ctx := ckdd.contextProvider.Context()

	for {
		err := h.ConnectToPeer(ctx, initialPeersList[startIndex])

		if err != nil {
			//could not connect, wait and try next one
			startIndex++
			startIndex = startIndex % len(initialPeersList)
			select {
			case <-ctx.Done():
				break
			case <-time.After(intervalBetweenAttempts):
				continue
			}
		}
		break

	}
	chanDone <- struct{}{}
}

// Name returns the name of the kad dht peer discovery implementation
func (ckdd *ContinousKadDhtDiscoverer) Name() string {
	return kadDhtName
}

// ApplyContext sets the context in which this discoverer is to be run
func (ckdd *ContinousKadDhtDiscoverer) ApplyContext(ctxProvider p2p.ContextProvider) error {
	if check.IfNil(ctxProvider) {
		return p2p.ErrNilContextProvider
	}

	ctx, ok := ctxProvider.(*libp2p.Libp2pContext)
	if !ok {
		return p2p.ErrWrongContextProvider
	}

	ckdd.contextProvider = ctx
	return nil
}

// ReconnectToNetwork will try to connect to one peer from the initial peer list
func (ckdd *ContinousKadDhtDiscoverer) ReconnectToNetwork() <-chan struct{} {
	return ckdd.connectToOnePeerFromInitialPeersList(ckdd.refreshInterval, ckdd.initialPeersList)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ckdd *ContinousKadDhtDiscoverer) IsInterfaceNil() bool {
	return ckdd == nil
}
