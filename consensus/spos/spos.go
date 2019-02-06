package spos

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators/groupSelectors"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

var log = logger.NewDefaultLogger()

// Spos defines the data needed by spos to do the consensus in each round
type Spos struct {
	Data []byte // hold the data on which validators do the consensus
	// (could be for example a hash of the block header proposed by the leader)
	*RoundConsensus
	*RoundThreshold
	*RoundStatus

	Chr *chronology.Chronology
}

// NewSpos creates a new Spos object
func NewSpos(
	data []byte,
	rcns *RoundConsensus,
	rthr *RoundThreshold,
	rstatus *RoundStatus,
	chr *chronology.Chronology,
) *Spos {

	sps := Spos{
		Data:           data,
		RoundConsensus: rcns,
		RoundThreshold: rthr,
		RoundStatus:    rstatus,
		Chr:            chr,
	}

	return &sps
}

// IsSelfLeaderInCurrentRound method checks if the current node is leader in the current round
func (sps *Spos) IsSelfLeaderInCurrentRound() bool {
	return sps.IsNodeLeaderInCurrentRound(sps.selfPubKey)
}

// IsNodeLeaderInCurrentRound method checks if the given node is leader in the current round
func (sps *Spos) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := sps.GetLeader()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return leader == node
}

// GetLeader method gets the leader of the current round
func (sps *Spos) GetLeader() (string, error) {
	if sps.consensusGroup == nil {
		return "", ErrNilConsensusGroup
	}

	if len(sps.consensusGroup) == 0 {
		return "", ErrEmptyConsensusGroup
	}

	return sps.consensusGroup[0], nil
}

// GetNextConsensusGroup gets the new consensus group for the current round based on current eligible list and a random
// source for the new selection
func (sps *Spos) GetNextConsensusGroup(randomSource string, hasher hashing.Hasher) ([]string, error) {
	validatorsList := make([]consensus.Validator, 0)

	// TODO: Uncomment the next line and remove the next after it (eligibile list should be the entire list)
	for i := 0; i < len(sps.EligibleList()); i++ {
		//for i := 0; i < sps.ConsensusGroupSize(); i++ {
		validator, err := validators.NewValidator(big.NewInt(0), 0, []byte(sps.EligibleList()[i]))

		if err != nil {
			return nil, err
		}

		validatorsList = append(validatorsList, validator)
	}

	ihgs, err := groupSelectors.NewIndexHashedGroupSelector(sps.ConsensusGroupSize(), hasher)

	if err != nil {
		return nil, err
	}

	err = ihgs.LoadEligibleList(validatorsList)

	if err != nil {
		return nil, err
	}

	validatorsGroup, err := ihgs.ComputeValidatorsGroup([]byte(randomSource))

	if err != nil {
		return nil, err
	}

	newConsensusGroup := make([]string, 0)

	for i := 0; i < len(validatorsGroup); i++ {
		newConsensusGroup = append(newConsensusGroup, string(validatorsGroup[i].PubKey()))
	}

	return newConsensusGroup, nil
}
