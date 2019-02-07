package interceptor

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (ti *topicInterceptor) Validator(message p2p.MessageP2P) error {
	return ti.validator(message)
}
