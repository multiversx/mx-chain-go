package logging

type baseLifeSpanner struct {
	lifeSpanChannel chan string
}

func newBaseLifeSpanner() *baseLifeSpanner {
	return &baseLifeSpanner{
		lifeSpanChannel: make(chan string),
	}
}

// GetChannel gets the channel associated with a log recreate event
func (sls *baseLifeSpanner) GetChannel() <-chan string {
	return sls.lifeSpanChannel
}
