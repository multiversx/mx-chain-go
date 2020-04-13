package genesis

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialBalance_MarshalUnmarshalJSON(t *testing.T) {
	t.Parallel()

	initialBalances := []InitialBalance{
		createBalanceGenesisEntry(),
		createBalanceAndStakingGenesisEntry(),
		createBalanceDelegationGenesisEntry(),
		createDelegationGenesisEntry(),
		createStakingGenesisEntry(),
	}

	buff, err := json.MarshalIndent(initialBalances, "", "    ")
	assert.Nil(t, err)

	fmt.Println(string(buff))

	recoveredInitialBalances := make([]InitialBalance, 0)
	err = json.Unmarshal(buff, &recoveredInitialBalances)

	assert.Equal(t, initialBalances, recoveredInitialBalances)
}

func createBalanceGenesisEntry() InitialBalance {
	return InitialBalance{
		Address: "erd1zqgf70wh53n8u3vvgc36pufdf7jujctqphpvztm5vx8808na8xxqsrurdg",
		Balance: "1000000000000000000000",
	}
}

func createBalanceAndStakingGenesisEntry() InitialBalance {
	return InitialBalance{
		Address:        "erd12sjxt9m6qppnj7n6kqx4cwctylq9dcazp2hnvmg2a7gucy65n89sqk69m3",
		Supply:         "3000000000000000000000",
		Balance:        "1000000000000000000000",
		StakingBalance: "2000000000000000000000",
	}
}

func createBalanceDelegationGenesisEntry() InitialBalance {
	return InitialBalance{
		Address: "erd12sjxt9m6qppnj7n6kqx4cwctylq9dcazp2hnvmg2a7gucy65n89sqk69m3",
		Supply:  "1500000000000000000000",
		Balance: "1000000000000000000000",
		Delegation: &DelegationData{
			Address: "erd1kg6cgxkjxx4yn6uad38szh24hfjs4460xdlgj3wh7pstm7lw8gcs4npsz7",
			Value:   "500000000000000000000",
		},
	}
}

func createDelegationGenesisEntry() InitialBalance {
	return InitialBalance{
		Address: "erd12sjxt9m6qppnj7n6kqx4cwctylq9dcazp2hnvmg2a7gucy65n89sqk69m3",
		Supply:  "500000000000000000000",
		Balance: "0",
		Delegation: &DelegationData{
			Address: "erd1kg6cgxkjxx4yn6uad38szh24hfjs4460xdlgj3wh7pstm7lw8gcs4npsz7",
			Value:   "500000000000000000000",
		},
	}
}

func createStakingGenesisEntry() InitialBalance {
	return InitialBalance{
		Address:        "erd12sjxt9m6qppnj7n6kqx4cwctylq9dcazp2hnvmg2a7gucy65n89sqk69m3",
		Supply:         "2000000000000000000000",
		Balance:        "0",
		StakingBalance: "2000000000000000000000",
	}
}
