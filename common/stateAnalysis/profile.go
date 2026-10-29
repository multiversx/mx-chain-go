package stateAnalysis

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

type profiler struct {
	dir          string
	everyNEpochs int
	numBlocks    int
	cpuFile      *os.File
	blocksLeft   int
	lastEpoch    int64 // last epoch for which a profile was attempted, -1 when none
}

func (p *profiler) maybeStart(epoch uint32, nonce uint64) {
	if p.everyNEpochs <= 0 || p.numBlocks <= 0 || p.cpuFile != nil {
		return
	}
	if int(epoch)%p.everyNEpochs != 0 {
		return
	}
	if int64(epoch) == p.lastEpoch {
		return // already attempted for this epoch
	}
	p.lastEpoch = int64(epoch)

	err := os.MkdirAll(p.dir, 0o755)
	if err != nil {
		log.Warn("stateAnalysis: cannot create profiles dir", "err", err)
		return
	}
	name := filepath.Join(p.dir, fmt.Sprintf("cpu_epoch_%d_nonce_%d.pprof", epoch, nonce))
	f, err := os.Create(name)
	if err != nil {
		log.Warn("stateAnalysis: cannot create cpu profile", "err", err)
		return
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.Warn("stateAnalysis: cannot start cpu profile", "err", err)
		_ = f.Close()
		_ = os.Remove(name)
		return
	}
	p.cpuFile = f
	p.blocksLeft = p.numBlocks
}

func (p *profiler) blockDone(epoch uint32, nonce uint64) {
	if p.cpuFile == nil {
		return
	}
	p.blocksLeft--
	if p.blocksLeft > 0 {
		return
	}
	p.stop()

	name := filepath.Join(p.dir, fmt.Sprintf("heap_epoch_%d_nonce_%d.pprof", epoch, nonce))
	f, err := os.Create(name)
	if err != nil {
		log.Warn("stateAnalysis: cannot create heap profile", "err", err)
		return
	}
	runtime.GC()
	err = pprof.WriteHeapProfile(f)
	if err != nil {
		log.Warn("stateAnalysis: cannot write heap profile", "err", err)
	}
	_ = f.Close()
}

func (p *profiler) stop() {
	if p.cpuFile == nil {
		return
	}
	pprof.StopCPUProfile()
	_ = p.cpuFile.Close()
	p.cpuFile = nil
	p.blocksLeft = 0
}
