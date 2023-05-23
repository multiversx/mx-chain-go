package block

import (
	"bytes"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/vm"
	"strconv"
	"strings"
)

const minRoundModulus = uint64(4)

func (mp *metaProcessor) ForceStart(metaHdr *block.MetaBlock) {
	forceEpochTrigger := mp.epochStartTrigger.(update.EpochHandler)

	txBlockTxs := mp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)

	for _, tx := range txBlockTxs {
		if bytes.Compare(tx.GetRcvAddr(), vm.ValidatorSCAddress) == 0 {
			tokens := strings.Split(string(tx.GetData()), "@")
			if len(tokens) != 3 {
				log.Info("epochfastforward", "invalid data", string(tx.GetData()))
				continue
			}
			epochs, err := strconv.ParseInt(tokens[1], 10, 64)
			if err != nil {
				log.Error("epochfastforward", "epochs could not be parsed", tokens[1])
			}

			roundsPerEpoch, err := strconv.ParseInt(tokens[2], 10, 64)
			if err != nil {
				log.Error("epochfastforward", "rounds could not be parsed", tokens[2])
			}
			roundsPerEpochUint := uint64(roundsPerEpoch)

			if roundsPerEpochUint < minRoundModulus {
				log.Warn("epochfastforward rounds per epoch too small", "rounds", roundsPerEpoch, "minRoundModulus", minRoundModulus)
				roundsPerEpochUint = minRoundModulus
			}

			mp.nrEpochsChanges = int(epochs)
			mp.roundsModulus = roundsPerEpochUint

			log.Warn("forcing epoch start", "round", metaHdr.GetRound(), "epoch", metaHdr.GetEpoch(), "still remaining epoch changes", mp.nrEpochsChanges, "rounds modulus", mp.roundsModulus)

			break
		}
	}

	if !check.IfNil(forceEpochTrigger) {
		if metaHdr.GetRound()%mp.roundsModulus == 0 && mp.nrEpochsChanges > 0 {
			forceEpochTrigger.ForceEpochStart(metaHdr.GetRound())
			mp.nrEpochsChanges--
			log.Debug("forcing epoch start", "round", metaHdr.GetRound(), "epoch", metaHdr.GetEpoch(), "still remaining epoch changes", mp.nrEpochsChanges, "rounds modulus", mp.roundsModulus)
		}
	}
}
