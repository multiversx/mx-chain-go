package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullDiscoverer_ApplyContext(t *testing.T) {
	t.Parallel()

	nd := NewNullDiscoverer()

	assert.Nil(t, nd.ApplyContext(nil))
}

func TestNullDiscoverer_Bootstrap(t *testing.T) {
	t.Parallel()

	nd := NewNullDiscoverer()

	assert.Nil(t, nd.Bootstrap())
}
