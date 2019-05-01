package heartbeat

import "time"

func (pe *PubkeyElement) SetTimeGetter(f func() time.Time) {
	pe.timeGetter = f
}
