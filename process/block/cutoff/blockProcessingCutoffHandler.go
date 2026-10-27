package cutoff

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/block/cutoff")

type blockProcessingCutoffHandler struct {
	config              config.BlockProcessingCutoffConfig
	stopRound           uint64
	stopNonce           uint64
	stopEpoch           uint32
	chanStopNodeProcess chan endProcess.ArgEndProcess
	closeChan           chan struct{}
	closeOnce           sync.Once
}

// NewBlockProcessingCutoffHandler will return a new instance of blockProcessingCutoffHandler
func NewBlockProcessingCutoffHandler(
	cfg config.BlockProcessingCutoffConfig,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (*blockProcessingCutoffHandler, error) {
	b := &blockProcessingCutoffHandler{
		config:              cfg,
		stopEpoch:           math.MaxUint32,
		stopNonce:           math.MaxUint64,
		stopRound:           math.MaxUint64,
		chanStopNodeProcess: chanStopNodeProcess,
		closeChan:           make(chan struct{}),
	}

	err := b.applyConfig(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Mode == common.BlockProcessingCutoffModeGracefulStop && chanStopNodeProcess == nil {
		return nil, errNilStopNodeChannel
	}

	log.Warn("node is started by using block processing cutoff", "mode", cfg.Mode, cfg.CutoffTrigger, cfg.Value)
	return b, nil
}

func (b *blockProcessingCutoffHandler) applyConfig(cfg config.BlockProcessingCutoffConfig) error {
	switch common.BlockProcessingCutoffMode(cfg.Mode) {
	case common.BlockProcessingCutoffModeProcessError:
	case common.BlockProcessingCutoffModePause:
	case common.BlockProcessingCutoffModeGracefulStop:
	default:
		return fmt.Errorf("%w, provided value=%s", errInvalidBlockProcessingCutOffMode, cfg.Mode)
	}

	switch common.BlockProcessingCutoffTrigger(cfg.CutoffTrigger) {
	case common.BlockProcessingCutoffByRound:
		b.stopRound = cfg.Value
	case common.BlockProcessingCutoffByNonce:
		b.stopNonce = cfg.Value
	case common.BlockProcessingCutoffByEpoch:
		b.stopEpoch = uint32(cfg.Value)
	default:
		return fmt.Errorf("%w, provided value=%s", errInvalidBlockProcessingCutOffTrigger, cfg.CutoffTrigger)
	}

	return nil
}

// HandleGracefulStopCutoff stops the node after all work for the cutoff block is complete
func (b *blockProcessingCutoffHandler) HandleGracefulStopCutoff(header data.HeaderHandler, beforeStop func()) {
	shouldSkip := !b.config.Enabled ||
		check.IfNil(header) ||
		b.config.Mode != common.BlockProcessingCutoffModeGracefulStop
	if shouldSkip {
		return
	}

	trigger, value, isTriggered := b.isTriggered(header)
	if !isTriggered {
		return
	}

	if beforeStop != nil {
		beforeStop()
	}

	log.Info("block processing cutoff reached, stopping the node", trigger, value)
	select {
	case b.chanStopNodeProcess <- endProcess.ArgEndProcess{
		Reason:      "BlockProcessingCutoff",
		Description: fmt.Sprintf("block processing completed through %s %d", trigger, value),
	}:
	case <-b.closeChan:
		return
	}

	<-b.closeChan
}

// Close releases a graceful cutoff waiting for component shutdown
func (b *blockProcessingCutoffHandler) Close() {
	b.closeOnce.Do(func() {
		close(b.closeChan)
	})
}

// HandlePauseCutoff will pause the processing if the required coordinates are met
func (b *blockProcessingCutoffHandler) HandlePauseCutoff(header data.HeaderHandler) {
	shouldSkip := !b.config.Enabled ||
		check.IfNil(header) ||
		b.config.Mode != common.BlockProcessingCutoffModePause
	if shouldSkip {
		return
	}

	trigger, value, isTriggered := b.isTriggered(header)
	if !isTriggered {
		return
	}

	log.Info("cutting off the block processing. The node will not advance", trigger, value)
	go func() {
		for {
			time.Sleep(time.Minute)
			log.Info("node is in block processing cut-off mode", trigger, value)
		}
	}()
	neverEndingChannel := make(chan struct{})
	<-neverEndingChannel
}

// HandleProcessErrorCutoff will return error if the processing block matches the required coordinates
func (b *blockProcessingCutoffHandler) HandleProcessErrorCutoff(header data.HeaderHandler) error {
	shouldSkip := !b.config.Enabled ||
		check.IfNil(header) ||
		b.config.Mode != common.BlockProcessingCutoffModeProcessError
	if shouldSkip {
		return nil
	}

	trigger, value, isTriggered := b.isTriggered(header)
	if !isTriggered {
		return nil
	}

	log.Info("block processing cutoff - return err", trigger, value)
	return errProcess
}

func (b *blockProcessingCutoffHandler) isTriggered(header data.HeaderHandler) (common.BlockProcessingCutoffTrigger, uint64, bool) {
	if header.GetRound() >= b.stopRound {
		return common.BlockProcessingCutoffByRound, header.GetRound(), true
	}
	if header.GetNonce() >= b.stopNonce {
		return common.BlockProcessingCutoffByNonce, header.GetNonce(), true
	}
	if header.GetEpoch() >= b.stopEpoch {
		return common.BlockProcessingCutoffByEpoch, uint64(header.GetEpoch()), true
	}

	return "", 0, false
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *blockProcessingCutoffHandler) IsInterfaceNil() bool {
	return b == nil
}
