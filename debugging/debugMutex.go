package debugging

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("debugging")

const stackDepth = 3

type DebugMutex struct {
	name  string
	mutex sync.RWMutex
}

func NewDebugMutex(name string) *DebugMutex {
	return &DebugMutex{
		name: name,
	}
}

func (dm *DebugMutex) Lock() string {
	tag := dm.logAttemptAcquire("Lock")

	dm.mutex.Lock()

	log.Debug("Lock success", "tag", tag)

	return tag
}

func (dm *DebugMutex) Unlock(tag string) {
	dm.logBeforeRelease("Unlock", tag)

	dm.mutex.Unlock()
}

func (dm *DebugMutex) RLock() string {
	tag := dm.logAttemptAcquire("RLock")

	dm.mutex.RLock()

	log.Debug("RLock success", "tag", tag)

	return tag
}

func (dm *DebugMutex) RUnlock(tag string) {
	dm.logBeforeRelease("RUnlock", tag)

	dm.mutex.RUnlock()
}

func (dm *DebugMutex) logAttemptAcquire(what string) string {
	routine := goid()
	token := createTag()
	tag := fmt.Sprintf("%d:%s", routine, token)
	callersSummary := getCallersSummary(3)

	log.Debug(fmt.Sprintf("%s.%s", dm.name, what), "tag", tag, "callers", callersSummary)

	return tag
}

func (dm *DebugMutex) logBeforeRelease(what string, tag string) {
	callersSummary := getCallersSummary(3)

	log.Debug(fmt.Sprintf("%s.%s", dm.name, what), "tag", tag, "callers", callersSummary)
}

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func createTag() string {
	buffer := make([]byte, 4)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}

func getCallersSummary(skip int) string {
	var sb strings.Builder

	for i := skip; i < skip+stackDepth; i++ {
		pc, file, no, _ := runtime.Caller(i)

		f := runtime.FuncForPC(pc)
		functionName := ""
		if f != nil {
			functionName = f.Name()
			functionName = functionName[strings.LastIndex(functionName, ".")+1:]
		}

		file = filepath.Base(file)
		sb.WriteString(fmt.Sprintf("%s (%s:%d) < ", functionName, file, no))
	}

	return sb.String()
}
