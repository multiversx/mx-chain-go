package esdtSupply

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

const (
	supplyCorrectionPrefix = "supplyCorrection"
)

type supplyCorrectionProcessor struct {
	shardID  uint32
	logsProc *logsProcessor
}

func newSupplyCorrectionProcessor(shardID uint32, logsProc *logsProcessor) *supplyCorrectionProcessor {
	return &supplyCorrectionProcessor{
		shardID:  shardID,
		logsProc: logsProc,
	}
}

func (scp *supplyCorrectionProcessor) applySupplyCorrections(supplyCorrections []config.SupplyCorrection) error {
	supplies := make(map[string]*SupplyESDT)
	for _, supplyCorrection := range supplyCorrections {
		shouldApply, err := scp.shouldApplyCorrection(supplyCorrection)
		if err != nil {
			return err
		}
		if !shouldApply {
			continue
		}

		tokenSupply, err := scp.logsProc.getESDTSupply([]byte(supplyCorrection.Token))
		if err != nil {
			return err
		}

		// if token supply was recomputed should not apply supply correction
		if tokenSupply.RecomputedSupply {
			return nil
		}

		correctionValue, ok := big.NewInt(0).SetString(supplyCorrection.Value, 10)
		if !ok {
			return errors.New("failed to parse supplyCorrection value")
		}

		switch {
		case correctionValue.Cmp(big.NewInt(0)) > 0:
			tokenSupply.Supply.Add(tokenSupply.Supply, correctionValue)
			tokenSupply.Minted.Add(tokenSupply.Minted, correctionValue)

		case correctionValue.Cmp(big.NewInt(0)) < 0:
			negCorrectionValue := big.NewInt(0).Neg(correctionValue)
			tokenSupply.Supply.Add(tokenSupply.Supply, correctionValue)
			tokenSupply.Burned.Add(tokenSupply.Minted, negCorrectionValue)
		}

		supplies[supplyCorrection.Token] = tokenSupply

		correction := &Correction{WasApplied: true}
		err = scp.saveCorrectionInfo(getSupplyCorrectionKey(supplyCorrection.ID), correction)
		if err != nil {
			return err
		}
	}

	err := scp.logsProc.saveSupplies(supplies)
	if err != nil {
		return err
	}

	return nil
}

func (scp *supplyCorrectionProcessor) shouldApplyCorrection(supplyCorrection config.SupplyCorrection) (bool, error) {
	if supplyCorrection.ShardID != scp.shardID {
		return false, nil
	}

	latestProcessedBlockNonce, err := scp.logsProc.nonceProc.getLatestProcessedBlockNonceFromStorage()
	if err != nil {
		return false, err
	}
	if latestProcessedBlockNonce < supplyCorrection.BlockNonce {
		return false, nil
	}

	storageKey := getSupplyCorrectionKey(supplyCorrection.ID)
	correctionFromStorage, err := scp.getCorrectionInfo(storageKey)
	if err != nil {
		return false, err
	}
	if correctionFromStorage.WasApplied {
		return false, nil
	}

	return true, nil
}

func (scp *supplyCorrectionProcessor) saveCorrectionInfo(storageKey []byte, correction *Correction) error {
	correctionBytes, err := scp.logsProc.marshalizer.Marshal(correction)
	if err != nil {
		return err
	}

	return scp.logsProc.suppliesStorer.Put(storageKey, correctionBytes)
}

func (scp *supplyCorrectionProcessor) getCorrectionInfo(key []byte) (*Correction, error) {
	correctionBytes, err := scp.logsProc.suppliesStorer.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return &Correction{}, nil
		}
		return nil, err
	}

	correction := &Correction{}
	err = scp.logsProc.marshalizer.Unmarshal(correction, correctionBytes)

	return correction, err
}

func getSupplyCorrectionKey(id string) []byte {
	return bytes.Join([][]byte{[]byte(supplyCorrectionPrefix), []byte(id)}, []byte{})
}
