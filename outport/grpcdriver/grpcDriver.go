package grpcdriver

import (
	"context"

	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/outport"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("grpcDriver")

type grpcDriver struct {
	client     grpcadapter.OutportClient
	marshaller marshal.Marshalizer
}

func NewGRPCDriver(client grpcadapter.OutportClient, marshaller marshal.Marshalizer) (outport.Driver, error) {
	return &grpcDriver{
		client:     client,
		marshaller: marshaller,
	}, nil
}

func (g *grpcDriver) SaveBlock(outportBlock *outportcore.OutportBlock) error {
	_, err := g.client.SaveBlock(context.Background(), outportBlock)
	if err != nil {
		return err
	}

	return nil
}

func (g *grpcDriver) RevertIndexedBlock(blockData *outportcore.BlockData) error {
	_, err := g.client.RevertIndexedBlock(context.Background(), blockData)
	if err != nil {
		return err
	}

	return nil
}

func (g *grpcDriver) SaveRoundsInfo(roundsInfos *outportcore.RoundsInfo) error {
	_, err := g.client.SaveRoundsInfo(context.Background(), roundsInfos)
	if err != nil {
		return err
	}

	return nil
}

func (g *grpcDriver) SaveValidatorsPubKeys(validatorsPubKeys *outportcore.ValidatorsPubKeys) error {
	_, err := g.client.SaveValidatorsPubKeys(context.Background(), validatorsPubKeys)
	if err != nil {
		return err
	}

	return nil
}

func (g *grpcDriver) SaveValidatorsRating(validatorsRating *outportcore.ValidatorsRating) error {
	_, err := g.client.SaveValidatorsRating(context.Background(), validatorsRating)
	if err != nil {
		return err
	}

	return nil
}

func (g *grpcDriver) SaveAccounts(accounts *outportcore.Accounts) error {
	_, err := g.client.SaveAccounts(context.Background(), accounts)
	if err != nil {
		return err
	}

	return nil
}

func (g *grpcDriver) FinalizedBlock(finalizedBlock *outportcore.FinalizedBlock) error {
	_, err := g.client.FinalizedBlockEvent(context.Background(), finalizedBlock)
	if err != nil {
		return err
	}

	return nil
}

func (g *grpcDriver) GetMarshaller() marshal.Marshalizer {
	return g.marshaller
}

func (g *grpcDriver) SetCurrentSettings(_ outportcore.OutportConfig) error {
	return nil
}

func (g *grpcDriver) RegisterHandler(_ func() error, _ string) error {
	return nil
}

func (g *grpcDriver) Close() error {
	return nil
}

func (g *grpcDriver) IsInterfaceNil() bool {
	return g == nil
}
