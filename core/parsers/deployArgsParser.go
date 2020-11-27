package parsers

import (
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
)

type deployArgsParser struct {
}

// DeployArgs represents the parsed deploy arguments
type DeployArgs struct {
	Code         []byte
	VMType       []byte
	CodeMetadata vmcommon.CodeMetadata
	Arguments    [][]byte
}

// NewDeployArgsParser creates a new parser
func NewDeployArgsParser() *deployArgsParser {
	return &deployArgsParser{}
}

// ParseData parses strings of the following format:
// codeHex@vmTypeHex@codeMetadataHex@argFooHex@argBarHex...
func (parser *deployArgsParser) ParseData(data string) (*DeployArgs, error) {
	result := &DeployArgs{}

	tokens, err := tokenize(data)
	if err != nil {
		return nil, err
	}

	if len(tokens) < minNumDeployArguments {
		return nil, ErrInvalidDeployArguments
	}

	result.Code, err = parser.parseCode(tokens)
	if err != nil {
		return nil, err
	}

	result.VMType, err = parser.parseVMType(tokens)
	if err != nil {
		return nil, err
	}

	result.CodeMetadata, err = parser.parseCodeMetadata(tokens)
	if err != nil {
		return nil, err
	}

	result.Arguments, err = parser.parseArguments(tokens)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (parser *deployArgsParser) parseCode(tokens []string) ([]byte, error) {
	codeHex := tokens[indexOfCode]
	code, err := decodeToken(codeHex)
	if err != nil {
		return nil, ErrInvalidCode
	}

	return code, nil
}

func (parser *deployArgsParser) parseVMType(tokens []string) ([]byte, error) {
	vmTypeHex := tokens[indexOfVMType]
	if len(vmTypeHex) == 0 {
		return nil, ErrInvalidVMType
	}

	vmType, err := decodeToken(vmTypeHex)
	if err != nil {
		return nil, ErrInvalidVMType
	}

	return vmType, nil
}

func (parser *deployArgsParser) parseCodeMetadata(tokens []string) (vmcommon.CodeMetadata, error) {
	codeMetadataHex := tokens[indexOfCodeMetadata]
	codeMetadataBytes, err := decodeToken(codeMetadataHex)
	if err != nil {
		return vmcommon.CodeMetadata{}, ErrInvalidCodeMetadata
	}

	codeMetadata := vmcommon.CodeMetadataFromBytes(codeMetadataBytes)
	return codeMetadata, nil
}

func (parser *deployArgsParser) parseArguments(tokens []string) ([][]byte, error) {
	arguments := make([][]byte, 0)

	for i := startIndexOfConstructorArguments; i < len(tokens); i++ {
		argument, err := decodeToken(tokens[i])
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, argument)
	}

	return arguments, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (parser *deployArgsParser) IsInterfaceNil() bool {
	return parser == nil
}
