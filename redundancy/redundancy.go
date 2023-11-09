package redundancy

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/redundancy/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("redundancy")

type redundancyHandler interface {
	IncrementRoundsOfInactivity()
	ResetRoundsOfInactivity()
	IsMainMachineActive(maxRoundsOfInactivity int) bool
	RoundsOfInactivity() int
}

type nodeRedundancy struct {
	mutNodeRedundancy     sync.RWMutex
	lastRoundIndexCheck   int64
	handler               redundancyHandler
	maxRoundsOfInactivity int
	messenger             P2PMessenger
	observerPrivateKey    crypto.PrivateKey
}

// ArgNodeRedundancy represents the DTO structure used by the nodeRedundancy's constructor
type ArgNodeRedundancy struct {
	MaxRoundsOfInactivity int
	Messenger             P2PMessenger
	ObserverPrivateKey    crypto.PrivateKey
}

// NewNodeRedundancy creates a node redundancy object which implements NodeRedundancyHandler interface
func NewNodeRedundancy(arg ArgNodeRedundancy) (*nodeRedundancy, error) {
	if check.IfNil(arg.Messenger) {
		return nil, ErrNilMessenger
	}
	if check.IfNil(arg.ObserverPrivateKey) {
		return nil, ErrNilObserverPrivateKey
	}
	err := common.CheckMaxRoundsOfInactivity(arg.MaxRoundsOfInactivity)
	if err != nil {
		return nil, err
	}

	nr := &nodeRedundancy{
		handler:               common.NewRedundancyHandler(),
		maxRoundsOfInactivity: arg.MaxRoundsOfInactivity,
		messenger:             arg.Messenger,
		observerPrivateKey:    arg.ObserverPrivateKey,
	}

	return nr, nil
}

// IsRedundancyNode returns true if the current instance is used as a redundancy node
func (nr *nodeRedundancy) IsRedundancyNode() bool {
	return !common.IsMainNode(nr.maxRoundsOfInactivity)
}

// IsMainMachineActive returns true if the main or lower level redundancy machines are active
func (nr *nodeRedundancy) IsMainMachineActive() bool {
	nr.mutNodeRedundancy.RLock()
	defer nr.mutNodeRedundancy.RUnlock()

	return nr.handler.IsMainMachineActive(nr.maxRoundsOfInactivity)
}

// AdjustInactivityIfNeeded increments rounds of inactivity for main or lower level redundancy machines if needed
func (nr *nodeRedundancy) AdjustInactivityIfNeeded(selfPubKey string, consensusPubKeys []string, roundIndex int64) {
	nr.mutNodeRedundancy.Lock()
	defer nr.mutNodeRedundancy.Unlock()

	if roundIndex <= nr.lastRoundIndexCheck {
		return
	}

	if nr.handler.IsMainMachineActive(nr.maxRoundsOfInactivity) {
		log.Debug("main or lower level redundancy machines are active for single-key operation",
			"max rounds of inactivity", nr.maxRoundsOfInactivity,
			"current rounds of inactivity", nr.handler.RoundsOfInactivity())
	} else {
		log.Warn("main or lower level redundancy machines are inactive for single-key operation",
			"max rounds of inactivity", nr.maxRoundsOfInactivity,
			"current rounds of inactivity", nr.handler.RoundsOfInactivity())
	}

	for _, pubKey := range consensusPubKeys {
		if pubKey == selfPubKey {
			nr.handler.IncrementRoundsOfInactivity()
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
	nr.handler.ResetRoundsOfInactivity()
	nr.mutNodeRedundancy.Unlock()
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
