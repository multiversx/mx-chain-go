package blackList

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/throttle/antiflood/blacklist")

const sizeBlacklistInfo = 4

type p2pBlackListProcessor struct {
	cacher              storage.Cacher
	peerBlacklistCacher process.PeerBlackListCacher
	selfPid             core.PeerID
	name                common.FloodPreventerType
	antifloodConfigs    common.AntifloodConfigsHandler
}

// NewP2PBlackListProcessor creates a new instance of p2pQuotaBlacklistProcessor able to determine
// a flooding peer and mark it accordingly
// TODO use argument on constructor
func NewP2PBlackListProcessor(
	cacher storage.Cacher,
	peerBlacklistCacher process.PeerBlackListCacher,
	name common.FloodPreventerType,
	selfPid core.PeerID,
	antifloodConfigs common.AntifloodConfigsHandler,
) (*p2pBlackListProcessor, error) {
	if check.IfNil(cacher) {
		return nil, fmt.Errorf("%w, NewP2PBlackListProcessor", process.ErrNilCacher)
	}
	if check.IfNil(peerBlacklistCacher) {
		return nil, fmt.Errorf("%w, NewP2PBlackListProcessor", process.ErrNilBlackListCacher)
	}
	if check.IfNil(antifloodConfigs) {
		return nil, fmt.Errorf("%w, NewP2PBlackListProcessor", process.ErrNilAntifloodConfigsHandler)
	}

	return &p2pBlackListProcessor{
		cacher:              cacher,
		peerBlacklistCacher: peerBlacklistCacher,
		selfPid:             selfPid,
		name:                name,
		antifloodConfigs:    antifloodConfigs,
	}, nil
}

// ResetStatistics checks if an identifier reached its maximum flooding rounds. If it did, it will remove its
// cached information and adds it to the black list handler
func (pbp *p2pBlackListProcessor) ResetStatistics() {
	keys := pbp.cacher.Keys()
	for _, key := range keys {
		val, ok := pbp.getFloodingValue(key)
		if !ok {
			pbp.cacher.Remove(key)
			continue
		}

		if val >= uint32(pbp.getNumFloodingRoundsVar())-1 { //-1 because the reset function is called before the AddQuota
			pbp.cacher.Remove(key)
			pid := core.PeerID(key)

			banDuration := pbp.getBadDuration()
			log.Debug("added new peer to black list",
				"peer ID", pid.Pretty(),
				"ban period", banDuration,
			)

			err := pbp.peerBlacklistCacher.Upsert(pid, banDuration)
			if err != nil {
				log.Warn("error adding peer id in peer ids cache", ""+
					"pid", p2p.PeerIdToShortString(pid),
					"error", err,
				)
			}
		}
	}
}

func (pbp *p2pBlackListProcessor) getBadDuration() time.Duration {
	currentConfig := pbp.antifloodConfigs.GetFloodPreventerConfigByType(pbp.name)
	return time.Duration(currentConfig.BlackList.PeerBanDurationInSeconds) * time.Second
}

func (pbp *p2pBlackListProcessor) getNumFloodingRoundsVar() uint32 {
	currentConfig := pbp.antifloodConfigs.GetFloodPreventerConfigByType(pbp.name)
	return currentConfig.BlackList.NumFloodingRounds
}

func (pbp *p2pBlackListProcessor) getFloodingValue(key []byte) (uint32, bool) {
	obj, ok := pbp.cacher.Peek(key)
	if !ok {
		return 0, false
	}

	val, ok := obj.(uint32)

	return val, ok
}

// AddQuota checks if the received quota for an identifier has exceeded the set thresholds
func (pbp *p2pBlackListProcessor) AddQuota(pid core.PeerID, numReceived uint32, sizeReceived uint64, _ uint32, _ uint64) {
	isFloodingPeer := numReceived >= pbp.getThresholdNumReceivedFlood() || sizeReceived >= pbp.getThresholdSizeReceivedFlood()
	if !isFloodingPeer {
		return
	}
	if pid == pbp.selfPid {
		log.Warn("current peer should have been blacklisted",
			"name", pbp.name,
			"total num messages", numReceived,
			"total size", sizeReceived,
		)
		return
	}

	pbp.incrementStatsFloodingPeer(pid)

}

func (pbp *p2pBlackListProcessor) getThresholdSizeReceivedFlood() uint64 {
	currentConfig := pbp.antifloodConfigs.GetFloodPreventerConfigByType(pbp.name)
	return currentConfig.BlackList.ThresholdSizePerInterval
}

func (pbp *p2pBlackListProcessor) getThresholdNumReceivedFlood() uint32 {
	currentConfig := pbp.antifloodConfigs.GetFloodPreventerConfigByType(pbp.name)
	return currentConfig.BlackList.ThresholdNumMessagesPerInterval
}

func (pbp *p2pBlackListProcessor) incrementStatsFloodingPeer(pid core.PeerID) {
	obj, ok := pbp.cacher.Get(pid.Bytes())
	if !ok {
		pbp.cacher.Put(pid.Bytes(), uint32(1), sizeBlacklistInfo)
		return
	}

	val, ok := obj.(uint32)
	if !ok {
		pbp.cacher.Put(pid.Bytes(), uint32(1), sizeBlacklistInfo)
		return
	}

	pbp.cacher.Put(pid.Bytes(), val+1, sizeBlacklistInfo)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pbp *p2pBlackListProcessor) IsInterfaceNil() bool {
	return pbp == nil
}
