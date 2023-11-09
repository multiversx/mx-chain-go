package goroutine

import (
	"bytes"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/debug/mock"
	"github.com/stretchr/testify/assert"
)

func TestGoRoutinesAnalyser_NewGoRoutinesAnalyser(t *testing.T) {
	t.Parallel()

	grpm := &mock.GoRoutineProcessorStub{
		ProcessGoRoutineBufferCalled: func(previousData map[string]mock.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]mock.GoRoutineHandlerMap {
			return make(map[string]mock.GoRoutineHandlerMap)
		},
	}
	analyser, err := NewGoRoutinesAnalyser(grpm)
	assert.NotNil(t, analyser)
	assert.Nil(t, err)
}

func TestGoRoutinesAnalyser_NewGoRoutinesAnalyserNilProcessorShouldErr(t *testing.T) {
	t.Parallel()

	analyser, err := NewGoRoutinesAnalyser(nil)
	assert.Nil(t, analyser)
	assert.Equal(t, core.ErrNilGoRoutineProcessor, err)
}

func TestGoRoutinesAnalyser_DumpGoRoutinesToLogWithTypesRoutineCountOk(t *testing.T) {
	t.Parallel()

	grpm := &mock.GoRoutineProcessorStub{}
	analyser, _ := NewGoRoutinesAnalyser(grpm)

	grpm.ProcessGoRoutineBufferCalled =
		func(previousData map[string]mock.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]mock.GoRoutineHandlerMap {
			newMap := make(map[string]mock.GoRoutineHandlerMap)
			newMap[newRoutine] = map[string]debug.GoRoutineHandler{
				"1": &goRoutineData{
					id:              "1",
					firstOccurrence: time.Now(),
					stackTrace:      "stack",
				},
			}

			return newMap
		}
	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 1, len(analyser.goRoutinesData[newRoutine]))
	assert.Equal(t, 0, len(analyser.goRoutinesData[oldRoutine]))

	grpm.ProcessGoRoutineBufferCalled =
		func(previousData map[string]mock.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]mock.GoRoutineHandlerMap {
			newMap := make(map[string]mock.GoRoutineHandlerMap)
			newMap[oldRoutine] = map[string]debug.GoRoutineHandler{
				"1": &goRoutineData{
					id:              "1",
					firstOccurrence: time.Now(),
					stackTrace:      "stack",
				},
			}

			return newMap
		}
	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 0, len(analyser.goRoutinesData[newRoutine]))
	assert.Equal(t, 1, len(analyser.goRoutinesData[oldRoutine]))

	grpm.ProcessGoRoutineBufferCalled =
		func(previousData map[string]mock.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]mock.GoRoutineHandlerMap {
			newMap := make(map[string]mock.GoRoutineHandlerMap)
			return newMap
		}

	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 0, len(analyser.goRoutinesData[newRoutine]))
	assert.Equal(t, 0, len(analyser.goRoutinesData[oldRoutine]))
}

func Benchmark_GetGoroutineId(b *testing.B) {
	goRoutineString := "goroutine 13 [runnable]:" +
		"github.com/libp2p/go-cidranger/net.NetworkNumberMask.Mask(0x15, 0xc000d26c60, 0x4, 0x4, 0x10, 0x10, 0xc000d26c70)" +
		"\t/home/radu/go/pkg/mod/github.com/libp2p/go-cidranger@v1.1.0/net/ip.go:276 +0xe5" +
		"github.com/libp2p/go-cidranger/net.Network.Masked(...)" +
		"\t/home/radu/go/pkg/mod/github.com/libp2p/go-cidranger@v1.1.0/net/ip.go:190" +
		"github.com/libp2p/go-cidranger.newPathprefixTrie(0xc000d26c60, 0x4, 0x4, 0x30, 0x15, 0x4)" +
		"\t/home/radu/go/pkg/mod/github.com/libp2p/go-cidranger@v1.1.0/trie.go:82 +0x19f"

	for i := 0; i < b.N; i++ {
		_ = getGoroutineId(goRoutineString)
	}
}
