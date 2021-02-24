package redundancy

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

var log = logger.GetOrCreate("redundancy")

// maxRoundsOfInactivityAccepted defines the maximum rounds of inactivity accepted, after which the main or lower
// level redundancy machines will be considered inactive
const maxRoundsOfInactivityAccepted = 5

type nodeRedundancy struct {
	redundancyLevel     int64
	lastRoundIndexCheck int64
	roundsOfInactivity  uint64
	mutNodeRedundancy   sync.RWMutex
	messenger           P2PMessenger
	observerPrivateKey  crypto.PrivateKey
}

// ArgNodeRedundancy represents the DTO structure used by the nodeRedundancy's constructor
type ArgNodeRedundancy struct {
	RedundancyLevel    int64
	Messenger          P2PMessenger
	ObserverPrivateKey crypto.PrivateKey
}

// NewNodeRedundancy creates a node redundancy object which implements NodeRedundancyHandler interface
func NewNodeRedundancy(arg ArgNodeRedundancy) (*nodeRedundancy, error) {
	if check.IfNil(arg.Messenger) {
		return nil, ErrNilMessenger
	}
	if check.IfNil(arg.ObserverPrivateKey) {
		return nil, ErrNilObserverPrivateKey
	}

	nr := &nodeRedundancy{
		redundancyLevel:    arg.RedundancyLevel,
		messenger:          arg.Messenger,
		observerPrivateKey: arg.ObserverPrivateKey,
	}

	return nr, nil
}

// IsRedundancyNode returns true if the current instance is used as a redundancy node
func (nr *nodeRedundancy) IsRedundancyNode() bool {
	return nr.redundancyLevel != 0
}

// IsMainMachineActive returns true if the main or lower level redundancy machines are active
func (nr *nodeRedundancy) IsMainMachineActive() bool {
	nr.mutNodeRedundancy.RLock()
	defer nr.mutNodeRedundancy.RUnlock()

	return nr.isMainMachineActive()
}

// AdjustInactivityIfNeeded increments rounds of inactivity for main or lower level redundancy machines if needed
func (nr *nodeRedundancy) AdjustInactivityIfNeeded(selfPubKey string, consensusPubKeys []string, roundIndex int64) {
	nr.mutNodeRedundancy.Lock()
	defer nr.mutNodeRedundancy.Unlock()

	if roundIndex <= nr.lastRoundIndexCheck {
		return
	}

	if nr.isMainMachineActive() {
		log.Debug("main or lower level redundancy machines are active", "node redundancy level", nr.redundancyLevel)
	} else {
		log.Warn("main or lower level redundancy machines are inactive", "node redundancy level", nr.redundancyLevel)
	}

	log.Debug("rounds of inactivity for main or lower level redundancy machines",
		"num", nr.roundsOfInactivity)

	for _, pubKey := range consensusPubKeys {
		if pubKey == selfPubKey {
			nr.roundsOfInactivity++
			break
		}
	}

	nr.lastRoundIndexCheck = roundIndex
}

// ResetInactivityIfNeeded resets rounds of inactivity for main or lower level redundancy machines if needed
func (nr *nodeRedundancy) ResetInactivityIfNeeded(selfPubKey string, consensusMsgPubKey string, consensusMsgPeerID core.PeerID) {
	if selfPubKey != consensusMsgPubKey {
		return
	}
	if consensusMsgPeerID == nr.messenger.ID() {
		return
	}

	nr.mutNodeRedundancy.Lock()
	nr.roundsOfInactivity = 0
	nr.mutNodeRedundancy.Unlock()
}

func (nr *nodeRedundancy) isMainMachineActive() bool {
	if nr.redundancyLevel < 0 {
		return true
	}

	return int64(nr.roundsOfInactivity) < maxRoundsOfInactivityAccepted*nr.redundancyLevel
}

// ObserverPrivateKey returns the stored private key by this instance. This key will be used whenever a new key,
// different from the main key is required. Example: sending anonymous heartbeat messages while the node is in backup mode.
func (nr *nodeRedundancy) ObserverPrivateKey() crypto.PrivateKey {
	return nr.observerPrivateKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (nr *nodeRedundancy) IsInterfaceNil() bool {
	return nr == nil
}
