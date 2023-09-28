package keysManagement

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

type redundancyHandler interface {
	IsRedundancyNode(maxRoundsOfInactivity int) bool
	IncrementRoundsOfInactivity()
	ResetRoundsOfInactivity()
	IsMainMachineActive(maxRoundsOfInactivity int) bool
	RoundsOfInactivity() int
}

type peerInfo struct {
	pid                core.PeerID
	p2pPrivateKeyBytes []byte
	privateKey         crypto.PrivateKey
	machineID          string
	nodeName           string
	nodeIdentity       string

	mutChangeableData          sync.RWMutex
	handler                    redundancyHandler
	nextPeerAuthenticationTime time.Time
	isValidator                bool
}

func (pInfo *peerInfo) incrementRoundsWithoutReceivedMessages() {
	pInfo.mutChangeableData.Lock()
	pInfo.handler.IncrementRoundsOfInactivity()
	pInfo.mutChangeableData.Unlock()
}

func (pInfo *peerInfo) resetRoundsWithoutReceivedMessages() {
	pInfo.mutChangeableData.Lock()
	pInfo.handler.ResetRoundsOfInactivity()
	pInfo.mutChangeableData.Unlock()
}

func (pInfo *peerInfo) isNodeActiveOnMainMachine(maxRoundsOfInactivity int) bool {
	pInfo.mutChangeableData.RLock()
	defer pInfo.mutChangeableData.RUnlock()

	return pInfo.handler.IsMainMachineActive(maxRoundsOfInactivity)
}

func (pInfo *peerInfo) isNodeValidator() bool {
	pInfo.mutChangeableData.RLock()
	defer pInfo.mutChangeableData.RUnlock()

	return pInfo.isValidator
}

func (pInfo *peerInfo) setNodeValidator(value bool) {
	pInfo.mutChangeableData.Lock()
	defer pInfo.mutChangeableData.Unlock()

	pInfo.isValidator = value
}

func (pInfo *peerInfo) getNextPeerAuthenticationTime() time.Time {
	pInfo.mutChangeableData.RLock()
	defer pInfo.mutChangeableData.RUnlock()

	return pInfo.nextPeerAuthenticationTime
}

func (pInfo *peerInfo) setNextPeerAuthenticationTime(value time.Time) {
	pInfo.mutChangeableData.Lock()
	defer pInfo.mutChangeableData.Unlock()

	pInfo.nextPeerAuthenticationTime = value
}
