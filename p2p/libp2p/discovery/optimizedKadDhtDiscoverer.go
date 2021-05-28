package discovery

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

type discovererStatus string

const statNotInitialized discovererStatus = "not initialized"
const statInitialized discovererStatus = "initialized"
const minIntervalForSeedersReconnection = time.Second
const optimizedKadDhtName = "optimized kad-dht discovery"

type optimizedKadDhtDiscoverer struct {
	kadDHT                      KadDhtHandler
	peersRefreshInterval        time.Duration
	seedersReconnectionInterval time.Duration
	protocolID                  string
	initialPeersList            []string
	bucketSize                  uint32
	routingTableRefresh         time.Duration
	hostConnManagement          *hostWithConnectionManagement
	sharder                     Sharder
	status                      discovererStatus
	chanInit                    chan struct{}
	errChanInit                 chan error
	chanConnectToSeeders        chan struct{}
	createKadDhtHandler         func(ctx context.Context) (KadDhtHandler, error)
}

// NewOptimizedKadDhtDiscoverer creates an optimized kad-dht discovery type implementation
// initialPeersList can be nil or empty, no initial connection will be attempted, a warning message will appear
// This implementation uses a single process loop function able to carry multiple tasks synchronously
func NewOptimizedKadDhtDiscoverer(arg ArgKadDht) (*optimizedKadDhtDiscoverer, error) {
	sharder, err := prepareArguments(arg)
	if err != nil {
		return nil, err
	}

	if arg.SeedersReconnectionInterval < minIntervalForSeedersReconnection {
		return nil, p2p.ErrInvalidSeedersReconnectionInterval
	}

	sharder.SetSeeders(arg.InitialPeersList)

	okdd := &optimizedKadDhtDiscoverer{
		sharder:                     sharder,
		peersRefreshInterval:        arg.PeersRefreshInterval,
		seedersReconnectionInterval: arg.SeedersReconnectionInterval,
		protocolID:                  arg.ProtocolID,
		initialPeersList:            arg.InitialPeersList,
		bucketSize:                  arg.BucketSize,
		routingTableRefresh:         arg.RoutingTableRefresh,
		status:                      statNotInitialized,
		chanInit:                    make(chan struct{}),
		errChanInit:                 make(chan error),
		chanConnectToSeeders:        make(chan struct{}),
	}

	okdd.createKadDhtHandler = okdd.createKadDht
	okdd.hostConnManagement, err = NewHostWithConnectionManagement(arg.Host, okdd.sharder)
	if err != nil {
		return nil, err
	}

	go okdd.processLoop(arg.Context)

	return okdd, nil
}

// Bootstrap will start the bootstrapping new peers process
func (okdd *optimizedKadDhtDiscoverer) Bootstrap() error {
	okdd.chanInit <- struct{}{}
	return <-okdd.errChanInit
}

func (okdd *optimizedKadDhtDiscoverer) processLoop(ctx context.Context) {
	chTimeSeedersReconnect := time.After(okdd.seedersReconnectionInterval)
	chTimeFindPeers := time.After(okdd.peersRefreshInterval)

	for {
		select {
		case <-okdd.chanInit:
			okdd.processInit(ctx)

		case <-chTimeSeedersReconnect:
			chTimeSeedersReconnect = okdd.processSeedersReconnect(ctx)

		case <-okdd.chanConnectToSeeders:
			chTimeSeedersReconnect = okdd.processSeedersReconnect(ctx)

		case <-chTimeFindPeers:
			okdd.findPeers(ctx)
			chTimeFindPeers = time.After(okdd.peersRefreshInterval)

		case <-ctx.Done():
			log.Debug("closing the p2p bootstrapping process")

			okdd.finishMainLoopProcessing(ctx)
			return
		}
	}
}

func (okdd *optimizedKadDhtDiscoverer) processInit(ctx context.Context) {
	err := okdd.init(ctx)
	okdd.errChanInit <- err
	if err != nil {
		return
	}

	okdd.tryToReconnectAtLeastToASeeder(ctx)
	okdd.findPeers(ctx)
}

