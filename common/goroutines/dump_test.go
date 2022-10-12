package goroutines

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGoRoutines(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	numGoRoutines := 1000
	for i := 0; i < numGoRoutines; i++ {
		go func() {
			<-ctx.Done()
		}()
	}

	val := GetGoRoutines()
	assert.Equal(t, numGoRoutines, strings.Count(val, "TestGetGoRoutines.func1()"))

	cancel()
}
