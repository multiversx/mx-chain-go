package discovery

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

const mdnsName = "mdns peer discovery"

var log = logger.DefaultLogger()

// MdnsPeerDiscoverer is the mdns discovery type implementation
type MdnsPeerDiscoverer struct {
	contextProvider *libp2p.Libp2pContext

	mutMdns sync.Mutex
	mdns    discovery.Service

	refreshInterval time.Duration
	serviceTag      string
}

// NewMdnsPeerDiscoverer creates a new mdns discovery type implementation
// if serviceTag provided is empty, a default mdns serviceTag will be applied
func NewMdnsPeerDiscoverer(refreshInterval time.Duration, serviceTag string) *MdnsPeerDiscoverer {
	return &MdnsPeerDiscoverer{
		refreshInterval: refreshInterval,
		serviceTag:      serviceTag,
	}
}

// Bootstrap will start the bootstrapping new peers process
func (mpd *MdnsPeerDiscoverer) Bootstrap() error {
	mpd.mutMdns.Lock()
	defer mpd.mutMdns.Unlock()

	if mpd.mdns != nil {
		return p2p.ErrPeerDiscoveryProcessAlreadyStarted
	}

	if mpd.contextProvider == nil {
		return p2p.ErrNilContextProvider
	}

	m, err := discovery.NewMdnsService(
		mpd.contextProvider.Context(),
		mpd.contextProvider.Host(),
		mpd.refreshInterval,
		mpd.serviceTag)

	if err != nil {
		return err
	}

	m.RegisterNotifee(mpd)
	mpd.mdns = m
	return nil
}

// ApplyContext sets the context in which this discoverer is to be run
func (mpd *MdnsPeerDiscoverer) ApplyContext(ctxProvider p2p.ContextProvider) error {
	if ctxProvider == nil {
		return p2p.ErrNilContextProvider
	}

	ctx, ok := ctxProvider.(*libp2p.Libp2pContext)

	if !ok {
		return p2p.ErrWrongContextApplier
	}

	mpd.contextProvider = ctx
	return nil
}

// Name returns the name of the mdns peer discovery implementation
func (mpd *MdnsPeerDiscoverer) Name() string {
	return mdnsName
}

// HandlePeerFound updates the routing table with this new peer
func (mpd *MdnsPeerDiscoverer) HandlePeerFound(pi peer.AddrInfo) {
	if mpd.contextProvider == nil {
		return
	}

	h := mpd.contextProvider.Host()
	ctx := mpd.contextProvider.Context()

	peers := h.Peerstore().Peers()
	found := false

	for i := 0; i < len(peers); i++ {
		if peers[i] == pi.ID {
			found = true
			break
		}
	}

	if found {
		return
	}

	for i := 0; i < len(pi.Addrs); i++ {
		h.Peerstore().AddAddr(pi.ID, pi.Addrs[i], peerstore.PermanentAddrTTL)
	}

	//will try to connect for now as the connections and peer filtering is not done yet
	//TODO design a connection manager component
	go func() {
		err := h.Connect(ctx, pi)

		if err != nil {
			log.Debug(err.Error())
		}
	}()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mpd *MdnsPeerDiscoverer) IsInterfaceNil() bool {
	if mpd == nil {
		return true
	}
	return false
}
