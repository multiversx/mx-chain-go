package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	protocol "github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
)

const (
	initReconnectMul     = 20
	peerDiscoveryTimeout = 5 * time.Second
	noOfQueries          = 3

	kadDhtName = "kad-dht discovery"

	minWatchdogTimeout = time.Second
)

var log = logger.GetOrCreate("p2p/libp2p/kaddht")

// KadDhtDiscoverer is the kad-dht discovery type implementation
type KadDhtDiscoverer struct {
	mutKadDht     sync.RWMutex
	kadDHT        *dht.IpfsDHT
	refreshCancel context.CancelFunc

	contextProvider *libp2p.Libp2pContext

	refreshInterval  time.Duration
	randezVous       string
	initialPeersList []string
	initConns        bool // Initiate new connections
	watchdogKick     chan struct{}
	watchdogCancel   context.CancelFunc
}

// NewKadDhtPeerDiscoverer creates a new kad-dht discovery type implementation
// initialPeersList can be nil or empty, no initial connection will be attempted, a warning message will appear
func NewKadDhtPeerDiscoverer(
	refreshInterval time.Duration,
	randezVous string,
	initialPeersList []string) *KadDhtDiscoverer {

	isListNilOrEmpty := initialPeersList == nil || len(initialPeersList) == 0

	if isListNilOrEmpty {
		log.Warn("nil or empty initial peers list provided to kad dht implementation. " +
			"No initial connection will be done")
	}

	return &KadDhtDiscoverer{
		refreshInterval:  refreshInterval,
		randezVous:       randezVous,
		initialPeersList: initialPeersList,
		initConns:        true,
	}
}

// Bootstrap will start the bootstrapping new peers process
func (kdd *KadDhtDiscoverer) Bootstrap() error {
	kdd.mutKadDht.Lock()
	defer kdd.mutKadDht.Unlock()

	if kdd.kadDHT != nil {
		return p2p.ErrPeerDiscoveryProcessAlreadyStarted
	}

	if kdd.contextProvider == nil {
		return p2p.ErrNilContextProvider
	}

	return kdd.startDHT()
}

// UpdateRandezVous change the randezVous string, and restart the discovery with the new protocols
func (kdd *KadDhtDiscoverer) UpdateRandezVous(s string) error {
	kdd.mutKadDht.Lock()
	defer kdd.mutKadDht.Unlock()

	if s == kdd.randezVous {
		return nil
	}

	err := kdd.stopDHT()
	if err != nil {
		log.Warn("Error wile stopDHT, ignore", "err", err)
	}
	kdd.randezVous = s
	return kdd.startDHT()
}

func (kdd *KadDhtDiscoverer) protocols() []protocol.ID {
	return []protocol.ID{
		protocol.ID(fmt.Sprintf("%s/erd_%s", opts.ProtocolDHT, kdd.randezVous)),
		protocol.ID(fmt.Sprintf("%s/erd", opts.ProtocolDHT)),
	}
}

func (kdd *KadDhtDiscoverer) startDHT() error {
	ctx := kdd.contextProvider.Context()
	h := kdd.contextProvider.Host()

	kademliaDHT, err := dht.New(ctx, h, opts.Protocols(kdd.protocols()...))
	if err != nil {
		return err
	}

	ctxrun, cancel := context.WithCancel(ctx)
	go kdd.connectToInitialAndBootstrap(ctxrun)

	kdd.kadDHT = kademliaDHT
	kdd.refreshCancel = cancel
	return nil
}

func (kdd *KadDhtDiscoverer) stopDHT() error {
	if kdd.refreshCancel == nil {
		return nil
	}

	kdd.refreshCancel()
	kdd.refreshCancel = nil

	h := kdd.contextProvider.Host()

	for _, p := range kdd.protocols() {
		h.RemoveStreamHandler(p)
	}

	err := kdd.kadDHT.Close()

	kdd.kadDHT = nil

	return err
}

func (kdd *KadDhtDiscoverer) connectToInitialAndBootstrap(ctx context.Context) {
	chanStartBootstrap := kdd.connectToOnePeerFromInitialPeersList(
		kdd.refreshInterval,
		kdd.initialPeersList)

	cfg := dht.BootstrapConfig{
		Period:  kdd.refreshInterval,
		Queries: noOfQueries,
		Timeout: peerDiscoveryTimeout,
	}

	go func() {
		<-chanStartBootstrap

		go func() {
			i := 1
			for {
				if kdd.initConns {
					var err error = nil
					kdd.mutKadDht.RLock()
					if kdd.kadDHT != nil {
						err = kdd.kadDHT.BootstrapOnce(ctx, cfg)
					}
					kdd.mutKadDht.RUnlock()
					if err == kbucket.ErrLookupFailure {
						<-kdd.ReconnectToNetwork()
					}
					i = 1
				} else {
					i++
					if (i % initReconnectMul) == 0 {
						<-kdd.ReconnectToNetwork()
						i = 1
					}
				}
				select {
				case <-time.After(cfg.Period):
				case <-ctx.Done():
					return
				}
			}

		}()
	}()
}

