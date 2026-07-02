package proxy

import (
	"errors"
)

// ErrNilChronologyHandler is the error returned when the chronology handler is nil
var ErrNilChronologyHandler = errors.New("nil chronology handler")

// ErrNilConsensusCoreHandler is the error returned when the consensus core handler is nil
var ErrNilConsensusCoreHandler = errors.New("nil consensus core handler")

// ErrNilConsensusState is the error returned when the consensus state is nil
var ErrNilConsensusState = errors.New("nil consensus state")

// ErrNilWorker is the error returned when the worker is nil
var ErrNilWorker = errors.New("nil worker")

// ErrNilSignatureThrottler is the error returned when the signature throttler is nil
var ErrNilSignatureThrottler = errors.New("nil signature throttler")

// ErrNilAppStatusHandler is the error returned when the app status handler is nil
var ErrNilAppStatusHandler = errors.New("nil app status handler")

// ErrNilOutportHandler is the error returned when the outport handler is nil
var ErrNilOutportHandler = errors.New("nil outport handler")

// ErrNilSentSignatureTracker is the error returned when the sent signature tracker is nil
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrNilChainID is the error returned when the chain ID is nil
var ErrNilChainID = errors.New("nil chain ID")

// ErrNilCurrentPid is the error returned when the current PID is nil
var ErrNilCurrentPid = errors.New("nil current PID")

// ErrNilEnableEpochsHandler is the error returned when the enable epochs handler is nil
var ErrNilEnableEpochsHandler = errors.New("nil enable epochs handler")
