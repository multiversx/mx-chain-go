package grpcdriver

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/outport"
)

const defaultGRPCCallTimeout = 30 * time.Second

// grpcDriver forwards outport driver calls to a remote gRPC outport service.
type grpcDriver struct {
	client         grpcadapter.OutportClient
	marshaller     marshal.Marshalizer
	requestTimeout time.Duration
}

// NewGRPCDriver creates a driver implementation backed by a gRPC outport client.
func NewGRPCDriver(client grpcadapter.OutportClient, marshaller marshal.Marshalizer, requestTimeout time.Duration) (outport.Driver, error) {
	if check.IfNil(client) {
		return nil, ErrNilGRPCClient
	}
	if check.IfNil(marshaller) {
		return nil, ErrNilMarshaller
	}

	if requestTimeout <= 0 {
		requestTimeout = defaultGRPCCallTimeout
	}

	return &grpcDriver{
		client:         client,
		marshaller:     marshaller,
		requestTimeout: requestTimeout,
	}, nil
}

func (g *grpcDriver) newCallContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), g.requestTimeout)
}

// SaveBlock forwards the block payload to the remote outport service.
func (g *grpcDriver) SaveBlock(outportBlock *outportcore.OutportBlock) error {
	ctx, cancel := g.newCallContext()
	defer cancel()
	_, err := g.client.SaveBlock(ctx, outportBlock)
	return err
}

// RevertIndexedBlock calls the remote service to rollback indexed data for a block.
func (g *grpcDriver) RevertIndexedBlock(blockData *outportcore.BlockData) error {
	ctx, cancel := g.newCallContext()
	defer cancel()
	_, err := g.client.RevertIndexedBlock(ctx, blockData)
	return err
}

// SaveRoundsInfo synchronizes latest rounds metadata with the outport service.
func (g *grpcDriver) SaveRoundsInfo(roundsInfos *outportcore.RoundsInfo) error {
	ctx, cancel := g.newCallContext()
	defer cancel()
	_, err := g.client.SaveRoundsInfo(ctx, roundsInfos)
	return err
}

// SaveValidatorsPubKeys sends validator keys to the remote store.
func (g *grpcDriver) SaveValidatorsPubKeys(validatorsPubKeys *outportcore.ValidatorsPubKeys) error {
	ctx, cancel := g.newCallContext()
	defer cancel()
	_, err := g.client.SaveValidatorsPubKeys(ctx, validatorsPubKeys)
	return err
}

// SaveValidatorsRating streams the current validators rating snapshot to gRPC.
func (g *grpcDriver) SaveValidatorsRating(validatorsRating *outportcore.ValidatorsRating) error {
	ctx, cancel := g.newCallContext()
	defer cancel()
	_, err := g.client.SaveValidatorsRating(ctx, validatorsRating)
	return err
}

// SaveAccounts writes account details via the remote controller.
func (g *grpcDriver) SaveAccounts(accounts *outportcore.Accounts) error {
	ctx, cancel := g.newCallContext()
	defer cancel()
	_, err := g.client.SaveAccounts(ctx, accounts)
	return err
}

// FinalizedBlock publishes the finalized block event over gRPC.
func (g *grpcDriver) FinalizedBlock(finalizedBlock *outportcore.FinalizedBlock) error {
	ctx, cancel := g.newCallContext()
	defer cancel()
	_, err := g.client.FinalizedBlockEvent(ctx, finalizedBlock)
	return err
}

// GetMarshaller returns the marshaller assigned to this driver for serialization.
func (g *grpcDriver) GetMarshaller() marshal.Marshalizer {
	return g.marshaller
}

// SetCurrentSettings does nothing
func (g *grpcDriver) SetCurrentSettings(_ outportcore.OutportConfig) error {
	return nil
}

// RegisterHandler does nothing
func (g *grpcDriver) RegisterHandler(_ func() error, _ string) error {
	return nil
}

// Close closes the underlying gRPC client connection.
func (g *grpcDriver) Close() error {
	return g.client.Close()
}

// IsInterfaceNil supports the interface nil check used by the caller.
func (g *grpcDriver) IsInterfaceNil() bool {
	return g == nil
}
