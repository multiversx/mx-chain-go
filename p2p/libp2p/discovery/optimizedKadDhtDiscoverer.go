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
const intervalBetweenSeedersReconnect = time.Minute * 5
const optimizedKadDhtName = "optimized kad-dht discovery"

type optimizedKadDhtDiscoverer struct {
	kadDHT               *dht.IpfsDHT
	peersRefreshInterval time.Duration
	protocolID           string
	initialPeersList     []string
	bucketSize           uint32
	routingTableRefresh  time.Duration
	hostConnManagement   *hostWithConnectionManagement
	sharder              Sharder

	status                   discovererStatus
	chanInit                 chan struct{}
	errChanInit              chan error
	chanConnectToSeeders     chan struct{}
	chanDoneConnectToSeeders chan struct{}
}

// NewOptimizedKadDhtDiscoverer creates an optimized kad-dht discovery type implementation
// initialPeersList can be nil or empty, no initial connection will be attempted, a warning message will appear
// This implementation uses a single process loop function able to carry multiple tasks synchronously
func NewOptimizedKadDhtDiscoverer(arg ArgKadDht) (*optimizedKadDhtDiscoverer, error) {
	sharder, err := prepareArguments(arg)
	if err != nil {
		return nil, err
	}

	okdd := &optimizedKadDhtDiscoverer{
		sharder:              sharder,
		peersRefreshInterval: arg.PeersRefreshInterval,
		protocolID:           arg.ProtocolID,
		initialPeersList:     arg.InitialPeersList,
		bucketSize:           arg.BucketSize,
		routingTableRefresh:  arg.RoutingTableRefresh,

		status:                   statNotInitialized,
		chanInit:                 make(chan struct{}),
		errChanInit:              make(chan error),
		chanConnectToSeeders:     make(chan struct{}),
		chanDoneConnectToSeeders: make(chan struct{}),
	}

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
	chTimeSeedersReconnect := time.After(intervalBetweenSeedersReconnect)
	chTimeFindPeers := time.After(okdd.peersRefreshInterval)

	for {
		select {
		case <-okdd.chanInit:
			err := okdd.init(ctx)
			okdd.errChanInit <- err
			okdd.connectToSeeders(ctx, false)
			okdd.findPeers(ctx)

		case <-chTimeSeedersReconnect:
			okdd.connectToSeeders(ctx, false)
			chTimeSeedersReconnect = time.After(intervalBetweenSeedersReconnect)

		case <-okdd.chanConnectToSeeders:
			okdd.connectToSeeders(ctx, true)
			//reset the automatic reconnect channel as we just tried to reconnect to the seeders
			chTimeSeedersReconnect = time.After(intervalBetweenSeedersReconnect)
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

func (okdd *optimizedKadDhtDiscoverer) init(ctx context.Context) error {
	if okdd.status != statNotInitialized {
		return p2p.ErrPeerDiscoveryProcessAlreadyStarted
	}

	var err error
	protocolID := protocol.ID(okdd.protocolID)
	okdd.kadDHT, err = dht.New(
		ctx,
		okdd.hostConnManagement,
		dht.ProtocolPrefix(protocolID),
		dht.RoutingTableRefreshPeriod(okdd.routingTableRefresh),
		dht.Mode(dht.ModeServer),
	)
	if err != nil {
		return err
	}

	okdd.status = statInitialized
	return nil
}

func (okdd *optimizedKadDhtDiscoverer) connectToSeeders(ctx context.Context, blocking bool) {
	if okdd.status != statInitialized {
		return
	}

	for {
		shouldStopReconnecting := okdd.tryToReconnectAtLeastToASeeder(ctx)
		if shouldStopReconnecting || !blocking {
			return
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
		case <-time.After(okdd.peersRefreshInterval):
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
