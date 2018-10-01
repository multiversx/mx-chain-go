package consensus

import "github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"

type IConsensusService interface {
	ComputeLeader(nodes []string, round *round.Round) (string, error)
	IsNodeLeader(node string, nodes []string, round *round.Round) (bool, error)
}
