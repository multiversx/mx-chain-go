package replaymetrics

import "time"

type Scope struct {
	attempt *Attempt
	kind    Kind
}

func Capture(kind Kind) Scope {
	if kind == Untracked {
		return Scope{}
	}
	r := active.Load()
	if r == nil {
		return Scope{}
	}
	return Scope{attempt: r.current.Load(), kind: kind}
}

func AttachTrie(tr any, kind Kind) {
	if kind == Untracked {
		return
	}
	if tracked, ok := tr.(interface{ SetReplayMetricsKind(Kind) }); ok {
		tracked.SetReplayMetricsKind(kind)
	}
}

func (s Scope) Active() bool {
	return s.attempt != nil
}

func (s Scope) Add(id Counter, value uint64) {
	if s.attempt != nil {
		s.attempt.counters[s.kind][id].Add(value)
	}
}

func (s Scope) Count(id Counter) uint64 {
	if s.attempt == nil {
		return 0
	}
	return s.attempt.counters[s.kind][id].Load()
}

func (s Scope) Start() time.Time {
	if s.attempt == nil {
		return time.Time{}
	}
	return time.Now()
}

func (s Scope) Collapse(kind Counter, timer Counter, count uint64, start time.Time) {
	if s.attempt == nil || count == 0 {
		return
	}
	ns := uint64(time.Since(start))
	s.Add(kind, count)
	s.Add(timer, ns)
	s.Add(CollapseNs, ns)
}

func (s Scope) RootOpen(empty bool, start time.Time, err error) {
	if s.attempt == nil {
		return
	}
	if empty {
		if err == nil {
			s.Add(EmptyTries, 1)
		}
		return
	}
	s.Add(RootOpenNs, uint64(time.Since(start)))
	if err != nil {
		s.Add(RootOpenErrors, 1)
	} else {
		s.Add(RootOpens, 1)
	}
}

func (s Scope) TrieBytes(size int) {
	switch s.kind {
	case User:
		s.peak(mainBytes, uint64(max(size, 0)))
	case Peer:
		s.peak(peerBytes, uint64(max(size, 0)))
	}
}

func (s Scope) CacheSize(entries int, size uint64, pending int, pendingSize uint64) {
	s.peak(dataCacheEntries, uint64(entries))
	s.peak(dataCacheBytes, size)
	s.peak(pendingDirtyEntries, uint64(pending))
	s.peak(pendingDirtyBytes, pendingSize)
}

func (s Scope) peak(metric gauge, value uint64) {
	if s.attempt == nil {
		return
	}
	field := &s.attempt.peaks[metric]
	for previous := field.Load(); value > previous; previous = field.Load() {
		if field.CompareAndSwap(previous, value) {
			return
		}
	}
}
