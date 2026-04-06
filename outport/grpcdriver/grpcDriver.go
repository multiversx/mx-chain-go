package grpcdriver

import (
	"context"

	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/outport"
)

// grpcDriver forwards outport driver calls to a remote gRPC outport service.
type grpcDriver struct {
	client     grpcadapter.OutportClient
	marshaller marshal.Marshalizer
}

// NewGRPCDriver creates a driver implementation backed by a gRPC outport client.
func NewGRPCDriver(client grpcadapter.OutportClient, marshaller marshal.Marshalizer) (outport.Driver, error) {
	return &grpcDriver{
		client:     client,
		marshaller: marshaller,
	}, nil
}

// SaveBlock forwards the block payload to the remote outport service.
func (g *grpcDriver) SaveBlock(outportBlock *outportcore.OutportBlock) error {
	_, err := g.client.SaveBlock(context.Background(), outportBlock)
	if err != nil {
		return err
	}

	return nil
}

// RevertIndexedBlock calls the remote service to rollback indexed data for a block.
func (g *grpcDriver) RevertIndexedBlock(blockData *outportcore.BlockData) error {
	_, err := g.client.RevertIndexedBlock(context.Background(), blockData)
	if err != nil {
		return err
	}

	return nil
}

// SaveRoundsInfo synchronizes latest rounds metadata with the outport service.
func (g *grpcDriver) SaveRoundsInfo(roundsInfos *outportcore.RoundsInfo) error {
	_, err := g.client.SaveRoundsInfo(context.Background(), roundsInfos)
	if err != nil {
		return err
	}

	return nil
}

// SaveValidatorsPubKeys sends validator keys to the remote store.
func (g *grpcDriver) SaveValidatorsPubKeys(validatorsPubKeys *outportcore.ValidatorsPubKeys) error {
	_, err := g.client.SaveValidatorsPubKeys(context.Background(), validatorsPubKeys)
	if err != nil {
		return err
	}

	return nil
}

// SaveValidatorsRating streams the current validators rating snapshot to gRPC.
func (g *grpcDriver) SaveValidatorsRating(validatorsRating *outportcore.ValidatorsRating) error {
	_, err := g.client.SaveValidatorsRating(context.Background(), validatorsRating)
	if err != nil {
		return err
	}

	return nil
}

// SaveAccounts writes account details via the remote controller.
func (g *grpcDriver) SaveAccounts(accounts *outportcore.Accounts) error {
	_, err := g.client.SaveAccounts(context.Background(), accounts)
	if err != nil {
		return err
	}

	return nil
}

// FinalizedBlock publishes the finalized block event over gRPC.
func (g *grpcDriver) FinalizedBlock(finalizedBlock *outportcore.FinalizedBlock) error {
	_, err := g.client.FinalizedBlockEvent(context.Background(), finalizedBlock)
	if err != nil {
		return err
	}

	return nil
}

// GetMarshaller returns the marshaller assigned to this driver for serialization.
func (g *grpcDriver) GetMarshaller() marshal.Marshalizer {
	return g.marshaller
}

// SetCurrentSettings does nothing
func (g *grpcDriver) SetCurrentSettings(_ outportcore.OutportConfig) error {
	return nil
}

// RegisterHandler does nothings
func (g *grpcDriver) RegisterHandler(_ func() error, _ string) error {
	return nil
}

// Close does nothing/
func (g *grpcDriver) Close() error {
	return nil
}

// IsInterfaceNil supports the interface nil check used by the caller.
func (g *grpcDriver) IsInterfaceNil() bool {
	return g == nil
}
