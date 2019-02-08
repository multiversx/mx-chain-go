package round

import (
	"time"
)

// Rounder is the main interface for a round
type Rounder interface {
	Index() int32
	UpdateRound(time.Time, time.Time)
	TimeStamp() time.Time
	TimeDuration() time.Duration
}
