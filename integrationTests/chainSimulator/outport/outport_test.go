package outport

import (
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/outport/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainSimulatorWithOutportGrpcEnabled(t *testing.T) {
	var count atomic.Int32
	indexer := &mock.DriverStub{
		SaveBlockCalled: func(outportBlock *outportcore.OutportBlock) error {
			assert.NotNil(t, outportBlock.BlockData)
			count.Add(1)
			return nil
		},
	}
	outportGRPCServer, err := grpcadapter.NewOutportGRPCServer("127.0.0.1:0", indexer)
	require.Nil(t, err)
	address := outportGRPCServer.Address()
	require.NotEmpty(t, address)

	defer func() {
		_ = outportGRPCServer.Close()
	}()
	go func() {
		_ = outportGRPCServer.Start()
	}()

	roundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            "../../../cmd/node/config/",
		NumOfShards:                    3,
		RoundDurationInMillis:          6000,
		SupernovaRoundDurationInMillis: 600,
		RoundsPerEpoch:                 roundsPerEpochOpt,
		SupernovaRoundsPerEpoch:        roundsPerEpochOpt,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.ExternalConfig.GRPCDriversConfig[0].Enabled = true
			cfg.ExternalConfig.GRPCDriversConfig[0].URL = address
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
	require.Equal(t, int32(8), count.Load())
}
