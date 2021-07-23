package logging

type baseLifeSpanner struct {
	tickChannel chan string
}

func newBaseLifeSpanner() *baseLifeSpanner {
	tickChannel := make(chan string)
	return &baseLifeSpanner{tickChannel: tickChannel}
}

// GetChannel gets the channel associated with a log recreate event
func (sls *baseLifeSpanner) GetChannel() <-chan string {
	return sls.tickChannel
}
