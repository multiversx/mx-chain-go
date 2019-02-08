package chronology

import (
	"time"
)

// SubroundHandler defines the actions which should be handled by a subround implementation
type SubroundHandler interface {
	DoWork(func() time.Duration) bool // DoWork implements of the subround's job
	Previous() int                    // Previous returns the ID of the previous subround
	Next() int                        // Next returns the ID of the next subround
	Current() int                     // Current returns the ID of the current subround
	StartTime() int64                 // StartTime returns the start time, in the rounder time, of the current subround
	EndTime() int64                   // EndTime returns the top limit time, in the rounder time, of the current subround
	Name() string                     // Name returns the name of the current rounder
}

// ChronologyHandler defines the actions which should be handled by a chronology implementation
type ChronologyHandler interface {
	AddSubround(SubroundHandler)
	RemoveAllSubrounds()
	StartRounds()
}
