package goroutines

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const okData = `
goroutine 45 [chan receive]:
testing.(*T).Run(0xc000102ea0, {0x10be45c, 0x49b513}, 0x12171b0)
	/usr/local/go/src/testing/testing.go:1307 +0x375
testing.runTests.func1(0xc00014f470)
	/usr/local/go/src/testing/testing.go:1598 +0x6e
testing.tRunner(0xc000102ea0, 0xc0008bfd18)
	/usr/local/go/src/testing/testing.go:1259 +0x102
testing.runTests(0xc000123780, {0x1b47ea0, 0x3, 0x3}, {0x4b62ad, 0x1020af7, 0x0})
	/usr/local/go/src/testing/testing.go:1596 +0x43f
testing.(*M).Run(0xc000123780)
	/usr/local/go/src/testing/testing.go:1504 +0x51d
main.main()
	_testmain.go:47 +0x14b
`

const notOkData = `gibberish data`

func TestNewGoRoutine_OkValue(t *testing.T) {
	t.Parallel()

	ri, err := NewGoRoutine(okData)
	assert.Nil(t, err)
	assert.NotNil(t, ri)
	assert.Equal(t, okData, ri.String())
	assert.Equal(t, "45", ri.ID())
}

func TestNewGoRoutine_NotOkValue(t *testing.T) {
	t.Parallel()

	ri, err := NewGoRoutine(notOkData)
	assert.Equal(t, errMissingGoRoutineMarker, err)
	assert.Nil(t, ri)
}
