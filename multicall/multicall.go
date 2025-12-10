package multicall

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

const abiJson = `[	{
		"inputs": [
			{
				"internalType": "bytes[]",
				"name": "data",
				"type": "bytes[]"
			}
		],
		"name": "multicall",
		"outputs": [
			{
				"internalType": "bytes[]",
				"name": "results",
				"type": "bytes[]"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	}]`

// MulticallData generates contract call data for multicall function
// callsData: array of encoded function calls
func MulticallData(callsData [][]byte) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		return nil, err
	}

	data, err := parsedABI.Pack("multicall", callsData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// EncodeMulticall is a helper function that combines multiple function calls into a single multicall
// It takes multiple function call data and returns the encoded multicall data
func EncodeMulticall(callsData ...[]byte) ([]byte, error) {
	return MulticallData(callsData)
}
