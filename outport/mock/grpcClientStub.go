package mock

import (
	"context"

	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
)

// OutportGRPCClientStub -
type OutportGRPCClientStub struct {
	SaveBlockCalled             func(ctx context.Context, in *outportcore.OutportBlock) error
	RevertIndexedBlockCalled    func(ctx context.Context, in *outportcore.BlockData) error
	SaveRoundsInfoCalled        func(ctx context.Context, in *outportcore.RoundsInfo) error
	SaveValidatorsPubKeysCalled func(ctx context.Context, in *outportcore.ValidatorsPubKeys) error
	SaveValidatorsRatingCalled  func(ctx context.Context, in *outportcore.ValidatorsRating) error
	SaveAccountsCalled          func(ctx context.Context, in *outportcore.Accounts) error
	FinalizedBlockEventCalled   func(ctx context.Context, in *outportcore.FinalizedBlock) error
}

// SaveBlock -
func (stub *OutportGRPCClientStub) SaveBlock(ctx context.Context, in *outportcore.OutportBlock) (*outportcore.ResponseData, error) {
	return nil, stub.SaveBlockCalled(ctx, in)
}

// RevertIndexedBlock -
func (stub *OutportGRPCClientStub) RevertIndexedBlock(ctx context.Context, in *outportcore.BlockData) (*outportcore.ResponseData, error) {
	return nil, stub.RevertIndexedBlockCalled(ctx, in)
}

// SaveRoundsInfo -
func (stub *OutportGRPCClientStub) SaveRoundsInfo(ctx context.Context, in *outportcore.RoundsInfo) (*outportcore.ResponseData, error) {
	return nil, stub.SaveRoundsInfoCalled(ctx, in)
}

// SaveValidatorsPubKeys -
func (stub *OutportGRPCClientStub) SaveValidatorsPubKeys(ctx context.Context, in *outportcore.ValidatorsPubKeys) (*outportcore.ResponseData, error) {
	return nil, stub.SaveValidatorsPubKeysCalled(ctx, in)
}

// SaveValidatorsRating -
func (stub *OutportGRPCClientStub) SaveValidatorsRating(ctx context.Context, in *outportcore.ValidatorsRating) (*outportcore.ResponseData, error) {
	return nil, stub.SaveValidatorsRatingCalled(ctx, in)
}

// SaveAccounts -
func (stub *OutportGRPCClientStub) SaveAccounts(ctx context.Context, in *outportcore.Accounts) (*outportcore.ResponseData, error) {
	return nil, stub.SaveAccountsCalled(ctx, in)
}

// FinalizedBlockEvent -
func (stub *OutportGRPCClientStub) FinalizedBlockEvent(ctx context.Context, in *outportcore.FinalizedBlock) (*outportcore.ResponseData, error) {
	return nil, stub.FinalizedBlockEventCalled(ctx, in)
}

// IsInterfaceNil -
func (stub *OutportGRPCClientStub) IsInterfaceNil() bool {
	return stub == nil
}
