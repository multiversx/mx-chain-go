package profiling

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("consensus/profiling")

// ArgRoundProfiler holds the arguments needed to create a roundProfiler
type ArgRoundProfiler struct {
	FolderPath    string
	ShardID       uint32
	RoundsPerFile int64
}

type roundProfiler struct {
	folderPath    string
	shardID       uint32
	roundsPerFile int64
	roundCount    int64
	mut           sync.Mutex
	file          *os.File
}

// NewRoundProfiler creates a new round CPU profiler that writes a pprof file per N consensus rounds
func NewRoundProfiler(arg ArgRoundProfiler) (*roundProfiler, error) {
	if len(arg.FolderPath) == 0 {
		return nil, errEmptyFolderPath
	}
	if arg.RoundsPerFile < 1 {
		return nil, errInvalidRoundsPerFile
	}

	err := os.MkdirAll(arg.FolderPath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("could not create round profiling directory: %w", err)
	}

	return &roundProfiler{
		folderPath:    arg.FolderPath,
		shardID:       arg.ShardID,
		roundsPerFile: arg.RoundsPerFile,
	}, nil
}

// OnRoundStart is called on each new consensus round. It rotates the pprof file every roundsPerFile rounds.
func (rp *roundProfiler) OnRoundStart(roundIndex int64, roundTimestamp time.Time) {
	rp.mut.Lock()
	defer rp.mut.Unlock()

	rp.roundCount++

	if rp.file != nil && rp.roundCount <= rp.roundsPerFile {
		return
	}

	// time to rotate
	rp.stopProfileNoLock()
	rp.roundCount = 1

	timestamp := roundTimestamp.Format("20060102150405")
	fileName := fmt.Sprintf("round_%d_%s_%d.pprof", roundIndex, timestamp, rp.shardID)
	filePath := filepath.Join(rp.folderPath, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		log.Error("could not create round CPU profile file", "error", err, "path", filePath)
		return
	}

	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.Error("could not start CPU profile for round", "error", err, "round", roundIndex)
		_ = f.Close()
		return
	}

	rp.file = f
	log.Debug("started CPU profile for round", "round", roundIndex, "file", filePath)
}

// Close stops any in-progress CPU profile
func (rp *roundProfiler) Close() error {
	rp.mut.Lock()
	defer rp.mut.Unlock()

	rp.stopProfileNoLock()
	return nil
}

func (rp *roundProfiler) stopProfileNoLock() {
	if rp.file == nil {
		return
	}

	pprof.StopCPUProfile()
	err := rp.file.Close()
	if err != nil {
		log.Error("could not close round CPU profile file", "error", err)
	}
	rp.file = nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rp *roundProfiler) IsInterfaceNil() bool {
	return rp == nil
}
