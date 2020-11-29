package process

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetElasticTemplatesAndPolicies_NoKibana(t *testing.T) {
	t.Parallel()

	pathToConfig := "../../../cmd/node/config/elasticIndexTemplates/noKibana"
	indexes := []string{openDistroIndex, txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, accountsESDTIndex,
		validatorsIndex, accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex}

	indexTemplates, _, err := GetElasticTemplatesAndPolicies(pathToConfig, false)
	require.NoError(t, err)
	require.NotNil(t, indexTemplates)

	for idx := 0; idx < len(indexes); idx++ {
		_, ok := indexTemplates[indexes[idx]]
		if !ok {
			require.Fail(t, fmt.Sprintf("cannot find index template for index %s", indexTemplates[indexes[idx]]))
		}
	}
}

func TestGetElasticTemplatesAndPolicies_WithKibana(t *testing.T) {
	t.Parallel()

	pathToConfig := "../../../cmd/node/config/elasticIndexTemplates/withKibana"
	indexes := []string{openDistroIndex, txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, accountsESDTIndex,
		validatorsIndex, accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex}

	indexTemplates, indexPolicies, err := GetElasticTemplatesAndPolicies(pathToConfig, true)
	require.NoError(t, err)
	require.NotNil(t, indexTemplates)
	require.NotNil(t, indexPolicies)

	for idx := 0; idx < len(indexes); idx++ {
		_, ok := indexTemplates[indexes[idx]]
		if !ok {
			require.Fail(t, fmt.Sprintf("cannot find index template for index %s", indexTemplates[indexes[idx]]))
		}
	}

	indexPoliciesName := []string{txPolicy, blockPolicy, miniblocksPolicy, ratingPolicy, roundPolicy,
		validatorsPolicy, accountsHistoryPolicy, accountsESDTHistoryPolicy, receiptsPolicy, scResultsPolicy}
	for idx := 0; idx < len(indexPoliciesName); idx++ {
		_, ok := indexPolicies[indexPoliciesName[idx]]
		if !ok {
			require.Fail(t, fmt.Sprintf("cannot find index policy for index %s", indexTemplates[indexPoliciesName[idx]]))
		}
	}
}
