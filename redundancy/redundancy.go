package redundancy

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go"
	cmn "github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/configs/dto"
	"github.com/multiversx/mx-chain-go/process"
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
	messenger             P2PMessenger
	observerPrivateKey    crypto.PrivateKey
	processConfigsHandler cmn.ProcessConfigsHandler
	redundancyLevel       int
}

// ArgNodeRedundancy represents the DTO structure used by the nodeRedundancy's constructor
type ArgNodeRedundancy struct {
	RedundancyLevel       int
	ProcessConfigsHandler cmn.ProcessConfigsHandler
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
	if check.IfNil(arg.ProcessConfigsHandler) {
		return nil, process.ErrNilProcessConfigsHandler
	}

	nr := &nodeRedundancy{
		handler:               common.NewRedundancyHandler(),
		messenger:             arg.Messenger,
		observerPrivateKey:    arg.ObserverPrivateKey,
		redundancyLevel:       arg.RedundancyLevel,
		processConfigsHandler: arg.ProcessConfigsHandler,
	}

	err := common.CheckMaxRoundsOfInactivity(nr.calcMaxRoundsOfInactivity())
	if err != nil {
		return nil, err
	}

	return nr, nil
}

func (nr *nodeRedundancy) calcMaxRoundsOfInactivity() int {
	return nr.redundancyLevel * int(nr.processConfigsHandler.GetValue(dto.MaxRoundsOfInactivityAccepted))
}

// IsRedundancyNode returns true if the current instance is used as a redundancy node
func (nr *nodeRedundancy) IsRedundancyNode() bool {
	return !common.IsMainNode(nr.calcMaxRoundsOfInactivity())
}

// IsMainMachineActive returns true if the main or lower level redundancy machines are active
func (nr *nodeRedundancy) IsMainMachineActive() bool {
	nr.mutNodeRedundancy.RLock()
	defer nr.mutNodeRedundancy.RUnlock()

	return nr.handler.IsMainMachineActive(nr.calcMaxRoundsOfInactivity())
}

// AdjustInactivityIfNeeded increments rounds of inactivity for main or lower level redundancy machines if needed
func (nr *nodeRedundancy) AdjustInactivityIfNeeded(selfPubKey string, consensusPubKeys []string, roundIndex int64) {
	nr.mutNodeRedundancy.Lock()
	defer nr.mutNodeRedundancy.Unlock()

	if roundIndex <= nr.lastRoundIndexCheck {
		return
	}

	maxRoundsOfInactivity := nr.calcMaxRoundsOfInactivity()
	if nr.handler.IsMainMachineActive(maxRoundsOfInactivity) {
		log.Debug("main or lower level redundancy machines are active for single-key operation",
			"max rounds of inactivity", maxRoundsOfInactivity,
			"current rounds of inactivity", nr.handler.RoundsOfInactivity())
	} else {
		log.Warn("main or lower level redundancy machines are inactive for single-key operation",
			"max rounds of inactivity", maxRoundsOfInactivity,
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
