package interceptor

import (
	"context"

	"github.com/libp2p/go-libp2p-pubsub"
)

func (ti *topicInterceptor) Validator(ctx context.Context, message *pubsub.Message) bool {
	return ti.validator(ctx, message)
}
