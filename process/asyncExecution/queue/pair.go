package queue

import "github.com/multiversx/mx-chain-core-go/data"

type HeaderBodyPair struct {
	Header data.HeaderHandler
	Body   data.BodyHandler
}
