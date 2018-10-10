package chronology

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
)

type state1 struct {
	Hits int
}

func (s *state1) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	fmt.Printf("State 1 DoWork, hits: %v\n", s.Hits)
	s.Hits++

	return true
}

func (s *state1) NextStater() int {
	return 1
}

func (s *state1) TimeEnd() int64 {
	return int64(time.Second * 2)
}

type state2 struct {
	Hits int
}

func (s *state2) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	fmt.Printf("State 2 DoWork, hits: %v\n", s.Hits)
	s.Hits++

	return true
}

func (s *state2) NextStater() int {
	return 2
}

func (s *state2) TimeEnd() int64 {
	return int64(time.Second * 3)
}

type state3 struct {
	Hits int
}

func (s *state3) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	fmt.Printf("State 3 DoWork, hits: %v\n", s.Hits)
	s.Hits++

	return true
}

func (s *state3) NextStater() int {
	if s.Hits >= 10 {
		return -1
	}

	return 0
}

func (s *state3) TimeEnd() int64 {
	return int64(time.Second * 4)
}

func TestExample(t *testing.T) {
	chn1 := ChronologyNew{}
	chn1.mutAvailable = sync.RWMutex{}
	chn1.AvailableRoundStates = make(map[round.RoundState]Stater)
	chn1.Staters = make([]round.RoundState, 0)

	chn1.mutAvailable.Lock()
	chn1.AvailableRoundStates[0] = &state1{}
	chn1.Staters = append(chn1.Staters, 0)
	chn1.AvailableRoundStates[1] = &state2{}
	chn1.Staters = append(chn1.Staters, 1)
	chn1.AvailableRoundStates[2] = &state3{}
	chn1.Staters = append(chn1.Staters, 2)
	chn1.mutAvailable.Unlock()

	chn1.GenesisTime = time.Now()
	chn1.Round = round.NewRoundFromDateTime(time.Now(), time.Now(), time.Second*4, []time.Duration{}, round.Subround{})
	chn1.DoRun = true
	chn1.DoSyncMode = true
	chn1.DoLog = true

	chn1.StartRounds()
}
