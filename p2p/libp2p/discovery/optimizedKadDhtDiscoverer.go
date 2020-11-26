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
	chanDoneConnectToSeeders    chan struct{}
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
		chanDoneConnectToSeeders:    make(chan struct{}),
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
			err := okdd.init(ctx)
			okdd.errChanInit <- err
			okdd.connectToSeeders(ctx, false)
			okdd.findPeers(ctx)

		case <-chTimeSeedersReconnect:
			isConnectedToSeeders := okdd.connectToSeeders(ctx, false)
			chTimeSeedersReconnect = okdd.createChTimeSeedersReconnect(isConnectedToSeeders)

		case <-okdd.chanConnectToSeeders:
			isConnectedToSeeders := okdd.connectToSeeders(ctx, true)
			chTimeSeedersReconnect = okdd.createChTimeSeedersReconnect(isConnectedToSeeders)
			okdd.chanDoneConnectToSeeders <- struct{}{}

		case <-chTimeFindPeers:
			okdd.findPeers(ctx)
			chTimeFindPeers = time.After(okdd.peersRefreshInterval)

		case <-ctx.Done():
			log.Debug("closing the p2p bootstrapping process")
			return
		}
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

func (okdd *optimizedKadDhtDiscoverer) connectToSeeders(ctx context.Context, blocking bool) bool {
	if okdd.status != statInitialized {
		return false
	}

	for {
		connectedToASeeder := okdd.tryToReconnectAtLeastToASeeder(ctx)
		log.Debug("optimizedKadDhtDiscoverer.tryToReconnectAtLeastToASeeder",
			"num seeders", len(okdd.initialPeersList), "connected to a seeder", connectedToASeeder)
		if connectedToASeeder || !blocking {
			return connectedToASeeder
		}

		select {
		case <-ctx.Done():
			return true
		case <-time.After(okdd.peersRefreshInterval):
			//we need a bit o delay before we retry connecting to seeder peers as to not solely rely on the libp2p back-off mechanism
		}
	}
}

func (okdd *optimizedKadDhtDiscoverer) tryToReconnectAtLeastToASeeder(ctx context.Context) bool {
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
			return true
		default:
		}
	}

	return connectedToOneSeeder
}

func (okdd *optimizedKadDhtDiscoverer) connectToSeeder(ctx context.Context, seederAddress string) error {
	seederInfo, err := okdd.hostConnManagement.AddressToPeerInfo(seederAddress)
	if err != nil {
		return err
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
func (okdd *optimizedKadDhtDiscoverer) ReconnectToNetwork() {
	okdd.chanConnectToSeeders <- struct{}{}
	<-okdd.chanDoneConnectToSeeders
}

// IsInterfaceNil returns true if there is no value under the interface
func (okdd *optimizedKadDhtDiscoverer) IsInterfaceNil() bool {
	return okdd == nil
}
