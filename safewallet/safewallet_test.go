package safewallet

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestCalcAddress(t *testing.T) {
	Safe := common.HexToAddress("0x41675C099F32341bf84BFc5382aF534df5C7461a")
	SafeL2 := common.HexToAddress("0x29fcB43b46531BcA003ddC8FCB67FFE91900C762")
	CompatibilityFallbackHandler := common.HexToAddress("0xfd0732Dc9E303f09fCEf3a7388Ad10A83459Ec99")
	SafeToL2Setup := common.HexToAddress("0xBD89A1CE4DDe368FFAB0eC35506eEcE0b1fFdc54")
	SafeProxyFactory := common.HexToAddress("0x4e1DCf7AD4e460CfD30791CCC4F9c8a4f820ec67")
	creationCode := `0x608060405234801561001057600080fd5b506040516101e63803806101e68339818101604052602081101561003357600080fd5b8101908080519060200190929190505050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614156100ca576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806101c46022913960400191505060405180910390fd5b806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505060ab806101196000396000f3fe608060405273ffffffffffffffffffffffffffffffffffffffff600054167fa619486e0000000000000000000000000000000000000000000000000000000060003514156050578060005260206000f35b3660008037600080366000845af43d6000803e60008114156070573d6000fd5b3d6000f3fea264697066735822122003d1488ee65e08fa41e58e888a9865554c535f2c77126a82cb4c0f917f31441364736f6c63430007060033496e76616c69642073696e676c65746f6e20616464726573732070726f7669646564`
	paymentReceiver := common.HexToAddress("0x5afe7A11E7000000000000000000000000000000")
	accounts := []common.Address{
		common.HexToAddress("0xac75c2015daf16f6463c946d32d1e15f8b8983e7"),
		common.HexToAddress("0x424a788927c197851c110f9a5c0c12d251256ecf"),
		common.HexToAddress("0x271a72ce18b80dcfa9bcbbaddfb6bc94d7c3cca0"),
	}
	threshold := big.NewInt(2)
	nonce := big.NewInt(0)

	addr, err := CalcSafeAddress(Safe,
		SafeL2,
		CompatibilityFallbackHandler,
		SafeToL2Setup,
		SafeProxyFactory,
		creationCode,
		accounts,
		threshold,
		nonce,
		paymentReceiver,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr.Hex())
	if addr.Hex() != "0xB6A5246A7952E50956a5C955095a8C78473EFAf4" {
		t.Fatal("not equal")
	}
}

func TestGetSafeTransactionHash(t *testing.T) {
	// Test parameters
	safeAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	chainId := big.NewInt(1) // Ethereum mainnet

	// Create a sample transaction
	txData := SafeTransactionData{
		To:             common.HexToAddress("0xabcdef0123456789abcdef0123456789abcdef01"),
		Value:          big.NewInt(1000000000000000000), // 1 ETH
		Data:           []byte{1, 2, 3, 4},
		Operation:      0, // Call operation
		SafeTxGas:      big.NewInt(21000),
		BaseGas:        big.NewInt(10000),
		GasPrice:       big.NewInt(1000000000), // 1 Gwei
		GasToken:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
		RefundReceiver: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Nonce:          big.NewInt(42),
	}

	// Calculate the transaction hash
	txHash, err := GetSafeTransactionHash(safeAddress, chainId, txData)
	if err != nil {
		t.Fatalf("Failed to get transaction hash: %v", err)
	}

	// Verify the hash is not empty
	if txHash == (common.Hash{}) {
		t.Errorf("Transaction hash is empty")
	}

	// Print the hash for debugging
	t.Logf("Transaction hash: %s", txHash.Hex())
}

