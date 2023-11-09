package blacklist

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("consensus/blacklist")

var durationSweepP2PBlacklist = time.Second * 5

// PeerBlackListArgs defines the arguments needed for peer blacklist component
type PeerBlackListArgs struct {
	PeerCacher spos.PeerBlackListCacher
}

type peerBlacklist struct {
	peerCacher spos.PeerBlackListCacher
	cancel     func()
}

// NewPeerBlacklist creates a new instance of peer blacklist
func NewPeerBlacklist(args PeerBlackListArgs) (*peerBlacklist, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	pb := &peerBlacklist{
		peerCacher: args.PeerCacher,
	}

	pb.startSweepingTimeCache()

	return pb, nil
}

func checkArgs(args PeerBlackListArgs) error {
	if check.IfNil(args.PeerCacher) {
		return spos.ErrNilPeerBlacklistCacher
	}

	return nil
}

// IsPeerBlacklisted will check if specified peer is blacklisted
func (pb *peerBlacklist) IsPeerBlacklisted(peer core.PeerID) bool {
	return pb.peerCacher.Has(peer)
}

// BlacklistPeer will blacklist a peer for a certain amount of time
func (pb *peerBlacklist) BlacklistPeer(peer core.PeerID, duration time.Duration) {
	peerIsBlacklisted := pb.peerCacher.Has(peer)

	err := pb.peerCacher.Upsert(peer, duration)
	if err != nil {
		log.Warn("error adding in blacklist",
			"pid", peer.Pretty(),
			"time", duration,
			"error", "err",
		)
		return
	}

	if !peerIsBlacklisted {
		log.Debug("blacklisted peer",
			"pid", peer.Pretty(),
			"time", duration,
		)
	}
}

// startSweepingTimeCache will trigger the sweeping cache goroutine
func (pb *peerBlacklist) startSweepingTimeCache() {
	var ctx context.Context
	ctx, pb.cancel = context.WithCancel(context.Background())

	go pb.startSweepingPeerCache(ctx)
}

func (pb *peerBlacklist) startSweepingPeerCache(ctx context.Context) {
	timer := time.NewTimer(durationSweepP2PBlacklist)
	defer timer.Stop()

	for {
		timer.Reset(durationSweepP2PBlacklist)

		select {
		case <-ctx.Done():
			log.Debug("peerBlacklist's go routine is stopping...")
			return
		case <-timer.C:
		}

		pb.peerCacher.Sweep()
	}
}

// Close will close the goroutine
func (pb *peerBlacklist) Close() error {
	pb.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pb *peerBlacklist) IsInterfaceNil() bool {
	return pb == nil
}
