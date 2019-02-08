package bn

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators/groupSelectors"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
)

type subroundStartRound struct {
	*subround

	blockChain             *blockchain.BlockChain
	consensusState         *spos.ConsensusState
	multiSigner            crypto.MultiSigner
	rounder                round.Rounder
	syncTimer              ntp.SyncTimer
	validatorGroupSelector groupSelectors.ValidatorGroupSelector
}

// NewSubroundStartRound creates a SubroundStartRound object
func NewSubroundStartRound(
	subround *subround,
	blockChain *blockchain.BlockChain,
	consensusState *spos.ConsensusState,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector groupSelectors.ValidatorGroupSelector,
	extend func(subroundId int),
) (*subroundStartRound, error) {

	err := checkNewSubroundStartRoundParams(
		subround,
		blockChain,
		consensusState,
		multiSigner,
		rounder,
		syncTimer,
		validatorGroupSelector,
	)

	if err != nil {
		return nil, err
	}

	srStartRound := subroundStartRound{
		subround,
		blockChain,
		consensusState,
		multiSigner,
		rounder,
		syncTimer,
		validatorGroupSelector,
	}

	srStartRound.job = srStartRound.doStartRoundJob
	srStartRound.check = srStartRound.doStartRoundConsensusCheck
	srStartRound.extend = extend

	return &srStartRound, nil
}

func checkNewSubroundStartRoundParams(
	subround *subround,
	blockChain *blockchain.BlockChain,
	consensusState *spos.ConsensusState,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector groupSelectors.ValidatorGroupSelector,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if blockChain == nil {
		return spos.ErrNilBlockChain
	}

	if consensusState == nil {
		return spos.ErrNilConsensusState
	}

	if multiSigner == nil {
		return spos.ErrNilMultiSigner
	}

	if rounder == nil {
		return spos.ErrNilRounder
	}

	if syncTimer == nil {
		return spos.ErrNilSyncTimer
	}

	if validatorGroupSelector == nil {
		return spos.ErrNilValidatorGroupSelector
	}

	return nil
}

// doStartRoundJob method is the function which actually does the job of the startRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in subround struct,
// from spos package)
func (sr *subroundStartRound) doStartRoundJob() bool {
	sr.consensusState.ResetConsensusState()

	err := sr.generateNextConsensusGroup(sr.rounder.Index())

	if err != nil {
		log.Error(err.Error())

		sr.consensusState.RoundCanceled = true

		return false
	}

	leader, err := sr.consensusState.GetLeader()

	if err != nil {
		log.Error(err.Error())
		leader = "Unknown"
	}

	msg := ""
	if leader == sr.consensusState.SelfPubKey() {
		msg = " (MY TURN)"
	}

	log.Info(fmt.Sprintf("%sStep 0: Preparing for this round with leader %s%s\n",
		sr.syncTimer.FormattedCurrentTime(), hex.EncodeToString([]byte(leader)), msg))

	pubKeys := sr.consensusState.ConsensusGroup()

	selfIndex, err := sr.consensusState.IndexSelfConsensusGroup()

	if err != nil {
		log.Info(fmt.Sprintf("%sCanceled round %d in subround %s, NOT IN THE CONSENSUS GROUP\n",
			sr.syncTimer.FormattedCurrentTime(), sr.rounder.Index(), getSubroundName(SrStartRound)))

		sr.consensusState.RoundCanceled = true

		return false
	}

	err = sr.multiSigner.Reset(pubKeys, uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())

		sr.consensusState.RoundCanceled = true

		return false
	}

	sr.consensusState.SetStatus(SrStartRound, spos.SsFinished)

	return true
}

func (sr *subroundStartRound) generateNextConsensusGroup(roundIndex int32) error {
	// TODO: replace random source with last block signature
	headerHash := sr.blockChain.CurrentBlockHeaderHash
	if sr.blockChain.CurrentBlockHeaderHash == nil {
		headerHash = sr.blockChain.GenesisHeaderHash
	}

	randomSource := fmt.Sprintf("%d-%s", roundIndex, toB64(headerHash))

	log.Info(fmt.Sprintf("random source used to determine the next consensus group is: %s\n", randomSource))

	nextConsensusGroup, err := sr.consensusState.GetNextConsensusGroup(randomSource, sr.validatorGroupSelector)

	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("consensus group for round %d is formed by next validators:\n",
		roundIndex))

	for i := 0; i < len(nextConsensusGroup); i++ {
		log.Info(fmt.Sprintf("%s", hex.EncodeToString([]byte(nextConsensusGroup[i]))))
	}

	log.Info(fmt.Sprintf("\n"))

	sr.consensusState.SetConsensusGroup(nextConsensusGroup)

	return nil
}

// doStartRoundConsensusCheck method checks if the consensus is achieved in the start subround.
func (sr *subroundStartRound) doStartRoundConsensusCheck() bool {
	return !sr.consensusState.RoundCanceled
}
