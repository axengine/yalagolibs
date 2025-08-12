package safewallet

import (
	"math/big"
	"testing"

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
		nonce)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr.Hex())
	// if addr.Hex() != "0xB6A5246A7952E50956a5C955095a8C78473EFAf4" {
	// 	t.Fatal("not equal")
	// }
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
