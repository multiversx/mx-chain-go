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

			switch idx % 7 {
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
			case 5:
				err := tcc.put(fmt.Sprintf("key%d", idx), fmt.Sprintf("valuey%d", idx), time.Second)
				assert.Nil(t, err)
			case 6:
				_, _, err := tcc.hasOrAdd(fmt.Sprintf("key%d", idx), fmt.Sprintf("valuey%d", idx), time.Second)
				assert.Nil(t, err)
			default:
				assert.Fail(t, "test setup error, change the line 'switch idx % xxx {' from this test")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
