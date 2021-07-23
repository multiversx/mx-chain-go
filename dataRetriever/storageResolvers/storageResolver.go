package storageResolvers

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type storageResolver struct {
	messenger                dataRetriever.MessageHandler
	responseTopicName        string
	manualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	chanGracefullyClose      chan endProcess.ArgEndProcess
	mutSignaled              sync.Mutex
	signaled                 bool
	delayBeforeGracefulClose time.Duration
}

// ProcessReceivedMessage does nothing, won't be able to process network requests
func (sr *storageResolver) ProcessReceivedMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// SetResolverDebugHandler returns nil, no debugging associated to this implementation
func (sr *storageResolver) SetResolverDebugHandler(_ dataRetriever.ResolverDebugHandler) error {
	return nil
}

// SetNumPeersToQuery does nothing
func (sr *storageResolver) SetNumPeersToQuery(_ int, _ int) {
}

// NumPeersToQuery returns (0, 0) tuple as it won't request any connected peer
func (sr *storageResolver) NumPeersToQuery() (int, int) {
	return 0, 0
}

func (sr *storageResolver) sendToSelf(buffToSend []byte) error {
	return sr.messenger.SendToConnectedPeer(sr.responseTopicName, buffToSend, sr.messenger.ID())
}

func (sr *storageResolver) signalGracefullyClose() {
	sr.mutSignaled.Lock()
	defer sr.mutSignaled.Unlock()

	if sr.signaled {
		return
	}

	sr.signaled = true
	go sr.asyncCallGracefullyClose()
}

func (sr *storageResolver) asyncCallGracefullyClose() {
	crtEpoch := sr.manualEpochStartNotifier.CurrentEpoch()

	argEndProcess := endProcess.ArgEndProcess{
		Reason: common.ImportComplete,
		Description: fmt.Sprintf("import ended because data from epochs %d or %d does not exist",
			crtEpoch-1, crtEpoch),
	}

	time.Sleep(sr.delayBeforeGracefulClose)

	select {
	case sr.chanGracefullyClose <- argEndProcess:
	default:
		log.Debug("storageResolver.RequestDataFromHash: could not wrote on the end chan")
	}
}
