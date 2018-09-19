package p2p

type GoRoutineStat int

const (
	// go routine is not started
	CLOSED GoRoutineStat = iota
	// go routine is running
	STARTED
	// go routine will close
	CLOSING
)
