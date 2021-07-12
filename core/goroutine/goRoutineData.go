package goroutine

import "time"

type goRoutineData struct {
	id              string
	firstOccurrence time.Time
	stackTrace      string
}

// ID - gets the unique identifier of the go routine
func (grd *goRoutineData) ID() string {
	return grd.id
}

// FirstOccurrence - gets the creation date of the go routine
func (grd *goRoutineData) FirstOccurrence() time.Time {
	return grd.firstOccurrence
}

// StackTrace - gets the call stack of the go routine
func (grd *goRoutineData) StackTrace() string {
	return grd.stackTrace
}
