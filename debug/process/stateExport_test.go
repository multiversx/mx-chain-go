package process

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/integrationtests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportUserAccountState(t *testing.T) {
	t.Parallel()

	address := []byte("12345678901234567890123456789112")
	accounts := integrationtests.CreateInMemoryShardAccountsDB()
	account, err := accounts.LoadAccount(address)
	require.Nil(t, err)

	userAccount := account.(state.UserAccountHandler)
	err = userAccount.AddToBalance(big.NewInt(1))
	require.Nil(t, err)

	code := []byte("test code")
	numKeys := 100

	userAccount.SetCode(code)
	for i := 0; i < numKeys; i++ {
		err = userAccount.SaveKeyValue([]byte(fmt.Sprintf("key %d", i)), []byte(fmt.Sprintf("val %d", i)))
		require.Nil(t, err)
	}

	err = accounts.SaveAccount(userAccount)
	require.Nil(t, err)

	_, err = accounts.Commit()
	require.Nil(t, err)

	t.Run("check fetched data", func(t *testing.T) {
		recoveredCode, recoveredLines, err := getCodeAndData(accounts, address)
		assert.Nil(t, err)
		assert.Equal(t, code, recoveredCode)
		assert.Equal(t, numKeys, len(recoveredLines))

		expectedLinesMap := make(map[string]struct{})
		for i := 0; i < numKeys; i++ {
			key := hex.EncodeToString([]byte(fmt.Sprintf("key %d", i)))
			val := hex.EncodeToString([]byte(fmt.Sprintf("val %d", i)))
			line := fmt.Sprintf("%s,%s", key, val)
			expectedLinesMap[line] = struct{}{}
		}

		recoveredLinesMap := make(map[string]struct{})
		for _, line := range recoveredLines {
			recoveredLinesMap[line] = struct{}{}
		}

		assert.EqualValues(t, expectedLinesMap, recoveredLinesMap)
	})
	t.Run("test files are produced", func(t *testing.T) {
		tempDir := t.TempDir()
		contents, errReadDir := ioutil.ReadDir(tempDir)
		require.Nil(t, errReadDir)
		assert.Zero(t, len(contents))

		ExportUserAccountState(accounts, address, tempDir)

		contents, errReadDir = ioutil.ReadDir(tempDir)
		require.Nil(t, errReadDir)
		assert.Equal(t, 2, len(contents))
	})
}
