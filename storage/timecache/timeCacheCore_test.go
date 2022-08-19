package timecache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeCacheCore_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	tcc := newTimeCacheCore(time.Second)
	numOperations := 1000
	wg := &sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)

			switch idx % 5 {
			case 0:
				_, err := tcc.upsert(fmt.Sprintf("key%d", idx), fmt.Sprintf("valuey%d", idx), time.Second)
				assert.Nil(t, err)
			case 1:
				tcc.sweep()
			case 2:
				_ = tcc.has(fmt.Sprintf("key%d", idx))
			case 3:
				_ = tcc.len()
			case 4:
				tcc.clear()
			default:
				assert.Fail(t, "test setup error, change this line 'switch idx%6{'")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
