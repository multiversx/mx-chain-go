package txsimulator

import "errors"

// ErrNilTxSimulatorProcessor signals that a nil transaction simulator processor has been provided
var ErrNilTxSimulatorProcessor = errors.New("nil transaction simulator processor")

// ErrNilIntermediateProcessorContainer signals that intermediate processors container is nil
var ErrNilIntermediateProcessorContainer = errors.New("intermediate processor container is nil")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to use a nil pubkey converter")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

// ErrNilAccountsAdapter signals that a nil accounts adapter has been provided
var ErrNilAccountsAdapter = errors.New("trying to set nil accounts adapter")
