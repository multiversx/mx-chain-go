package statistic

import "sync"

type Chronology struct {
	mut              sync.RWMutex
	rounds           int
	roundsWithBlocks int
}

func New() Chronology {
	c := Chronology{rounds: 0, roundsWithBlocks: 0}
	return c
}

func (c *Chronology) AddRound() {
	c.mut.Lock()
	c.rounds++
	c.mut.Unlock()
}

func (c *Chronology) AddRoundWithBlock() {
	c.mut.Lock()
	c.roundsWithBlocks++
	c.mut.Unlock()
}

func (c *Chronology) GetRounds() int {
	c.mut.RLock()
	rounds := c.rounds
	c.mut.RUnlock()
	return rounds
}

func (c *Chronology) GetRoundsWithBlock() int {
	c.mut.RLock()
	roundsWithBlocks := c.roundsWithBlocks
	c.mut.RUnlock()
	return roundsWithBlocks
}
