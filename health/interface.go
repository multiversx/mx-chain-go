package health

type diagnosable interface {
	Diagnose(deep bool)
}
