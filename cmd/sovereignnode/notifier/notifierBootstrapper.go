package notifier

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
	notifierProcess "github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

var log = logger.GetOrCreate("notifier-bootstrap")

type ArgsNotifierBootstrapper struct {
	IncomingHeaderHandler process.IncomingHeaderSubscriber
	SovereignNotifier     notifierProcess.SovereignNotifier
	ForkDetector          process.ForkDetector
	Bootstrapper          process.Bootstrapper
	RoundDuration         uint64
}

type notifierBootstrapper struct {
	incomingHeaderHandler process.IncomingHeaderSubscriber
	sovereignNotifier     notifierProcess.SovereignNotifier
	forkDetector          process.ForkDetector

	nodeSyncedChan chan bool
	cancelFunc     func()
	roundDuration  uint64
}

func NewNotifierBootstrapper(args ArgsNotifierBootstrapper) (*notifierBootstrapper, error) {
	nb := &notifierBootstrapper{
		incomingHeaderHandler: args.IncomingHeaderHandler,
		sovereignNotifier:     args.SovereignNotifier,
		forkDetector:          args.ForkDetector,
		nodeSyncedChan:        make(chan bool, 1),
		cancelFunc:            nil,
		roundDuration:         args.RoundDuration,
	}

	args.Bootstrapper.AddSyncStateListener(nb.receivedSyncState)

	return nb, nil
}

func (nb *notifierBootstrapper) receivedSyncState(isNodeSynchronized bool) {
	if isNodeSynchronized && nb.forkDetector.GetHighestFinalBlockNonce() != 0 {
		select {
		case nb.nodeSyncedChan <- true:
		default:
		}
	}
}

func (nb *notifierBootstrapper) Start() {
	var ctx context.Context
	ctx, nb.cancelFunc = context.WithCancel(context.Background())
	go nb.checkNodeState(ctx)
}

func (nb *notifierBootstrapper) checkNodeState(ctx context.Context) {
	timeToWaitReSync := (process.MaxRoundsWithoutNewBlockReceived + 1) * nb.roundDuration
	ticker := time.NewTicker(time.Duration(timeToWaitReSync) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("worker's go routine is stopping...")
			return
		case <-nb.nodeSyncedChan:
			err := nb.sovereignNotifier.RegisterHandler(nb.incomingHeaderHandler)
			if err != nil {
				log.Error("notifierBootstrapper: sovereignNotifier.RegisterHandler", "err", err)
			}
			log.Debug("notifierBootstrapper.checkNodeState", "is node synced", true)
			return
		case <-ticker.C:
			log.Debug("notifierBootstrapper.checkNodeState", "is node synced", false)
		}

	}
}

func (nb *notifierBootstrapper) Close() error {
	if nb.cancelFunc != nil {
		nb.cancelFunc()
	}

	nrReads := core.EmptyChannel(nb.nodeSyncedChan)
	log.Debug("notifierBootstrapper: emptied channel", "nodeSyncedChan nrReads", nrReads)
	return nil
}
