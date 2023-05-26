package block

import (
	"bytes"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/vm"
	"strconv"
	"strings"
)

const minRoundModulus = uint64(4)

func (mp *metaProcessor) TryForceEpochStart(metaHdr *block.MetaBlock) {
	txBlockTxs := mp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)

	for _, tx := range txBlockTxs {
		if isEpochsFastForwardTx(tx) {
			epochs, roundsPerEpochUint := parseFastForwardArguments(tx)
			if epochs > 0 && roundsPerEpochUint > 0 {
				mp.nrEpochsChanges = epochs
				mp.roundsModulus = roundsPerEpochUint

				log.Warn("forcing epoch start", "round", metaHdr.GetRound(), "epoch", metaHdr.GetEpoch(), "still remaining epoch changes", mp.nrEpochsChanges, "rounds modulus", mp.roundsModulus)
				break
			}
		}
	}

	mp.tryFastForwardEpoch(metaHdr)
}

func isEpochsFastForwardTx(tx data.TransactionHandler) bool {
	return bytes.Equal(tx.GetRcvAddr(), vm.ValidatorSCAddress) &&
		strings.HasPrefix(string(tx.GetData()), "epochFastForward")
}

func parseFastForwardArguments(tx data.TransactionHandler) (int, uint64) {
	tokens := strings.Split(string(tx.GetData()), "@")
	if len(tokens) != 3 {
		log.Error("epochFastForward", "invalid data", string(tx.GetData()))
		return 0, 0
	}
	epochs, err := strconv.ParseInt(tokens[1], 10, 64)
	if err != nil {
		log.Error("epochFastForward", "epochs could not be parsed", tokens[1])
	}

	roundsPerEpoch, err := strconv.ParseInt(tokens[2], 10, 64)
	if err != nil {
		log.Error("epochFastForward", "rounds could not be parsed", tokens[2])
	}
	roundsPerEpochUint := uint64(roundsPerEpoch)

	if roundsPerEpochUint < minRoundModulus {
		log.Warn("epochFastForward rounds per epoch too small", "rounds", roundsPerEpoch, "minRoundModulus", minRoundModulus)
		roundsPerEpochUint = minRoundModulus
	}
	return int(epochs), roundsPerEpochUint
}

func (mp *metaProcessor) tryFastForwardEpoch(metaHdr *block.MetaBlock) {
	forceEpochTrigger, ok := mp.epochStartTrigger.(update.EpochHandler)
	if !ok {
		return
	}

	correctRoundsAndEpochs := mp.roundsModulus > 0 && metaHdr.GetRound()%mp.roundsModulus == 0 && mp.nrEpochsChanges > 0

	if !check.IfNil(forceEpochTrigger) && correctRoundsAndEpochs {
		forceEpochTrigger.ForceEpochStart(metaHdr.GetRound())
		mp.nrEpochsChanges--
		log.Debug("forcing epoch start", "round", metaHdr.GetRound(), "epoch", metaHdr.GetEpoch(), "still remaining epoch changes", mp.nrEpochsChanges, "rounds modulus", mp.roundsModulus)
	}
}
