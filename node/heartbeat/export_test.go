package heartbeat

import "time"

func (hbmi *HeartbeatMessageInfo) SetTimeGetter(f func() time.Time) {
	hbmi.timeGetter = f
}