func TestSignSafeTransaction(t *testing.T) {
	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Test parameters
	safeAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	chainId := big.NewInt(1) // Ethereum mainnet

	// Create a sample transaction
	txData := SafeTransactionData{
		To:             common.HexToAddress("0xabcdef0123456789abcdef0123456789abcdef01"),
		Value:          big.NewInt(1000000000000000000), // 1 ETH
		Data:           []byte{1, 2, 3, 4},
		Operation:      0, // Call operation
		SafeTxGas:      big.NewInt(21000),
		BaseGas:        big.NewInt(10000),
		GasPrice:       big.NewInt(1000000000), // 1 Gwei
		GasToken:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
		RefundReceiver: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Nonce:          big.NewInt(42),
	}

	// Sign the transaction
	signature, err := SignSafeTransaction(safeAddress, chainId, txData, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Verify the signature length
	if len(signature) != 65 {
		t.Errorf("Wrong signature length. Expected: 65, Got: %d", len(signature))
	}

	// Verify the signature can be used to recover the signer address
	txHash, err := GetSafeTransactionHash(safeAddress, chainId, txData)
	if err != nil {
		t.Fatalf("Failed to get transaction hash: %v", err)
	}

	recoveredAddress, err := GetSignerAddress(txHash, signature)
	if err != nil {
		t.Fatalf("Failed to recover signer address: %v", err)
	}

	expectedAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	if recoveredAddress != expectedAddress {
		t.Errorf("Wrong recovered address. Expected: %s, Got: %s", expectedAddress.Hex(), recoveredAddress.Hex())
	}
}

func TestFormatSignatureBytes(t *testing.T) {
	// Generate test signatures
	sig1 := make([]byte, 65)
	sig1[0] = 1
	sig1[32] = 2
	sig1[64] = 27

	sig2 := make([]byte, 65)
	sig2[0] = 3
	sig2[32] = 4
	sig2[64] = 28

	// Format the signatures
	formattedSigs := FormatSignatureBytes(sig1, sig2)

	// Verify the formatted signatures
	expectedLength := 65 * 2
	if len(formattedSigs) != expectedLength {
		t.Errorf("Wrong formatted signatures length. Expected: %d, Got: %d", expectedLength, len(formattedSigs))
	}

	// Verify the first signature
	if formattedSigs[0] != 1 || formattedSigs[32] != 2 || formattedSigs[64] != 27 {
		t.Errorf("First signature was not correctly formatted")
	}

	// Verify the second signature
	if formattedSigs[65] != 3 || formattedSigs[65+32] != 4 || formattedSigs[65+64] != 28 {
		t.Errorf("Second signature was not correctly formatted")
	}
}

func TestGetSignerAddress(t *testing.T) {
	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create a test hash
	testHash := crypto.Keccak256Hash([]byte("test message"))

	// Sign the hash
	signature, err := SignSafeTransactionHash(testHash, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign hash: %v", err)
	}

	// Recover the signer address
	recoveredAddress, err := GetSignerAddress(testHash, signature)
	if err != nil {
		t.Fatalf("Failed to recover signer address: %v", err)
	}

	// Verify the recovered address matches the expected address
	expectedAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	if recoveredAddress != expectedAddress {
		t.Errorf("Wrong recovered address. Expected: %s, Got: %s", expectedAddress.Hex(), recoveredAddress.Hex())
	}
}

func TestGetNonceData(t *testing.T) {
	// Generate contract call data
	data, err := GetNonceData()
	if err != nil {
		t.Fatalf("Failed to generate nonce call data: %v", err)
	}

	// Check function selector (first 4 bytes)
	// Function selector for nonce()
	parsedABI, err := abi.JSON(strings.NewReader(ABI_SafeWallet))
	if err != nil {
		t.Fatalf("Failed to parse ABI: %v", err)
	}
	expectedSelector := parsedABI.Methods["nonce"].ID

	actualSelector := hex.EncodeToString(data[:4])

	if !bytes.Equal(expectedSelector, data[:4]) {
		t.Errorf("Wrong function selector. Expected: %s, Got: %s", expectedSelector, actualSelector)
	}

	// Check that the data length is correct (should only be the selector for a function with no params)
	if len(data) != 4 {
		t.Errorf("Wrong data length. Expected: 4, Got: %d", len(data))
	}
}

func TestParseNonceResult(t *testing.T) {
	// Create a mock result
	// This represents a nonce of 42 encoded according to the ABI
	// 0x000000000000000000000000000000000000000000000000000000000000002a
	mockResult, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000002a")

	// Parse the result
	nonce, err := ParseNonceResult(mockResult)
	if err != nil {
		t.Fatalf("Failed to parse nonce result: %v", err)
	}

	// Check that the parsed nonce is correct
	expectedNonce := big.NewInt(42)
	if nonce.Cmp(expectedNonce) != 0 {
		t.Errorf("Wrong nonce value. Expected: %s, Got: %s", expectedNonce.String(), nonce.String())
	}
}

func TestExecTransactionData(t *testing.T) {
	// Test parameters
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	value := big.NewInt(1000000000000000000) // 1 ETH
	txData := []byte{1, 2, 3, 4}
	operation := uint8(0) // Call operation
	safeTxGas := big.NewInt(21000)
	baseGas := big.NewInt(10000)
	gasPrice := big.NewInt(1000000000)                                            // 1 Gwei
	gasToken := common.HexToAddress("0x0000000000000000000000000000000000000000") // ETH
	refundReceiver := common.HexToAddress("0x0000000000000000000000000000000000000000")
	// Example signature (r, s, v) - in real scenario this would be generated from private key
	signatures := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65}

	// Generate contract call data
	callData, err := ExecTransactionData(
		to,
		value,
		txData,
		operation,
		safeTxGas,
		baseGas,
		gasPrice,
		gasToken,
		refundReceiver,
		signatures,
	)
	if err != nil {
		t.Fatalf("Failed to generate execTransaction call data: %v", err)
	}

	// Check function selector (first 4 bytes)
	parsedABI, err := abi.JSON(strings.NewReader(ABI_SafeWallet))
	if err != nil {
		t.Fatalf("Failed to parse ABI: %v", err)
	}
	expectedSelector := parsedABI.Methods["execTransaction"].ID

	if !bytes.Equal(expectedSelector, callData[:4]) {
		t.Errorf("Wrong function selector. Expected: %x, Got: %x", expectedSelector, callData[:4])
	}

	// Verify data length is greater than just the selector (should include all parameters)
	if len(callData) <= 4 {
		t.Errorf("Data length too short. Got: %d", len(callData))
	}
}

func TestParseExecTransactionResult(t *testing.T) {
	// Create a mock result for success case
	// This represents a boolean true encoded according to the ABI
	// 0x0000000000000000000000000000000000000000000000000000000000000001
	mockSuccessResult, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")

	// Parse the success result
	success, err := ParseExecTransactionResult(mockSuccessResult)
	if err != nil {
		t.Fatalf("Failed to parse execTransaction result: %v", err)
	}

	// Check that the parsed result is correct (should be true)
	if !success {
		t.Errorf("Wrong success value. Expected: true, Got: false")
	}

	// Create a mock result for failure case
	// This represents a boolean false encoded according to the ABI
	// 0x0000000000000000000000000000000000000000000000000000000000000000
	mockFailureResult, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")

	// Parse the failure result
	success, err = ParseExecTransactionResult(mockFailureResult)
	if err != nil {
		t.Fatalf("Failed to parse execTransaction result: %v", err)
	}

	// Check that the parsed result is correct (should be false)
	if success {
		t.Errorf("Wrong success value. Expected: false, Got: true")
	}
}
