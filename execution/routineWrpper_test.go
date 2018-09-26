package execution_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, execution.Started, rw.Stat())

	//wait 0.5 sec
	time.Sleep(time.Millisecond * 500)

	//counter CN01 should have been 1 by now, closing
	assert.Equal(t, int32(1), atomic.LoadInt32(&counterCN01))

	rw.Stop()
	//since go routine is still waiting, status should be CLOSING
	assert.Equal(t, execution.Closing, rw.Stat())
	//starting should not produce effects here
	rw.Start()
	assert.Equal(t, execution.Closing, rw.Stat())

	time.Sleep(time.Second)

	//it should have stopped
	assert.Equal(t, execution.Closed, rw.Stat())
}

func TestDelay(t *testing.T) {
	counterCN01 := int32(0)

	fnc := func(rw interface{}) {
		atomic.AddInt32(&counterCN01, 1)
		fmt.Printf("Variable is now: %d\n", atomic.LoadInt32(&counterCN01))
	}

	rw := execution.NewRoutineWrapper()
	rw.DurCalls = time.Second
	rw.OnDoSimpleTask = fnc
	rw.Start()

	time.Sleep(time.Millisecond * 3500)

	rw.Stop()
	time.Sleep(time.Second)

	assert.Equal(t, int32(4), atomic.LoadInt32(&counterCN01))

}