func (okdd *optimizedKadDhtDiscoverer) processSeedersReconnect(ctx context.Context) <-chan time.Time {
	isConnectedToSeeders := okdd.tryToReconnectAtLeastToASeeder(ctx)
	return okdd.createChTimeSeedersReconnect(isConnectedToSeeders)
}

func (okdd *optimizedKadDhtDiscoverer) finishMainLoopProcessing(ctx context.Context) {
	select {
	case okdd.errChanInit <- ctx.Err():
	default:
	}
}

func (okdd *optimizedKadDhtDiscoverer) createChTimeSeedersReconnect(isConnectedToSeeders bool) <-chan time.Time {
	if isConnectedToSeeders {
		//the reconnection will be done less often
		return time.After(okdd.seedersReconnectionInterval)
	}

	// no connection to seeders, let's try a little bit faster
	return time.After(okdd.peersRefreshInterval)
}

func (okdd *optimizedKadDhtDiscoverer) init(ctx context.Context) error {
	if okdd.status != statNotInitialized {
		return p2p.ErrPeerDiscoveryProcessAlreadyStarted
	}

	kadDhtHandler, err := okdd.createKadDhtHandler(ctx)
	if err != nil {
		return err
	}

	okdd.kadDHT = kadDhtHandler
	okdd.status = statInitialized

	return nil
}

func (okdd *optimizedKadDhtDiscoverer) createKadDht(ctx context.Context) (KadDhtHandler, error) {
	protocolID := protocol.ID(okdd.protocolID)
	return dht.New(
		ctx,
		okdd.hostConnManagement,
		dht.ProtocolPrefix(protocolID),
		dht.RoutingTableRefreshPeriod(okdd.routingTableRefresh),
		dht.Mode(dht.ModeServer),
	)
}

func (okdd *optimizedKadDhtDiscoverer) tryToReconnectAtLeastToASeeder(ctx context.Context) bool {
	if okdd.status != statInitialized {
		return false
	}

	if len(okdd.initialPeersList) == 0 {
		return true
	}

	connectedToOneSeeder := false
	for _, seederAddress := range okdd.initialPeersList {
		err := okdd.connectToSeeder(ctx, seederAddress)
		if err != nil {
			log.Debug("error connecting to seeder", "seeder", seederAddress, "error", err.Error())
		} else {
			connectedToOneSeeder = true
		}

		select {
		case <-ctx.Done():
			log.Debug("optimizedKadDhtDiscoverer.tryToReconnectAtLeastToASeeder",
				"num seeders", len(okdd.initialPeersList), "connected to a seeder", true, "context", "done")
			return true
		default:
		}
	}

	log.Debug("optimizedKadDhtDiscoverer.tryToReconnectAtLeastToASeeder",
		"num seeders", len(okdd.initialPeersList), "connected to a seeder", connectedToOneSeeder)

	return connectedToOneSeeder
}

func (okdd *optimizedKadDhtDiscoverer) connectToSeeder(ctx context.Context, seederAddress string) error {
	seederInfo, err := okdd.hostConnManagement.AddressToPeerInfo(seederAddress)
	if err != nil {
		return err
	}

	if okdd.hostConnManagement.IsConnected(*seederInfo) {
		return nil
	}

	return okdd.hostConnManagement.Connect(ctx, *seederInfo)
}

func (okdd *optimizedKadDhtDiscoverer) findPeers(ctx context.Context) {
	if okdd.status != statInitialized {
		return
	}

	err := okdd.kadDHT.Bootstrap(ctx)
	if err != nil {
		log.Debug("kad dht bootstrap", "error", err)
	}
}

// Name returns the name of the kad dht peer discovery implementation
func (okdd *optimizedKadDhtDiscoverer) Name() string {
	return optimizedKadDhtName
}

// ReconnectToNetwork will try to connect to one peer from the initial peer list
func (okdd *optimizedKadDhtDiscoverer) ReconnectToNetwork(_ context.Context) {
	select {
	case okdd.chanConnectToSeeders <- struct{}{}:
	default:
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (okdd *optimizedKadDhtDiscoverer) IsInterfaceNil() bool {
	return okdd == nil
}
