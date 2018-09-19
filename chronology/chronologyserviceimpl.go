package chronology

import (
	"fmt"
	"time"
)

type ChronologyServiceImpl struct {
}

func (c ChronologyServiceImpl) GetRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration) *Round {

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	if delta < 0 {
		fmt.Print("genesisRoundTimeStamp should be lower or equal to timestamp!\n")
		return nil
	}

	var r Round

	r.SetIndex(delta / roundTimeDuration.Nanoseconds())
	r.SetStartTimeStamp(genesisRoundTimeStamp.Add(time.Duration(r.GetIndex() * roundTimeDuration.Nanoseconds())))
	r.SetRoundTimeDuration(roundTimeDuration)

	var d []time.Duration

	for i := RS_START_ROUND; i <= RS_END_ROUND; i++ {
		switch i {
		case RS_START_ROUND:
			d = append(d, time.Duration(0*r.GetRoundTimeDuration()/100))
		case RS_PROPOSE_BLOCK:
			d = append(d, time.Duration(25*r.GetRoundTimeDuration()/100))
		case RS_SEND_COMITMENT_HASH:
			d = append(d, time.Duration(40*r.GetRoundTimeDuration()/100))
		case RS_SEND_BITMAP:
			d = append(d, time.Duration(55*r.GetRoundTimeDuration()/100))
		case RS_SEND_COMITMENT:
			d = append(d, time.Duration(70*r.GetRoundTimeDuration()/100))
		case RS_SEND_AGGREGATE_COMITMENT:
			d = append(d, time.Duration(85*r.GetRoundTimeDuration()/100))
		case RS_END_ROUND:
			d = append(d, time.Duration(100*r.GetRoundTimeDuration()/100))
		}
	}

	r.SetRoundTimeDivision(d)

	return &r
}

func (c ChronologyServiceImpl) GetRoundState(round *Round, timeStamp time.Time) RoundState {

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
		if delta < v.Nanoseconds() {
			return RS_START_ROUND + RoundState(i)
		}
	}

	return RS_UNKNOWN
}

func (c ChronologyServiceImpl) GetRoundStateName(roundState RoundState) string {

	switch roundState {
	case RS_BEFORE_ROUND:
		return ("RS_BEFORE_ROUND")
	case RS_START_ROUND:
		return ("RS_START_ROUND")
	case RS_PROPOSE_BLOCK:
		return ("RS_PROPOSE_BLOCK")
	case RS_SEND_COMITMENT_HASH:
		return ("RS_SEND_COMITMENT_HASH")
	case RS_SEND_BITMAP:
		return ("RS_SEND_BITMAP")
	case RS_SEND_COMITMENT:
		return ("RS_SEND_COMITMENT")
	case RS_SEND_AGGREGATE_COMITMENT:
		return ("RS_SEND_AGGREGATE_COMITMENT")
	case RS_END_ROUND:
		return ("RS_END_ROUND")
	case RS_AFTER_ROUND:
		return ("RS_END_ROUND")
	case RS_UNKNOWN:
		return ("RS_UNKNOWN")
	default:
		return ("Undifined round state")
	}
}
