package statistic

import "sync"

type ChronologyStatistic struct {
	mut              sync.RWMutex
	Rounds           int
	RoundsWithBlocks int
}

func NewChronologyStatistic() ChronologyStatistic {
	c := ChronologyStatistic{Rounds: 0, RoundsWithBlocks: 0}
	return c
}

func (c *ChronologyStatistic) AddRound() {
	c.mut.Lock()
	c.Rounds++
	c.mut.Unlock()
}

func (c *ChronologyStatistic) AddRoundWithBlock() {
	c.mut.Lock()
	c.RoundsWithBlocks++
	c.mut.Unlock()
}

func (c *ChronologyStatistic) GetRounds() int {
	c.mut.RLock()
	rounds := c.Rounds
	c.mut.RUnlock()
	return rounds
}

func (c *ChronologyStatistic) GetRoundsWithBlock() int {
	c.mut.RLock()
	roundsWithBlocks := c.RoundsWithBlocks
	c.mut.RUnlock()
	return roundsWithBlocks
}
