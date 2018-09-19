package consensus

import "github.com/ElrondNetwork/elrond-go-sandbox/chronology"

type IConsensusService interface {
	computeLeader(nodes []string, round *chronology.Round) string
}
