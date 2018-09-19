package consensus

import (
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type ConsensusServiceImpl struct {
}

func (c ConsensusServiceImpl) ComputeLeader(nodes []string, round *chronology.Round) (string, error) {

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

func (c ConsensusServiceImpl) IsNodeLeader(node string, nodes []string, round *chronology.Round) (bool, error) {

	v, err := c.ComputeLeader(nodes, round)

	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return v == node, nil
}
