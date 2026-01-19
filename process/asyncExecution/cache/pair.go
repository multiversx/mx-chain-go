package cache

import "github.com/multiversx/mx-chain-core-go/data"

// HeaderBodyPair groups a block header and its corresponding body.
type HeaderBodyPair struct {
	HeaderHash []byte
	Header     data.HeaderHandler
	Body       data.BodyHandler
}
