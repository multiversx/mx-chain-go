package epoch

import "time"

type Epoch struct {
	Index       int
	GenesisTime time.Time
}

func New(index int, genesisTime time.Time) Epoch {
	e := Epoch{index, genesisTime}
	return e
}
