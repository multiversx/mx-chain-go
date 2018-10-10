package chronology

import (
	"fmt"
	"sync"
	"time"
)

type Stater interface {
	DoWork(ChronologyNew) bool
	NextStater() int
	TimeEnd() int64
	StarterName() string
}

type ChronologyNew struct {
	DoRun      bool
	DoLog      bool
	DoSyncMode bool

	round       Round
	GenesisTime time.Time

	SelfRoundState RoundState
	//ChRcvBlk       chan *block.Block
	ClockOffset time.Duration

	mutAvailable         sync.RWMutex
	Staters              []RoundState
	AvailableRoundStates map[RoundState]Stater
	//OnNeedToBroadcastBlock func(block *block.Block) bool
}

func (c *ChronologyNew) Round() int {
	return c.round.Index
}

func (c *ChronologyNew) State() int {
	return int(c.SelfRoundState)
}

func (c *ChronologyNew) StartRounds() {
	for c.DoRun {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		_, roundState := c.UpdateRound()

		rs := c.LoadRoundStater(roundState)

		if rs == nil {
			c.Log("nil new round state")
			return
		}

		if c.SelfRoundState == roundState {
			if rs.DoWork(c) {
				c.SelfRoundState = round.RoundState(rs.NextStater())
			}
		}
	}

}

func (c *ChronologyNew) LoadRoundStater(state round.RoundState) Stater {
	c.mutAvailable.RLock()
	defer c.mutAvailable.RUnlock()

	val, ok := c.AvailableRoundStates[state]

	if !ok {
		return nil
	}

	return val
}

func (c *ChronologyNew) UpdateRound() (int, round.RoundState) {
	oldRoundIndex := c.Round.Index
	oldRoundState := c.Round.State

	c.UpdateRoundFromDateTimeNew(c.GenesisTime, c.GetCurrentTime())

	if oldRoundIndex != c.Round.Index {
		//c.Statistic.AddRound() // only for statistic

		c.Log("new round")
	}

	if oldRoundState != c.Round.State {
		c.Log(fmt.Sprintf("\n" + c.GetFormatedCurrentTime() + ".................... SUBROUND " + c.Round.GetRoundStateName(c.Round.State) + " BEGINS ....................\n"))
	}

	roundState := c.SelfRoundState

	if c.DoSyncMode {
		roundState = c.Round.State
	}

	return c.Round.Index, roundState
}

func (c *ChronologyNew) FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}

func (c *ChronologyNew) GetCurrentTime() time.Time {
	return time.Now().Add(c.ClockOffset)
}

func (c *ChronologyNew) GetFormatedCurrentTime() string {
	return c.FormatTime(c.GetCurrentTime())
}

func (c *ChronologyNew) Log(message string) {
	if c.DoLog {
		fmt.Printf(message + "\n")
	}
}

func (c *ChronologyNew) UpdateRoundFromDateTimeNew(genesisRoundTimeStamp time.Time, timeStamp time.Time) {

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	index := int(delta / c.Round.TimeDuration.c.Round.TimeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(index) * c.Round.TimeDuration.Nanoseconds()))
	Nanoseconds())

	if c.Round.Index != index {
		c.Round.Index = index
		c.Round.TimeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(index) * c.Round.TimeDuration.Nanoseconds()))
	}

	c.Round.State = c.GetRoundStateFromDateTime(timeStamp)
}

func (c *ChronologyNew) GetRoundStateFromDateTime(timeStamp time.Time) int {

	delta := timeStamp.Sub(c.Round.TimeStamp).Nanoseconds()

	if delta < 0 {
		return -1
	}

	if delta > c.Round.TimeDuration.Nanoseconds() {
		return -1
	}

	for _, k := range c.Staters {
		val, ok := c.AvailableRoundStates[c.Staters[k]]
		if ok {
			if delta <= val.TimeEnd() {
				return 0 + k
			}
		}

	}

	return -1
}
