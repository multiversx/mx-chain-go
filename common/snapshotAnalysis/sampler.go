package snapshotAnalysis

import (
	"runtime"
	"time"
)

// LinkStorer associates a trie storage manager identifier with the meter of its main storer, so the
// snapshot rows and the timeline can report the storage counters of the right storer
func LinkStorer(identifier string, m *StorerMeter) {
	if m == nil {
		return
	}
	c := global
	if c == nil {
		return
	}
	c.storers.Store(m.identifier, m)
	c.mut.Lock()
	if c.links == nil {
		c.links = make(map[string]*StorerMeter)
	}
	c.links[identifier] = m
	c.mut.Unlock()
}

func (c *collector) linkedStorer(identifier string) *StorerMeter {
	c.mut.Lock()
	defer c.mut.Unlock()

	return c.links[identifier]
}

func (s *Snapshot) startSampler() {
	if s == nil || s.c == nil {
		return
	}
	s.c.samplers.Add(1)
	go s.runSampler()
}

// stopSampler stops the sampler and waits for the final row to be written
func (s *Snapshot) stopSampler() {
	if s == nil || s.c == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	<-s.samplerDone
}

func (s *Snapshot) runSampler() {
	c := s.c
	defer c.samplers.Done()
	defer close(s.samplerDone)
	defer c.forget(s)

	interval := c.cfg.sampleInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	prev := s.timelineState(time.Now())
	for {
		select {
		case <-s.stopCh:
			s.writeFinalRows(prev)
			return
		case now := <-ticker.C:
			prev = s.sample(prev, now)
			deadline := s.afterDeadline.Load()
			if s.completed.Load() && deadline != 0 && now.UnixNano() >= deadline {
				s.writeSnapshotRow(now)
				return
			}
		}
	}
}

func (s *Snapshot) sample(prev timelineState, now time.Time) timelineState {
	cur := s.timelineState(now)
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	err := s.c.timelineFile.write(s.c.timelineRow(s, prev, cur, ms.HeapAlloc).vals)
	if err != nil {
		log.Warn("snapshotAnalysis: cannot write timeline row", "error", err)
	}
	return cur
}

func (s *Snapshot) writeFinalRows(prev timelineState) {
	now := time.Now()
	if now.Sub(prev.at) > time.Millisecond {
		_ = s.sample(prev, now)
	}
	if !s.completed.Load() {
		s.Skipped("stopped before completion")
	}
	s.writeSnapshotRow(now)
}

func (s *Snapshot) writeSnapshotRow(now time.Time) {
	err := s.c.snapshotsFile.write(s.c.snapshotRow(s, now).vals)
	if err != nil {
		log.Warn("snapshotAnalysis: cannot write snapshot row", "error", err)
	}
}
