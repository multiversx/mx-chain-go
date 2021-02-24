package ginwebserver

import (
	"bytes"
	"reflect"

	"gopkg.in/go-playground/validator.v8"
)

type ginWriter struct {
}

func (gv *ginWriter) Write(p []byte) (n int, err error) {
	trimmed := bytes.TrimSpace(p)
	log.Trace("gin server", "message", string(trimmed))

	return len(p), nil
}

type ginErrorWriter struct {
}

func (gev *ginErrorWriter) Write(p []byte) (n int, err error) {
	trimmed := bytes.TrimSpace(p)
	log.Trace("gin server", "error", string(trimmed))

	return len(p), nil
}

// skValidator validates a secret key from user input for correctness
func skValidator(
	_ *validator.Validate,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Type,
	_ reflect.Kind,
	_ string,
) bool {
	return true
}

type validatorInput struct {
	Name      string
	Validator validator.Func
}
