package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
)

var _ p2p.PeerDiscoverer = (*ContinuousKadDhtDiscoverer)(nil)
var _ p2p.Reconnecter = (*ContinuousKadDhtDiscoverer)(nil)

// ContinuousKadDhtDiscoverer is the kad-dht discovery type implementation
// This implementation does not support pausing and resuming of the discovery process
type ContinuousKadDhtDiscoverer struct {
	host          ConnectableHost
	context       context.Context
	mutKadDht     sync.RWMutex
	kadDHT        *dht.IpfsDHT
	refreshCancel context.CancelFunc

	peersRefreshInterval time.Duration
	randezVous           string
	initialPeersList     []string
	bucketSize           uint32
	routingTableRefresh  time.Duration
	hostConnManagement   *hostWithConnectionManagement
	sharder              Sharder
}

// NewContinuousKadDhtDiscoverer creates a new kad-dht discovery type implementation
// initialPeersList can be nil or empty, no initial connection will be attempted, a warning message will appear
func NewContinuousKadDhtDiscoverer(arg ArgKadDht) (*ContinuousKadDhtDiscoverer, error) {
	if check.IfNilReflect(arg.Context) {
		return nil, p2p.ErrNilContext
	}
	if check.IfNilReflect(arg.Host) {
		return nil, p2p.ErrNilHost
	}
	if check.IfNil(arg.KddSharder) {
		return nil, p2p.ErrNilSharder
	}
	sharder, ok := arg.KddSharder.(Sharder)
	if !ok {
		return nil, fmt.Errorf("%w for sharder: expected discovery.Sharder type of interface", p2p.ErrWrongTypeAssertion)
	}
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

	return &ContinuousKadDhtDiscoverer{
		context:              arg.Context,
		host:                 arg.Host,
		sharder:              sharder,
		peersRefreshInterval: arg.PeersRefreshInterval,
		randezVous:           arg.RandezVous,
		initialPeersList:     arg.InitialPeersList,
		bucketSize:           arg.BucketSize,
		routingTableRefresh:  arg.RoutingTableRefresh,
	}, nil
}

// Bootstrap will start the bootstrapping new peers process
func (ckdd *ContinuousKadDhtDiscoverer) Bootstrap() error {
	ckdd.mutKadDht.Lock()
	defer ckdd.mutKadDht.Unlock()

	if ckdd.kadDHT != nil {
		return p2p.ErrPeerDiscoveryProcessAlreadyStarted
	}

	return ckdd.startDHT()
}

// UpdateRandezVous change the randezVous string, and restart the discovery with the new protocols
func (ckdd *ContinuousKadDhtDiscoverer) UpdateRandezVous(s string) error {
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

func (ckdd *ContinuousKadDhtDiscoverer) protocols() []protocol.ID {
	return []protocol.ID{
		protocol.ID(fmt.Sprintf("%s/erd_%s", opts.ProtocolDHT, ckdd.randezVous)),
		protocol.ID(fmt.Sprintf("%s/erd", opts.ProtocolDHT)),
		opts.ProtocolDHT,
	}
}

func (ckdd *ContinuousKadDhtDiscoverer) startDHT() error {
	defaultOptions := opts.Defaults
	customOptions := func(opt *opts.Options) error {
		err := defaultOptions(opt)
		if err != nil {
			return err
		}

		return nil
	}

	ctxrun, cancel := context.WithCancel(ckdd.context)
	var err error
	ckdd.hostConnManagement, err = NewHostWithConnectionManagement(ckdd.host, ckdd.sharder)
	if err != nil {
		cancel()
		return err
	}

	kademliaDHT, err := dht.New(ckdd.context, ckdd.hostConnManagement, opts.Protocols(ckdd.protocols()...), customOptions)
	if err != nil {
		cancel()
		return err
	}

	go ckdd.connectToInitialAndBootstrap(ctxrun)

	ckdd.kadDHT = kademliaDHT
	ckdd.refreshCancel = cancel
	return nil
}

func (ckdd *ContinuousKadDhtDiscoverer) stopDHT() error {
	if ckdd.refreshCancel == nil {
		return nil
	}

	ckdd.refreshCancel()
	ckdd.refreshCancel = nil

	for _, p := range ckdd.protocols() {
		ckdd.host.RemoveStreamHandler(p)
	}

	err := ckdd.kadDHT.Close()

	ckdd.kadDHT = nil

	return err
}

func (ckdd *ContinuousKadDhtDiscoverer) connectToInitialAndBootstrap(ctx context.Context) {
	chanStartBootstrap := ckdd.connectToOnePeerFromInitialPeersList(
		ckdd.peersRefreshInterval,
		ckdd.initialPeersList,
	)

	go func() {
		<-chanStartBootstrap
		ckdd.bootstrap(ctx)
	}()
}

func (ckdd *ContinuousKadDhtDiscoverer) bootstrap(ctx context.Context) {
	cfg := dht.BootstrapConfig{
		Period:  ckdd.peersRefreshInterval,
		Queries: noOfQueries,
		Timeout: peerDiscoveryTimeout,
	}

	for {
		ckdd.mutKadDht.RLock()
		kadDht := ckdd.kadDHT
		ckdd.mutKadDht.RUnlock()

		shouldReconnect := kadDht != nil && kbucket.ErrLookupFailure == kadDht.BootstrapOnce(ctx, cfg)
		if shouldReconnect {
			<-ckdd.ReconnectToNetwork()
		}

		select {
		case <-time.After(ckdd.peersRefreshInterval):
		case <-ctx.Done():
			return
		}
	}
}

func (ckdd *ContinuousKadDhtDiscoverer) connectToOnePeerFromInitialPeersList(
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

func (ckdd *ContinuousKadDhtDiscoverer) tryConnectToSeeder(
	intervalBetweenAttempts time.Duration,
	initialPeersList []string,
	chanDone chan struct{},
) {

	startIndex := 0

	for {
		err := ckdd.host.ConnectToPeer(ckdd.context, initialPeersList[startIndex])
		if err != nil {
			//could not connect, wait and try next one
			startIndex++
			startIndex = startIndex % len(initialPeersList)
			select {
			case <-ckdd.context.Done():
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
func (ckdd *ContinuousKadDhtDiscoverer) Name() string {
	return kadDhtName
}

// ReconnectToNetwork will try to connect to one peer from the initial peer list
func (ckdd *ContinuousKadDhtDiscoverer) ReconnectToNetwork() <-chan struct{} {
	return ckdd.connectToOnePeerFromInitialPeersList(ckdd.peersRefreshInterval, ckdd.initialPeersList)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ckdd *ContinuousKadDhtDiscoverer) IsInterfaceNil() bool {
	return ckdd == nil
}
