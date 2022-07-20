package blacklist

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
)

var log = logger.GetOrCreate("consensus/blacklist")

var durationSweepP2PBlacklist = time.Second * 5

type PeerBlackListArgs struct {
	PeerCacher spos.PeerBlackListCacher
}

type peerBlacklist struct {
	peerCacher spos.PeerBlackListCacher
	cancel     func()
}

func NewPeerBlacklist(args PeerBlackListArgs) (*peerBlacklist, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &peerBlacklist{
		peerCacher: args.PeerCacher,
	}, nil
}

func checkArgs(args PeerBlackListArgs) error {
	if check.IfNil(args.PeerCacher) {
		return spos.ErrNilPeerBlacklistCacher
	}

	return nil
}

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

func (pb *peerBlacklist) StartSweepingTimeCache() {
	var ctx context.Context
	ctx, pb.cancel = context.WithCancel(context.Background())

	go pb.startSweepingPeerCache(ctx)
}

func (pb *peerBlacklist) startSweepingPeerCache(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("peerBlacklist's go routine is stopping...")
			return
		case <-time.After(durationSweepP2PBlacklist):
		}

		pb.peerCacher.Sweep()
	}
}

func (pb *peerBlacklist) Close() error {
	pb.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pb *peerBlacklist) IsInterfaceNil() bool {
	return pb == nil
}
