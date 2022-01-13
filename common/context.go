package common

import "context"

// WasContextClosed will return true if the provided context has been closed
func WasContextClosed(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
