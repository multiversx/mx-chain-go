package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
)

type sovereignValidatorInfoCreator struct {
	*validatorInfoCreator
}

// NewSovereignValidatorInfoCreator creates a sovereign validator info creator
func NewSovereignValidatorInfoCreator(vic *validatorInfoCreator) (*sovereignValidatorInfoCreator, error) {
	if check.IfNil(vic) {
		return nil, process.ErrNilEpochStartValidatorInfoCreator
	}

	return &sovereignValidatorInfoCreator{
		vic,
	}, nil
}

// CreateMarshalledData creates the marshalled data to be sent to shards
func (svic *sovereignValidatorInfoCreator) CreateMarshalledData(body *block.Body) map[string][][]byte {
	if check.IfNil(body) {
		return nil
	}

	marshalledValidatorInfoTxs := make([][]byte, 0)
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		marshalledValidatorInfoTxs = append(marshalledValidatorInfoTxs, svic.getMarshalledValidatorInfoTxs(miniBlock)...)
	}

	// create broadcast txs as in normal run type processing for now
	mapMarshalledValidatorInfoTxs := make(map[string][][]byte)
	if len(marshalledValidatorInfoTxs) > 0 {
		mapMarshalledValidatorInfoTxs[common.ValidatorInfoTopic] = marshalledValidatorInfoTxs
	}

	return mapMarshalledValidatorInfoTxs
}

func (svic *sovereignValidatorInfoCreator) getMarshalledValidatorInfoTxs(miniBlock *block.MiniBlock) [][]byte {
	validatorInfoCacher := svic.dataPool.CurrentEpochValidatorInfo()

	marshalledValidatorInfoTxs := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		validatorInfoTx, err := validatorInfoCacher.GetValidatorInfo(txHash)
		if err != nil {
			log.Error("validatorInfoCreator.getMarshalledValidatorInfoTxs.GetValidatorInfo", "hash", txHash, "error", err)
			continue
		}

		marshalledData, err := svic.marshalizer.Marshal(validatorInfoTx)
		if err != nil {
			log.Error("validatorInfoCreator.getMarshalledValidatorInfoTxs.Marshal", "hash", txHash, "error", err)
			continue
		}

		// add data directly to pool for sovereign chain, we don't need to broadcast and intercept it
		strCache := process.ShardCacherIdentifier(core.SovereignChainShardId, core.SovereignChainShardId)
		svic.dataPool.ValidatorsInfo().AddData(txHash, validatorInfoTx, validatorInfoTx.Size(), strCache)

		marshalledValidatorInfoTxs = append(marshalledValidatorInfoTxs, marshalledData)
	}

	return marshalledValidatorInfoTxs
}

// CreateValidatorInfoMiniBlocks creates the validatorInfo mini blocks according to the provided validatorInfo map
func (svic *sovereignValidatorInfoCreator) CreateValidatorInfoMiniBlocks(validatorsInfo state.ShardValidatorsInfoMapHandler) (block.MiniBlockSlice, error) {
	if validatorsInfo == nil {
		return nil, epochStart.ErrNilValidatorInfo
	}

	svic.clean()

	miniBlocks := make([]*block.MiniBlock, 0)
	validators := validatorsInfo.GetShardValidatorsInfoMap()[core.SovereignChainShardId]
	if len(validators) == 0 {
		return miniBlocks, nil
	}

	miniBlock, err := svic.createMiniBlock(validators, core.SovereignChainShardId)
	if err != nil {
		return nil, err
	}

	miniBlocks = append(miniBlocks, miniBlock)
	return miniBlocks, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (svic *sovereignValidatorInfoCreator) IsInterfaceNil() bool {
	return svic == nil
}
