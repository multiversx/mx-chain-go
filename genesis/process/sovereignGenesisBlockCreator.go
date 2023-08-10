package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type sovereignGenesisBlockCreator struct {
	*genesisBlockCreator
}

// NewSovereignGenesisBlockCreator creates a new sovereign genesis block creator instance
func NewSovereignGenesisBlockCreator(gbc *genesisBlockCreator) (*sovereignGenesisBlockCreator, error) {
	if gbc == nil {
		return nil, errNilGenesisBlockCreator
	}

	return &sovereignGenesisBlockCreator{
		genesisBlockCreator: gbc,
	}, nil
}

// CreateGenesisBlocks will try to create the genesis blocks for all shards
func (gbc *sovereignGenesisBlockCreator) CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	if !mustDoGenesisProcess(gbc.arg) {
		return gbc.createSovereignEmptyGenesisBlocks()
	}

	if mustDoHardForkImportProcess(gbc.arg) {
		err := gbc.arg.importHandler.ImportAll()
		if err != nil {
			return nil, err
		}

		err = gbc.computeDNSAddresses(gbc.arg.EpochConfig.EnableEpochs)
		if err != nil {
			return nil, err
		}
	}

	shardIDs := make([]uint32, 1)
	shardIDs[0] = core.SovereignChainShardId
	return gbc.baseCreateGenesisBlocks(shardIDs)
}

func (gbc *sovereignGenesisBlockCreator) createSovereignEmptyGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.computeDNSAddresses(createGenesisConfig())
	if err != nil {
		return nil, err
	}

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(gbc.arg)

	mapEmptyGenesisBlocks := make(map[uint32]data.HeaderHandler, 1)
	mapEmptyGenesisBlocks[core.SovereignChainShardId] = &block.SovereignChainHeader{
		Header: &block.Header{
			Round:     round,
			Nonce:     nonce,
			Epoch:     epoch,
			TimeStamp: gbc.arg.GenesisTime,
			ShardID:   core.SovereignChainShardId,
		},
	}

	return mapEmptyGenesisBlocks, nil
}
