package execution

// ExecCode will retain the defined codes after execution of a task
type ExecCode int32

// ExecSummary holds the return status carried by a task
type ExecSummary struct {
	ec  ExecCode
	err error
}

// NewExecSummary creates a new ExecSummary object
func NewExecSummary(ec ExecCode, err error) *ExecSummary {
	es := ExecSummary{ec: ec, err: err}

	return &es
}

// Code returns the code associated with the return status
func (es *ExecSummary) Code() ExecCode {
	return es.ec
}

// Error returns the error associated with the return status
func (es *ExecSummary) Error() error {
	return es.err
}
