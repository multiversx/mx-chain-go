package consensus

import "github.com/ElrondNetwork/elrond-go-sandbox/chronology"

type IConsensusService interface {
	ComputeLeader(nodes []string, round *chronology.Round) (string, error)
	IsNodeLeader(node string, nodes []string, round *chronology.Round) (bool, error)
}
