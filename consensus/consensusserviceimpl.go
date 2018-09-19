package consensus

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type ConsensusServiceImpl struct {
}

func (ConsensusServiceImpl) computeLeader(nodes []string, round *chronology.Round) (string, error) {

	if round == nil {
		return "", errors.New("Round is null")
	}

	if nodes == nil {
		return "", errors.New("List of nodes is null")
	}

	if len(nodes) == 0 {
		return "", errors.New("List of nodes is empty")
	}

	index := round.GetIndex() % int64(len(nodes))
	return nodes[index], nil
}
