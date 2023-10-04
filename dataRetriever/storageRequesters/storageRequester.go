package storagerequesters

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

type storageRequester struct {
	messenger                dataRetriever.MessageHandler
	responseTopicName        string
	manualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	chanGracefullyClose      chan endProcess.ArgEndProcess
	mutSignaled              sync.Mutex
	signaled                 bool
	delayBeforeGracefulClose time.Duration
}

// SetDebugHandler returns nil, no debugging associated to this implementation
func (sr *storageRequester) SetDebugHandler(_ dataRetriever.DebugHandler) error {
	return nil
}

// SetNumPeersToQuery does nothing
func (sr *storageRequester) SetNumPeersToQuery(_ int, _ int) {
}

// NumPeersToQuery returns (0, 0) tuple as it won't request any connected peer
func (sr *storageRequester) NumPeersToQuery() (int, int) {
	return 0, 0
}

func (sr *storageRequester) sendToSelf(buffToSend []byte) error {
	return sr.messenger.SendToConnectedPeer(sr.responseTopicName, buffToSend, sr.messenger.ID())
}

func (sr *storageRequester) signalGracefullyClose() {
	sr.mutSignaled.Lock()
	defer sr.mutSignaled.Unlock()

	if sr.signaled {
		return
	}

	sr.signaled = true
	go sr.asyncCallGracefullyClose()
}

func (sr *storageRequester) asyncCallGracefullyClose() {
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
		log.Debug("storageRequester.RequestDataFromHash: could not wrote on the end chan")
	}
}
