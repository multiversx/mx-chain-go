package epoch

import "time"

type Epoch struct {
	index          int
	startTimeStamp time.Time
}

func New(index int, startTimeStamp time.Time) Epoch {

	e := Epoch{index, startTimeStamp}
	return e
}

func (e *Epoch) SetIndex(index int) {
	e.index = index
}

func (e *Epoch) GetIndex() int {
	return e.index
}

func (e *Epoch) SetStartTimeStamp(startTimeStamp time.Time) {
	e.startTimeStamp = startTimeStamp
}

func (e *Epoch) GetStartTimeStamp() time.Time {
	return e.startTimeStamp
}
