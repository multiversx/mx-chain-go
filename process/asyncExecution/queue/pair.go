package queue

import "github.com/multiversx/mx-chain-core-go/data"

// HeaderBodyPair groups a block header and its corresponding body.
type HeaderBodyPair struct {
	Header data.HeaderHandler
	Body   data.BodyHandler
}
