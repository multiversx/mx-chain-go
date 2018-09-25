package chronology

import (
	"github.com/davecgh/go-spew/spew"
	"time"
)

type Epoch struct {
	index          int64
	startTimeStamp time.Time
}

func New(index int64, startTimeStamp time.Time) Epoch {

	e := Epoch{index, startTimeStamp}
	return e
}

func (e *Epoch) SetIndex(index int64) {
	e.index = index
}

func (e *Epoch) GetIndex() int64 {
	return e.index
}

func (e *Epoch) SetStartTimeStamp(startTimeStamp time.Time) {
	e.startTimeStamp = startTimeStamp
}

func (e *Epoch) GetStartTimeStamp() time.Time {
	return e.startTimeStamp
}

// impl

type EpochImpl struct {
}

func (ei EpochImpl) Print(epoch *Epoch) {
	spew.Dump(epoch)
}
