package keysManagement

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

type peerInfo struct {
	pid                core.PeerID
	p2pPrivateKeyBytes []byte
	privateKey         crypto.PrivateKey
	machineID          string
	nodeName           string
	nodeIdentity       string

	mutPeerAuthenticationData  sync.RWMutex
	isValidator                bool
	nextPeerAuthenticationTime time.Time

	mutChangeableData             sync.RWMutex
	roundsWithoutReceivedMessages int
}

func (pInfo *peerInfo) incrementRoundsWithoutReceivedMessages() {
	pInfo.mutChangeableData.Lock()
	pInfo.roundsWithoutReceivedMessages++
	pInfo.mutChangeableData.Unlock()
}

func (pInfo *peerInfo) resetRoundsWithoutReceivedMessages() {
	pInfo.mutChangeableData.Lock()
	pInfo.roundsWithoutReceivedMessages = 0
	pInfo.mutChangeableData.Unlock()
}

func (pInfo *peerInfo) isNodeActiveOnMainMachine(maxRoundsWithoutReceivedMessages int) bool {
	pInfo.mutChangeableData.RLock()
	defer pInfo.mutChangeableData.RUnlock()

	return pInfo.roundsWithoutReceivedMessages < maxRoundsWithoutReceivedMessages
}

func (pInfo *peerInfo) isNodeValidator() bool {
	pInfo.mutPeerAuthenticationData.RLock()
	defer pInfo.mutPeerAuthenticationData.RUnlock()

	return pInfo.isValidator
}

func (pInfo *peerInfo) setNodeValidator(value bool) {
	pInfo.mutPeerAuthenticationData.Lock()
	defer pInfo.mutPeerAuthenticationData.Unlock()

	pInfo.isValidator = value
}

func (pInfo *peerInfo) getNextPeerAuthenticationTime() time.Time {
	pInfo.mutPeerAuthenticationData.RLock()
	defer pInfo.mutPeerAuthenticationData.RUnlock()

	return pInfo.nextPeerAuthenticationTime
}

func (pInfo *peerInfo) setNextPeerAuthenticationTime(value time.Time) {
	pInfo.mutPeerAuthenticationData.Lock()
	defer pInfo.mutPeerAuthenticationData.Unlock()

	pInfo.nextPeerAuthenticationTime = value
}
