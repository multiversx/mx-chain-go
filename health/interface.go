package health

// diagnosable is an internal interface, which external components can implement in order to be "diagnosed" by the health service
type diagnosable interface {
	Diagnose(deep bool)
}

// record in an internal interface, implemented by various health records (e.g. "memoryRecord")
type record interface {
	getFilename() string
	save() error
	isMoreImportantThan(otherRecord record) bool
}
