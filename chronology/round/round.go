package round

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"time"
)

type Round struct {
	index             int64
	startTimeStamp    time.Time
	roundTimeDuration time.Duration
	roundTimeDivision []time.Duration
	roundState        RoundState
}

func New(index int64, startTimeStamp time.Time, roundTimeDuration time.Duration, roundTimeDivision []time.Duration, roundState RoundState) Round {

	r := Round{index, startTimeStamp, roundTimeDuration, roundTimeDivision, roundState}
	return r
}

func (r *Round) SetIndex(index int64) {
	r.index = index
}

func (r *Round) GetIndex() int64 {
	return r.index
}

func (r *Round) SetStartTimeStamp(startTimeStamp time.Time) {
	r.startTimeStamp = startTimeStamp
}

func (r *Round) GetStartTimeStamp() time.Time {
	return r.startTimeStamp
}

func (r *Round) SetRoundTimeDuration(roundTimeDuration time.Duration) {
	r.roundTimeDuration = roundTimeDuration
}

func (r *Round) GetRoundTimeDuration() time.Duration {
	return r.roundTimeDuration
}

func (r *Round) SetRoundTimeDivision(roundTimeDivision []time.Duration) {
	r.roundTimeDivision = roundTimeDivision
}

func (r *Round) GetRoundTimeDivision() []time.Duration {
	return r.roundTimeDivision
}

func (r *Round) SetRoundState(roundState RoundState) {
	r.roundState = roundState
}

func (r *Round) GetRoundState() RoundState {
	return r.roundState
}

// A RoundState specifies in which state is the current round
type RoundState int

const (
	RS_BEFORE_ROUND RoundState = iota
	RS_START_ROUND
	RS_BLOCK
	RS_COMITMENT_HASH
	RS_BITMAP
	RS_COMITMENT
	RS_SIGNATURE
	RS_END_ROUND
	RS_AFTER_ROUND
	RS_ABORDED
	RS_UNKNOWN
)

// impl

type RoundImpl struct {
}

func (ri RoundImpl) Print(round *Round) {
	spew.Dump(round)
}

func (ri RoundImpl) CreateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration, roundTimeDivision []time.Duration) Round {

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	var r Round

	r.SetIndex(delta / roundTimeDuration.Nanoseconds())
	r.SetStartTimeStamp(genesisRoundTimeStamp.Add(time.Duration(r.GetIndex() * roundTimeDuration.Nanoseconds())))
	r.SetRoundTimeDuration(roundTimeDuration)
	r.SetRoundTimeDivision(roundTimeDivision)
	r.SetRoundState(ri.GetRoundStateFromDateTime(&r, timeStamp))

	return r
}

func (ri RoundImpl) UpdateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, round *Round) {

	if round == nil {
		fmt.Print("round should be not null")
	}

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	index := delta / round.GetRoundTimeDuration().Nanoseconds()

	if round.GetIndex() != index {
		round.SetIndex(index)
		round.SetStartTimeStamp(genesisRoundTimeStamp.Add(time.Duration(index * round.GetRoundTimeDuration().Nanoseconds())))
	}

	round.SetRoundState(ri.GetRoundStateFromDateTime(round, timeStamp))
}

func (ri RoundImpl) CreateRoundTimeDivision(duration time.Duration) []time.Duration {

	var d []time.Duration

	for i := RS_START_ROUND; i <= RS_END_ROUND; i++ {
		switch i {
		case RS_START_ROUND:
			d = append(d, time.Duration(5*duration/100))
		case RS_BLOCK:
			d = append(d, time.Duration(25*duration/100))
		case RS_COMITMENT_HASH:
			d = append(d, time.Duration(40*duration/100))
		case RS_BITMAP:
			d = append(d, time.Duration(55*duration/100))
		case RS_COMITMENT:
			d = append(d, time.Duration(70*duration/100))
		case RS_SIGNATURE:
			d = append(d, time.Duration(85*duration/100))
		case RS_END_ROUND:
			d = append(d, time.Duration(100*duration/100))
		}
	}

	return d
}

func (ri RoundImpl) GetRoundStateFromDateTime(round *Round, timeStamp time.Time) RoundState {

	if round == nil {
		return RS_UNKNOWN
	}

	delta := timeStamp.Sub(round.GetStartTimeStamp()).Nanoseconds()

	if delta < 0 {
		return RS_BEFORE_ROUND
	}

	if delta > round.GetRoundTimeDuration().Nanoseconds() {
		return RS_AFTER_ROUND
	}

	for i, v := range round.GetRoundTimeDivision() {
		if delta <= v.Nanoseconds() {
			return RS_START_ROUND + RoundState(i)
		}
	}

	return RS_UNKNOWN
}

func (ri RoundImpl) GetRoundStateName(roundState RoundState) string {

	switch roundState {
	case RS_BEFORE_ROUND:
		return ("<BEFORE_ROUND>")
	case RS_START_ROUND:
		return ("<START_ROUND>")
	case RS_BLOCK:
		return ("<BLOCK>")
	case RS_COMITMENT_HASH:
		return ("<COMITMENT_HASH>")
	case RS_BITMAP:
		return ("<BITMAP>")
	case RS_COMITMENT:
		return ("<COMITMENT>")
	case RS_SIGNATURE:
		return ("<SIGNATURE>")
	case RS_END_ROUND:
		return ("<END_ROUND>")
	case RS_AFTER_ROUND:
		return ("<AFTER_ROUND>")
	case RS_ABORDED:
		return ("<ABORDED>")
	case RS_UNKNOWN:
		return ("<UNKNOWN>")
	default:
		return ("Undifined round state")
	}
}
