package discovery

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
)

var peerDiscoveryTimeout = time.Duration(time.Second * 10)
var noOfQueries = 1

const kadDhtName = "kad-dht discovery"

// KadDhtDiscoverer is the kad-dht discovery type implementation
type KadDhtDiscoverer struct {
	mutKadDht sync.Mutex
	kadDHT    *dht.IpfsDHT

	contextProvider *libp2p.Libp2pContext

	refreshInterval  time.Duration
	randezVous       string
	initialPeersList []string
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

	ctx := kdd.contextProvider.Context()
	h := kdd.contextProvider.Host()

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		return err
	}

	go kdd.connectToInitialAndBootstrap()

	kdd.kadDHT = kademliaDHT
	return nil
}

func (kdd *KadDhtDiscoverer) connectToInitialAndBootstrap() {
	chanStartBootstrap := kdd.connectToOnePeerFromInitialPeersList(
		kdd.refreshInterval,
		kdd.initialPeersList)

	cfg := dht.BootstrapConfig{
		Period:  kdd.refreshInterval,
		Queries: noOfQueries,
		Timeout: time.Duration(peerDiscoveryTimeout),
	}

	ctx := kdd.contextProvider.Context()

	go func() {
		<-chanStartBootstrap

		kdd.mutKadDht.Lock()
		err := kdd.kadDHT.BootstrapWithConfig(ctx, cfg)
		kdd.mutKadDht.Unlock()
		if err != nil {
			log.Error(err.Error())
			return
		}
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
	if ctxProvider == nil {
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

// IsInterfaceNil returns true if there is no value under the interface
func (kdd *KadDhtDiscoverer) IsInterfaceNil() bool {
	if kdd == nil {
		return true
	}
	return false
}
