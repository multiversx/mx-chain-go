package keysManagement

import (
	"sync"

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
