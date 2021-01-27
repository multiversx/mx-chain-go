package process

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// GetElasticTemplatesAndPolicies will return elastic templates and policies
func GetElasticTemplatesAndPolicies(path string, useKibana bool) (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	indexTemplates := make(map[string]*bytes.Buffer)
	indexPolicies := make(map[string]*bytes.Buffer)
	var err error

	indexes := []string{openDistroIndex, txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, accountsESDTIndex,
		validatorsIndex, accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex}
	for _, index := range indexes {
		indexTemplates[index], err = getTemplateByIndex(path, index)
		if err != nil {
			return nil, nil, err
		}
	}

	if useKibana {
		indexesPolicies := []string{txPolicy, blockPolicy, miniblocksPolicy, ratingPolicy, roundPolicy,
			validatorsPolicy, accountsHistoryPolicy, accountsESDTHistoryPolicy, receiptsPolicy, scResultsPolicy}
		for _, indexPolicy := range indexesPolicies {
			indexPolicies[indexPolicy], err = getPolicyByIndex(path, indexPolicy)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return indexTemplates, indexPolicies, nil
}

func getTemplateByIndex(path string, index string) (*bytes.Buffer, error) {
	indexTemplate := &bytes.Buffer{}

	fileName := fmt.Sprintf("%s.json", index)
	filePath := filepath.Join(path, fileName)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("getTemplateByIndex: %w, path %s, error %s", ErrReadTemplatesFile, filePath, err.Error())
	}

	indexTemplate.Grow(len(fileBytes))
	_, err = indexTemplate.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("getTemplateByIndex: %w, path %s, error %s", ErrWriteToBuffer, filePath, err.Error())
	}

	return indexTemplate, nil
}

func getPolicyByIndex(path string, index string) (*bytes.Buffer, error) {
	indexPolicy := &bytes.Buffer{}

	fileName := fmt.Sprintf("%s.json", index)
	filePath := filepath.Join(path, fileName)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("getPolicyByIndex: %w, path %s, error %s", ErrReadPolicyFile, filePath, err.Error())
	}

	indexPolicy.Grow(len(fileBytes))
	_, err = indexPolicy.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("getPolicyByIndex: %w, path %s, error %s", ErrWriteToBuffer, filePath, err.Error())
	}

	return indexPolicy, nil
}

func getDecodedResponseMultiGet(response objectsMap) map[string]bool {
	founded := make(map[string]bool)
	interfaceSlice, ok := response["docs"].([]interface{})
	if !ok {
		return founded
	}

	for _, element := range interfaceSlice {
		obj := element.(objectsMap)
		_, ok = obj["error"]
		if ok {
			continue
		}
		founded[obj["_id"].(string)] = obj["found"].(bool)
	}

	return founded
}
