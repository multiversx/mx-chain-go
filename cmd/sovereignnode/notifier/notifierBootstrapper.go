package notifier

import (
	"context"
	"os"
	"syscall"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	notifierProcess "github.com/multiversx/mx-chain-sovereign-notifier-go/process"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

const roundsThreshold = process.MaxRoundsWithoutNewBlockReceived + 1

var log = logger.GetOrCreate("notifier-bootstrap")

// ArgsNotifierBootstrapper defines args needed to create a new notifier bootstrapper
type ArgsNotifierBootstrapper struct {
	IncomingHeaderHandler process.IncomingHeaderSubscriber
	SovereignNotifier     notifierProcess.SovereignNotifier
	ForkDetector          process.ForkDetector
	Bootstrapper          process.Bootstrapper
	SigStopNode           chan os.Signal
	RoundDuration         uint64
}

type notifierBootstrapper struct {
	incomingHeaderHandler process.IncomingHeaderSubscriber
	sovereignNotifier     notifierProcess.SovereignNotifier
	forkDetector          process.ForkDetector
	sigStopNode           chan os.Signal

	syncedRoundsChan chan int32
	cancelFunc       func()
	roundDuration    uint64
}

// NewNotifierBootstrapper creates a ws receiver connection registration bootstrapper
func NewNotifierBootstrapper(args ArgsNotifierBootstrapper) (*notifierBootstrapper, error) {
	if err := checkArgs(args); err != nil {
		return nil, err
	}

	nb := &notifierBootstrapper{
		incomingHeaderHandler: args.IncomingHeaderHandler,
		sovereignNotifier:     args.SovereignNotifier,
		forkDetector:          args.ForkDetector,
		syncedRoundsChan:      make(chan int32, 1),
		cancelFunc:            nil,
		roundDuration:         args.RoundDuration,
		sigStopNode:           args.SigStopNode,
	}

	args.Bootstrapper.AddSyncStateListener(nb.receivedSyncState)

	return nb, nil
}

func checkArgs(args ArgsNotifierBootstrapper) error {
	if check.IfNil(args.IncomingHeaderHandler) {
		return errors.ErrNilIncomingHeaderSubscriber
	}
	if check.IfNil(args.SovereignNotifier) {
		return errNilSovereignNotifier
	}
	if check.IfNil(args.ForkDetector) {
		return errors.ErrNilForkDetector
	}
	if check.IfNil(args.Bootstrapper) {
		return process.ErrNilBootstrapper
	}
	if args.RoundDuration == 0 {
		return errors.ErrInvalidRoundDuration
	}

	return nil
}

func (nb *notifierBootstrapper) receivedSyncState(isNodeSynchronized bool) {
	if isNodeSynchronized && nb.forkDetector.GetHighestFinalBlockNonce() != 0 {
		nb.writeRoundsDeltaOnChan(1)
	} else if !isNodeSynchronized {
		nb.writeRoundsDeltaOnChan(-1)
	}
}

func (nb *notifierBootstrapper) writeRoundsDeltaOnChan(delta int32) {
	select {
	case nb.syncedRoundsChan <- delta:
	default:
	}
}

// Start will start waiting on a go routine to be notified via syncedRoundsChan when the sovereign node is synced.
// Meanwhile, it will print the current node state in log. When node is fully synced, it will register the incoming header
// processor to the websocket listener and exit the waiting loop.
func (nb *notifierBootstrapper) Start() {
	var ctx context.Context
	ctx, nb.cancelFunc = context.WithCancel(context.Background())
	go nb.checkNodeState(ctx)
}

func (nb *notifierBootstrapper) checkNodeState(ctx context.Context) {
	timeToWaitReSync := uint64(roundsThreshold) * nb.roundDuration
	ticker := time.NewTicker(time.Duration(timeToWaitReSync) * time.Millisecond)
	defer ticker.Stop()

	var syncedRounds uint32

	for {
		select {
		case <-ctx.Done():
			log.Debug("notifierBootstrapper.checkNodeState: worker's go routine is stopping...")
			return
		case delta := <-nb.syncedRoundsChan:
			syncedRounds = updateSyncedRounds(syncedRounds, delta)
			if syncedRounds < uint32(roundsThreshold) {
				log.Debug("notifierBootstrapper.checkNodeState", "syncedRounds", syncedRounds)
				continue
			}

			err := nb.sovereignNotifier.RegisterHandler(nb.incomingHeaderHandler)
			if err != nil {
				log.Error("notifierBootstrapper: sovereignNotifier.RegisterHandler", "err", err)
				nb.sigStopNode <- syscall.SIGTERM
			} else {
				log.Info("notifierBootstrapper.checkNodeState", "is node synced", true)
			}

			return
		case <-ticker.C:
			log.Debug("notifierBootstrapper.checkNodeState", "is node synced", false)
		}
	}
}

func updateSyncedRounds(syncedRounds uint32, delta int32) uint32 {
	if delta > 0 {
		syncedRounds += uint32(delta)
	} else if syncedRounds > 0 {
		syncedRounds--
	}

	return syncedRounds
}

// Close cancels current context and empties channel reads
func (nb *notifierBootstrapper) Close() error {
	if nb.cancelFunc != nil {
		nb.cancelFunc()
	}

	nrReads := emptyChannel(nb.syncedRoundsChan)
	log.Debug("notifierBootstrapper: emptied channel", "syncedRoundsChan nrReads", nrReads)
	return nil
}

func emptyChannel(ch chan int32) int {
	readsCnt := 0
	for {
		select {
		case <-ch:
			readsCnt++
		default:
			return readsCnt
		}
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (nb *notifierBootstrapper) IsInterfaceNil() bool {
	return nb == nil
}
