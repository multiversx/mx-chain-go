package health

import (
	"runtime"
	"time"
)

// diagnosable is an internal interface, which external components can implement in order to be "diagnosed" by the health service
type diagnosable interface {
	Diagnose(deep bool)
	IsInterfaceNil() bool
}

// record in an internal interface, implemented by various health records (e.g. "memoryUsageRecord")
type record interface {
	save() error
	delete() error
	isMoreImportantThan(otherRecord record) bool
}

// clock is an internal interface that defines time-related functions
type clock interface {
	now() time.Time
	after(d time.Duration) <-chan time.Time
}

// memory is an internal interface that defines memory-related functions
type memory interface {
	getStats() runtime.MemStats
}
