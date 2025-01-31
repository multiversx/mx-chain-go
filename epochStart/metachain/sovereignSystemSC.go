package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
)

const metaChainShardIdentifier uint8 = 255

type sovereignSystemSC struct {
	*systemSCProcessor
}

// NewSovereignSystemSCProcessor creates a sovereign system sc processor
func NewSovereignSystemSCProcessor(sysSC *systemSCProcessor) (*sovereignSystemSC, error) {
	if check.IfNil(sysSC) {
		return nil, process.ErrNilEpochStartSystemSCProcessor
	}

	return &sovereignSystemSC{
		systemSCProcessor: sysSC,
	}, nil
}

// ProcessDelegationRewards will process the rewards which are directed towards the delegation system smart contracts
func (s *sovereignSystemSC) ProcessDelegationRewards(
	miniBlocks block.MiniBlockSlice,
	txCache epochStart.TransactionCacher,
) error {
	if txCache == nil {
		return epochStart.ErrNilLocalTxCache
	}

	rwdMb := getRewardsMiniBlockForSovereign(miniBlocks)
	if rwdMb == nil {
		return nil
	}

	for _, txHash := range rwdMb.TxHashes {
		rwdTx, err := txCache.GetTx(txHash)
		if err != nil {
			return err
		}

		err = s.executeRewardTx(txHash, rwdTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func getRewardsMiniBlockForSovereign(miniBlocks block.MiniBlockSlice) *block.MiniBlock {
	for _, miniBlock := range miniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		return miniBlock
	}
	return nil
}

func (s *sovereignSystemSC) executeRewardTx(txHash []byte, rwdTx data.TransactionHandler) error {
	if core.IsSmartContractOnMetachain([]byte{metaChainShardIdentifier}, rwdTx.GetRcvAddr()) {
		return s.legacySystemSCProcessor.executeRewardTx(rwdTx)
	}

	return s.addRewardTxValue(txHash, rwdTx)
}

func (s *sovereignSystemSC) addRewardTxValue(txHash []byte, rwdTx data.TransactionHandler) error {
	userAcc, err := s.getUserAccount(rwdTx.GetRcvAddr())
	if err != nil {
		return err
	}

	log.Trace("executing sovereign reward tx",
		"txHash", txHash,
		"receiver", rwdTx.GetRcvAddr(),
		"value", rwdTx.GetValue().String(),
	)

	err = userAcc.AddToBalance(rwdTx.GetValue())
	if err != nil {
		return err
	}

	return s.userAccountsDB.SaveAccount(userAcc)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (s *sovereignSystemSC) IsInterfaceNil() bool {
	return s == nil
}
