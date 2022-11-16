package process

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
)

// ExportUserAccountState will try to export the account state of a provided address
func ExportUserAccountState(accountsDB state.AccountsAdapter, address []byte, parentDirForFiles string) {
	code, csvHexedData, err := getCodeAndData(accountsDB, address)
	if err != nil {
		log.Error("ExportUserAccountState", "error", err)
		return
	}

	err = exportCode(code, parentDirForFiles)
	if err != nil {
		log.Error("ExportUserAccountState", "error", err)
		return
	}

	err = exportData(csvHexedData, parentDirForFiles)
	if err != nil {
		log.Error("ExportUserAccountState", "error", err)
		return
	}
}

func getCodeAndData(accountsDB state.AccountsAdapter, address []byte) (code []byte, csvHexedData []string, err error) {
	account, err := accountsDB.GetExistingAccount(address)
	if err != nil {
		return nil, nil, fmt.Errorf("%w while getting existing data for hex address %s",
			err, hex.EncodeToString(code))
	}

	userAccount := account.(state.UserAccountHandler)
	codeHash := userAccount.GetCodeHash()
	if len(codeHash) > 0 {
		code, err = getCode(accountsDB, codeHash)
		if err != nil {
			return nil, nil, err
		}
	}

	rootHash := userAccount.GetRootHash()
	if len(rootHash) > 0 {
		csvHexedData, err = getData(accountsDB, rootHash, address)
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

func getData(accountsDB state.AccountsAdapter, rootHash []byte, address []byte) ([]string, error) {
	leavesChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}

	err := accountsDB.GetAllLeaves(leavesChannels, context.Background(), rootHash)
	if err != nil {
		return nil, fmt.Errorf("%w while trying to export data on hex root hash %s, address %s",
			err, hex.EncodeToString(rootHash), hex.EncodeToString(address))
	}

	lines := make([]string, 0)
	for keyVal := range leavesChannels.LeavesChan {
		suffix := append(keyVal.Key(), address...)
		valWithoutSuffix, errTrim := keyVal.ValueWithoutSuffix(suffix)
		if errTrim != nil {
			return nil, fmt.Errorf("%w while trying to export data on hex root hash %s, address %s",
				errTrim, hex.EncodeToString(rootHash), hex.EncodeToString(address))
		}

		lines = append(lines, fmt.Sprintf("%s,%s",
			hex.EncodeToString(keyVal.Key()),
			hex.EncodeToString(valWithoutSuffix)))
	}

	err = <-leavesChannels.ErrChan
	if err != nil {
		return nil, fmt.Errorf("%w while trying to export data on hex root hash %s, address %s",
			err, hex.EncodeToString(rootHash), hex.EncodeToString(address))
	}

	return lines, nil
}

func exportCode(code []byte, parentDirForFiles string) error {
	if len(code) == 0 {
		return nil
	}

	fileArgs := core.ArgCreateFileArgument{
		Directory:     parentDirForFiles,
		Prefix:        "code",
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

func exportData(lines []string, parentDirForFiles string) error {
	if len(lines) == 0 {
		return nil
	}

	fileArgs := core.ArgCreateFileArgument{
		Directory:     parentDirForFiles,
		Prefix:        "data",
		FileExtension: "hex",
	}
	f, err := core.CreateFile(fileArgs)
	if err != nil {
		return err
	}

	for _, line := range lines {
		_, err = f.Write([]byte(fmt.Sprintf("%s\n", line)))
		if err != nil {
			_ = f.Close()

			return err
		}
	}

	log.Info("ExportUserAccountState.exportData", "num (key,values)", len(lines))

	return f.Close()
}
