package discovery

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
	"github.com/libp2p/go-libp2p-kbucket"
)

const (
	initReconnectMul = 20
	kadDhtName       = "kad-dht discovery"
)

var log = logger.GetOrCreate("p2p/libp2p/kaddht")

// ArgKadDht represents the kad-dht config argument DTO
type ArgKadDht struct {
	PeersRefreshInterval time.Duration
	RandezVous           string
	InitialPeersList     []string
	BucketSize           uint32
	RoutingTableRefresh  time.Duration
}

// KadDhtDiscoverer is the kad-dht discovery type implementation
type KadDhtDiscoverer struct {
	mutKadDht            sync.Mutex
	kadDHT               *dht.IpfsDHT
	contextProvider      *libp2p.Libp2pContext
	peersRefreshInterval time.Duration
	randezVous           string
	initialPeersList     []string
	routingTableRefresh  time.Duration
	initConns            bool // Initiate new connections
	bucketSize           uint32
}

// NewKadDhtPeerDiscoverer creates a new kad-dht discovery type implementation
// initialPeersList can be nil or empty, no initial connection will be attempted, a warning message will appear
func NewKadDhtPeerDiscoverer(arg ArgKadDht) (*KadDhtDiscoverer, error) {
	if arg.PeersRefreshInterval < time.Second {
		return nil, fmt.Errorf("%w, PeersRefreshInterval should have been at least 1 second", p2p.ErrInvalidValue)
	}
	if arg.RoutingTableRefresh < time.Second {
		return nil, fmt.Errorf("%w, RoutingTableRefresh should have been at least 1 second", p2p.ErrInvalidValue)
	}
	isListNilOrEmpty := len(arg.InitialPeersList) == 0
	if isListNilOrEmpty {
		log.Warn("nil or empty initial peers list provided to kad dht implementation. " +
			"No initial connection will be done")
	}

	return &KadDhtDiscoverer{
		peersRefreshInterval: arg.PeersRefreshInterval,
		randezVous:           arg.RandezVous,
		initialPeersList:     arg.InitialPeersList,
		bucketSize:           arg.BucketSize,
		routingTableRefresh:  arg.RoutingTableRefresh,
		initConns:            true,
	}, nil
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

	ctx := kdd.contextProvider.Context()
	h := kdd.contextProvider.Host()

	defaultOptions := opts.Defaults
	customOptions := func(opt *opts.Options) error {
		err := defaultOptions(opt)
		if err != nil {
			return err
		}

		opt.BucketSize = int(kdd.bucketSize)
		opt.RoutingTable.RefreshPeriod = kdd.routingTableRefresh

		return nil
	}

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h, customOptions)
	if err != nil {
		return err
	}

	go kdd.connectToInitialAndBootstrap()

	kdd.kadDHT = kademliaDHT
	return nil
}

func (kdd *KadDhtDiscoverer) connectToInitialAndBootstrap() {
	chanStartBootstrap := kdd.connectToOnePeerFromInitialPeersList(
		kdd.peersRefreshInterval,
		kdd.initialPeersList,
	)

	ctx := kdd.contextProvider.Context()

	go func() {
		<-chanStartBootstrap

		kdd.mutKadDht.Lock()
		go func() {
			i := 1
			for {
				if kdd.initConns {
					err := kdd.kadDHT.Bootstrap(ctx)
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
				case <-time.After(kdd.peersRefreshInterval):
				case <-ctx.Done():
					return
				}
			}
		}()
		kdd.mutKadDht.Unlock()
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
				startIndex %= len(initialPeersList)

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
	return kdd.connectToOnePeerFromInitialPeersList(
		kdd.peersRefreshInterval,
		kdd.initialPeersList,
	)
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
	kdd.mutKadDht.Lock()
	defer kdd.mutKadDht.Unlock()
	return !kdd.initConns
}

// IsInterfaceNil returns true if there is no value under the interface
func (kdd *KadDhtDiscoverer) IsInterfaceNil() bool {
	return kdd == nil
}
