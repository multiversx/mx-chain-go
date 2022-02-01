package common

import "context"

// IsContextDone will return true if the provided context signals it is done
func IsContextDone(ctx context.Context) bool {
	if ctx == nil {
		return true
	}

	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
