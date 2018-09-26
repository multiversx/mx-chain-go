package execution_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestRoutineWrapper(t *testing.T) {
	counterCN01 := int32(0)

	fnc := func(rw interface{}) {
		atomic.AddInt32(&counterCN01, 1)

		time.Sleep(time.Second)
	}

	rw := execution.NewRoutineWrapper()

	rw.OnDoSimpleTask = fnc

	rw.Start()

	assert.Equal(t, execution.STARTED, rw.Stat())

	//wait 0.5 sec
	time.Sleep(time.Millisecond * 500)

	//counter CN01 should have been 1 by now, closing
	assert.Equal(t, int32(1), atomic.LoadInt32(&counterCN01))

	rw.Stop()
	//since go routine is still waiting, status should be CLOSING
	assert.Equal(t, execution.CLOSING, rw.Stat())
	//starting should not produce effects here
	rw.Start()
	assert.Equal(t, execution.CLOSING, rw.Stat())

	time.Sleep(time.Second)

	//it should have stopped
	assert.Equal(t, execution.CLOSED, rw.Stat())
}
