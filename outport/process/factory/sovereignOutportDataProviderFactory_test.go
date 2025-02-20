package factory

import (
	"testing"
)

func TestCreateSovereignOutportDataProvider(t *testing.T) {
	t.Parallel()

	factory := NewSovereignOutportDataProviderFactory()
	testCreateOutportDataProviderDisabled(t, factory)
	testCreateOutportDataProviderError(t, factory)
	testCreateOutportDataProvider(t, factory, "*process.sovereignOutportDataProvider")
}
