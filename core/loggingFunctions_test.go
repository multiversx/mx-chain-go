package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDumpGoRoutinesToLogShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not paniced %v", r))
		}
	}()

	DumpGoRoutinesToLog(0)
}