func (kdd *KadDhtDiscoverer) connectToOnePeerFromInitialPeersList(
	intervalBetweenAttempts time.Duration,
	initialPeersList []string) <-chan struct{} {

	h := kdd.contextProvider.Host()
	ctx := kdd.contextProvider.Context()

	chanDone := make(chan struct{}, 1)

	if initialPeersList == nil {
		chanDone <- struct{}{}
		return chanDone
	}

	if len(initialPeersList) == 0 {
		chanDone <- struct{}{}
		return chanDone
	}

	go func() {
		startIndex := 0

		for {
			err := h.ConnectToPeer(ctx, initialPeersList[startIndex])

			if err != nil {
				//could not connect, wait and try next one
				startIndex++
				startIndex = startIndex % len(initialPeersList)

				time.Sleep(intervalBetweenAttempts)

				continue
			}

			chanDone <- struct{}{}
			return
		}
	}()

	return chanDone
}

// Name returns the name of the kad dht peer discovery implementation
func (kdd *KadDhtDiscoverer) Name() string {
	return kadDhtName
}

// ApplyContext sets the context in which this discoverer is to be run
func (kdd *KadDhtDiscoverer) ApplyContext(ctxProvider p2p.ContextProvider) error {
	if ctxProvider == nil || ctxProvider.IsInterfaceNil() {
		return p2p.ErrNilContextProvider
	}

	ctx, ok := ctxProvider.(*libp2p.Libp2pContext)

	if !ok {
		return p2p.ErrWrongContextApplier
	}

	kdd.contextProvider = ctx
	return nil
}

// ReconnectToNetwork will try to connect to one peer from the initial peer list
func (kdd *KadDhtDiscoverer) ReconnectToNetwork() <-chan struct{} {
	return kdd.connectToOnePeerFromInitialPeersList(kdd.refreshInterval, kdd.initialPeersList)
}

// Pause will suspend the discovery process
func (kdd *KadDhtDiscoverer) Pause() {
	kdd.mutKadDht.Lock()
	defer kdd.mutKadDht.Unlock()
	kdd.initConns = false
}

// Resume will resume the discovery process
func (kdd *KadDhtDiscoverer) Resume() {
	kdd.mutKadDht.Lock()
	defer kdd.mutKadDht.Unlock()
	kdd.initConns = true
}

// IsDiscoveryPaused will return true if the discoverer is initiating connections
func (kdd *KadDhtDiscoverer) IsDiscoveryPaused() bool {
	kdd.mutKadDht.RLock()
	defer kdd.mutKadDht.RUnlock()
	return !kdd.initConns
}

// IsInterfaceNil returns true if there is no value under the interface
func (kdd *KadDhtDiscoverer) IsInterfaceNil() bool {
	if kdd == nil {
		return true
	}
	return false
}

// StartWatchdog start the watchdog
func (kdd *KadDhtDiscoverer) StartWatchdog(timeout time.Duration) error {
	kdd.mutKadDht.Lock()
	defer kdd.mutKadDht.Unlock()

	if kdd.contextProvider == nil {
		return p2p.ErrNilContextProvider
	}

	if kdd.watchdogKick != nil {
		return p2p.ErrWatchdogAlreadyStarted
	}

	if timeout < minWatchdogTimeout {
		return p2p.ErrInvalidDurationProvided
	}

	kdd.watchdogKick = make(chan struct{})
	ctx := kdd.contextProvider.Context()
	wdCtx, wdCancel := context.WithCancel(ctx)
	go func(kick <-chan struct{}) {
		for {
			select {
			case <-time.After(timeout):
				kdd.Resume()
			case <-wdCtx.Done():
				return
			case <-kick:
			}
		}
	}(kdd.watchdogKick)

	kdd.watchdogCancel = wdCancel
	return nil
}

// StopWatchdog stops the discovery watchdog
func (kdd *KadDhtDiscoverer) StopWatchdog() error {
	kdd.mutKadDht.Lock()
	defer kdd.mutKadDht.Unlock()

	if kdd.watchdogCancel == nil {
		return p2p.ErrWatchdogNotStarted
	}

	kdd.watchdogCancel()
	kdd.watchdogCancel = nil

	close(kdd.watchdogKick)
	kdd.watchdogKick = nil
	return nil
}

// KickWatchdog extends the discovery resume timeout
func (kdd *KadDhtDiscoverer) KickWatchdog() error {
	kdd.mutKadDht.RLock()
	defer kdd.mutKadDht.RUnlock()

	if kdd.watchdogKick == nil {
		return p2p.ErrWatchdogNotStarted
	}

	select {
	case kdd.watchdogKick <- struct{}{}:
	default:
	}
	return nil
}
