package process

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state"
)

// ExportUserAccountState will export the account state of a provided address
func ExportUserAccountState(accountsDB state.AccountsAdapter, identifier string, address []byte, parentDirForFiles string) error {
	code, csvHexedData, err := getCodeAndData(accountsDB, address)
	if err != nil {
		return err
	}

	err = exportCode(identifier, code, parentDirForFiles)
	if err != nil {
		return err
	}

	return exportData(identifier, csvHexedData, parentDirForFiles)
}

func getCodeAndData(accountsDB state.AccountsAdapter, address []byte) (code []byte, csvHexedData []string, err error) {
	account, err := accountsDB.GetExistingAccount(address)
	if err != nil {
		return nil, nil, fmt.Errorf("%w while getting existing data for address %s",
			err, hex.EncodeToString(address))
	}

	userAccount := account.(common.UserAccountHandler)
	codeHash := userAccount.GetCodeHash()
	if len(codeHash) > 0 {
		code, err = getCode(accountsDB, codeHash)
		if err != nil {
			return nil, nil, err
		}
	}

	rootHash := userAccount.GetRootHash()
	if len(rootHash) > 0 {
		csvHexedData, err = getData(userAccount)
		if err != nil {
			return nil, nil, err
		}
	}

	return code, csvHexedData, nil
}

func getCode(accountsDB state.AccountsAdapter, codeHash []byte) ([]byte, error) {
	code := accountsDB.GetCode(codeHash)
	if len(code) == 0 {
		return nil, fmt.Errorf("empty code for hex code hash %s", hex.EncodeToString(codeHash))
	}

	return code, nil
}

func getData(account common.UserAccountHandler) ([]string, error) {
	leavesChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}

	err := account.GetAllLeaves(leavesChannels, context.Background())
	if err != nil {
		return nil, fmt.Errorf("%w while trying to export data on hex root hash %s, address %s",
			err, hex.EncodeToString(account.GetRootHash()), hex.EncodeToString(account.AddressBytes()))
	}

	lines := make([]string, 0)
	for keyVal := range leavesChannels.LeavesChan {
		lines = append(lines, fmt.Sprintf("%s,%s",
			hex.EncodeToString(keyVal.Key()),
			hex.EncodeToString(keyVal.Value())))
	}

	err = leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, fmt.Errorf("%w while trying to export data on hex root hash %s, address %s",
			err, hex.EncodeToString(account.GetRootHash()), hex.EncodeToString(account.AddressBytes()))
	}

	return lines, nil
}

func exportCode(identifier string, code []byte, parentDirForFiles string) error {
	if len(code) == 0 {
		return nil
	}

	fileArgs := core.ArgCreateFileArgument{
		Directory:     parentDirForFiles,
		Prefix:        computePrefix("code", identifier),
		FileExtension: "wasm",
	}
	f, err := core.CreateFile(fileArgs)
	if err != nil {
		return err
	}

	_, err = f.Write(code)
	if err != nil {
		_ = f.Close()

		return err
	}

	log.Info("ExportUserAccountState.exportCode", "contract data", core.ConvertBytes(uint64(len(code))))

	return f.Close()
}

func exportData(identifier string, lines []string, parentDirForFiles string) error {
	if len(lines) == 0 {
		return nil
	}

	fileArgs := core.ArgCreateFileArgument{
		Directory:     parentDirForFiles,
		Prefix:        computePrefix("data", identifier),
		FileExtension: "hex",
	}
	f, err := core.CreateFile(fileArgs)
	if err != nil {
		return err
	}

	buffer := bytes.Buffer{}
	for _, line := range lines {
		buffer.Write([]byte(fmt.Sprintf("%s\n", line)))
	}

	_, _ = f.Write(buffer.Bytes())
	log.Info("ExportUserAccountState.exportData", "num (key,values)", len(lines))

	return f.Close()
}

func computePrefix(basePrefix string, identifier string) string {
	identifier = strings.TrimSpace(identifier)
	if len(identifier) > 0 {
		basePrefix = basePrefix + "_" + identifier
	}

	return basePrefix
}
