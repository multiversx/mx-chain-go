package chronology

// Rounder is the main interface for round
type Rounder interface {
	Index() int32
}

// SubroundHandler defines the actions that can be handled in a sub-round
type SubroundHandler interface {
	DoWork(func() SubroundId, func() bool) bool // DoWork implements of the subround's job
	Next() SubroundId                           // Next returns the ID of the next subround
	Current() SubroundId                        // Current returns the ID of the current subround
	EndTime() int64                             // EndTime returns the top limit time, in the round time, of the current subround
	Name() string                               // Name returns the name of the current round
	Check() bool
}
