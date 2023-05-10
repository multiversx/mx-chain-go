package cutoff

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/block/cutoff")

type blockProcessingCutoffHandler struct {
	config config.BlockProcessingCutoffConfig
}

// NewBlockProcessingCutoffHandler will return a new instance of blockProcessingCutoffHandler
func NewBlockProcessingCutoffHandler(cfg config.BlockProcessingCutoffConfig) (*blockProcessingCutoffHandler, error) {
	err := checkConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &blockProcessingCutoffHandler{
		config: cfg,
	}, nil
}

// HandlePauseCutoff will pause the processing if the required coordinates are met
func (b *blockProcessingCutoffHandler) HandlePauseCutoff(header data.HeaderHandler) {
	shouldSkip := !b.config.Enabled ||
		check.IfNil(header) ||
		b.config.Mode != common.BlockProcessingCutoffModePause
	if shouldSkip {
		return
	}

	blockingCutoffFunction := func(printArgs ...interface{}) error {
		log.Info("cutting off the block processing. The node will not advance", printArgs...)
		go func() {
			for {
				time.Sleep(time.Minute)
				log.Info("node is in block processing cut-off mode", printArgs...)
			}
		}()
		neverEndingChannel := make(chan struct{})
		<-neverEndingChannel

		return nil // should not reach this point
	}

	_ = b.handleCutoffIfCoordinatesAreMet(header, blockingCutoffFunction)
	// should never reach this point
}

// HandleProcessErrorCutoff will return error if the processing the block at the required coordinates
func (b *blockProcessingCutoffHandler) HandleProcessErrorCutoff(header data.HeaderHandler) error {
	shouldSkip := !b.config.Enabled ||
		check.IfNil(header) ||
		b.config.Mode != common.BlockProcessingCutoffModeProcessError
	if shouldSkip {
		return nil
	}

	return b.handleCutoffIfCoordinatesAreMet(header, func(printArgs ...interface{}) error {
		log.Info("block processing cutoff - return err", printArgs...)
		return errProcess
	})
}

func (b *blockProcessingCutoffHandler) handleCutoffIfCoordinatesAreMet(header data.HeaderHandler, cutOffFunction func(printArgs ...interface{}) error) error {
	value := b.config.Value

	switch common.BlockProcessingCutoffTrigger(b.config.CutoffTrigger) {
	case common.BlockProcessingCutoffByRound:
		if header.GetRound() >= value {
			err := cutOffFunction("round", header.GetRound())
			if err != nil {
				return err
			}
		}
	case common.BlockProcessingCutoffByNonce:
		if header.GetNonce() >= value {
			err := cutOffFunction("nonce", header.GetNonce())
			if err != nil {
				return err
			}
		}
	case common.BlockProcessingCutoffByEpoch:
		if header.GetEpoch() >= uint32(value) {
			err := cutOffFunction("epoch", header.GetEpoch())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *blockProcessingCutoffHandler) IsInterfaceNil() bool {
	return b == nil
}

func checkConfig(cutOffConfig config.BlockProcessingCutoffConfig) error {
	if !cutOffConfig.Enabled {
		// don't even check the configs if the feature is disabled. Useful when a node doesn't update `prefs.toml` with
		// the new configuration
		return nil
	}
	mode := common.BlockProcessingCutoffMode(cutOffConfig.Mode)
	isValidMode := mode == common.BlockProcessingCutoffModePause || mode == common.BlockProcessingCutoffModeProcessError
	if !isValidMode {
		return fmt.Errorf("%w. provided value=%s", errInvalidBlockProcessingCutOffMode, mode)
	}

	cutOffTrigger := common.BlockProcessingCutoffTrigger(cutOffConfig.CutoffTrigger)
	isValidCutOffTrigger := cutOffTrigger == common.BlockProcessingCutoffByRound ||
		cutOffTrigger == common.BlockProcessingCutoffByNonce ||
		cutOffTrigger == common.BlockProcessingCutoffByEpoch
	if !isValidCutOffTrigger {
		return fmt.Errorf("%w. provided value=%s", errInvalidBlockProcessingCutOffTrigger, cutOffTrigger)
	}

	return nil
}
