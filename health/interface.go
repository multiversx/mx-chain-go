package health

type diagnosable interface {
	Diagnose(deep bool)
}

type record interface {
	getFilename() string
	save() error
	isMoreImportantThan(otherRecord record) bool
}
