package multicall

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func TestMulticallData(t *testing.T) {
	// Generate some dummy call data
	call1 := []byte{0x01, 0x02, 0x03, 0x04}
	call2 := []byte{0x05, 0x06, 0x07, 0x08}

	callsData := [][]byte{call1, call2}

	data, err := MulticallData(callsData)
	if err != nil {
		t.Fatalf("Failed to generate multicall data: %v", err)
	}

	// Verify the function selector (first 4 bytes)
	parsedABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		t.Fatalf("Failed to parse ABI: %v", err)
	}

	expectedSelector := parsedABI.Methods["multicall"].ID
	if len(data) < 4 || string(data[:4]) != string(expectedSelector) {
		t.Errorf("Invalid function selector. Expected: %x, Got: %x", expectedSelector, data[:4])
	}
}
