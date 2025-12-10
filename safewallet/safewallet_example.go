package safewallet

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// ExampleSignSafeTransaction demonstrates how to sign a Safe transaction using a private key
func ExampleSignSafeTransaction() {
	// Generate a private key for demonstration (in production, you would load your private key securely)
	privateKey, err := crypto.HexToECDSA("8da4ef21b864d2cc526dbdb2a120bd2874c36c9d0a1fb7f8c63d7f7a8b41de8f")
	if err != nil {
		fmt.Printf("Failed to parse private key: %v\n", err)
		return
	}

	// Get the owner address from the private key
	ownerAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("Owner address: %s\n", ownerAddress.Hex())

	// Safe wallet address and chain ID
	safeAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	chainId := big.NewInt(1) // Ethereum mainnet

	// Create a transaction to sign
	txData := SafeTransactionData{
		To:             common.HexToAddress("0xabcdef0123456789abcdef0123456789abcdef01"),
		Value:          big.NewInt(1000000000000000000), // 1 ETH
		Data:           []byte{},                        // Empty data for a simple ETH transfer
		Operation:      0,                               // Call operation
		SafeTxGas:      big.NewInt(21000),
		BaseGas:        big.NewInt(0),
		GasPrice:       big.NewInt(0),
		GasToken:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
		RefundReceiver: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Nonce:          big.NewInt(42),
	}

	// Get the transaction hash for demonstration
	txHash, err := GetSafeTransactionHash(safeAddress, chainId, txData)
	if err != nil {
		fmt.Printf("Failed to get transaction hash: %v\n", err)
		return
	}

	fmt.Printf("Transaction hash: %s\n", txHash.Hex())

	// Sign the transaction
	signature, err := SignSafeTransaction(safeAddress, chainId, txData, privateKey)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %v\n", err)
		return
	}

	fmt.Printf("Signature: %s\n", hexutil.Encode(signature))

	// Verify the signature by recovering the signer address
	recoveredAddress, err := GetSignerAddress(txHash, signature)
	if err != nil {
		fmt.Printf("Failed to recover signer address: %v\n", err)
		return
	}

	fmt.Printf("Recovered signer address: %s\n", recoveredAddress.Hex())

	// Format the signature for use with Safe's execTransaction
	formattedSignature := FormatSignatureBytes(signature)
	fmt.Printf("Formatted signature for execTransaction: %s\n", hexutil.Encode(formattedSignature))

	// Output:
	// Owner address: 0x5ce9454909639D2D17A3F753ce7d93fa0b9aB12E
	// Transaction hash: 0x...
	// Signature: 0x...
	// Recovered signer address: 0x5ce9454909639D2D17A3F753ce7d93fa0b9aB12E
	// Formatted signature for execTransaction: 0x...
}

// ExampleMultipleSigners demonstrates how to collect signatures from multiple owners
func ExampleMultipleSigners() {
	// Generate private keys for demonstration (in production, these would be securely stored)
	privateKey1, _ := crypto.HexToECDSA("8da4ef21b864d2cc526dbdb2a120bd2874c36c9d0a1fb7f8c63d7f7a8b41de8f")
	privateKey2, _ := crypto.HexToECDSA("4762e04d10832808a0aebdaa79c12de54afbe006bfffd228b3abcc494fe986f9")

	// Get the owner addresses
	owner1Address := crypto.PubkeyToAddress(privateKey1.PublicKey)
	owner2Address := crypto.PubkeyToAddress(privateKey2.PublicKey)

	fmt.Printf("Owner 1: %s\n", owner1Address.Hex())
	fmt.Printf("Owner 2: %s\n", owner2Address.Hex())

	// Safe wallet address and chain ID
	safeAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	chainId := big.NewInt(1) // Ethereum mainnet

	// Create a transaction to sign
	txData := SafeTransactionData{
		To:             common.HexToAddress("0xabcdef0123456789abcdef0123456789abcdef01"),
		Value:          big.NewInt(1000000000000000000), // 1 ETH
		Data:           []byte{},                        // Empty data for a simple ETH transfer
		Operation:      0,                               // Call operation
		SafeTxGas:      big.NewInt(21000),
		BaseGas:        big.NewInt(0),
		GasPrice:       big.NewInt(0),
		GasToken:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
		RefundReceiver: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Nonce:          big.NewInt(42),
	}

	// We don't need to get the transaction hash directly for this example
	// as SignSafeTransaction will calculate it internally

	// Sign the transaction with both private keys
	signature1, err := SignSafeTransaction(safeAddress, chainId, txData, privateKey1)
	if err != nil {
		fmt.Printf("Failed to sign transaction with key 1: %v\n", err)
		return
	}

	signature2, err := SignSafeTransaction(safeAddress, chainId, txData, privateKey2)
	if err != nil {
		fmt.Printf("Failed to sign transaction with key 2: %v\n", err)
		return
	}

	// Format the signatures for use with Safe's execTransaction
	// Note: Signatures must be sorted by signer address (ascending order) for Safe contracts
	formattedSignatures := FormatSignatureBytes(signature1, signature2)
	fmt.Printf("Combined signatures for execTransaction: %s\n", hexutil.Encode(formattedSignatures))

	// Output:
	// Owner 1: 0x5ce9454909639D2D17A3F753ce7d93fa0b9aB12E
	// Owner 2: 0x...
	// Combined signatures for execTransaction: 0x...
}

// ExampleExecTransactionWithSignature demonstrates how to use the signature with execTransaction
func ExampleExecTransactionWithSignature() {
	// Generate a private key for demonstration
	privateKey, _ := crypto.HexToECDSA("8da4ef21b864d2cc526dbdb2a120bd2874c36c9d0a1fb7f8c63d7f7a8b41de8f")

	// Safe wallet address and chain ID
	safeAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	chainId := big.NewInt(1) // Ethereum mainnet

	// Create a transaction to sign
	txData := SafeTransactionData{
		To:             common.HexToAddress("0xabcdef0123456789abcdef0123456789abcdef01"),
		Value:          big.NewInt(1000000000000000000), // 1 ETH
		Data:           []byte{},                        // Empty data for a simple ETH transfer
		Operation:      0,                               // Call operation
		SafeTxGas:      big.NewInt(21000),
		BaseGas:        big.NewInt(0),
		GasPrice:       big.NewInt(0),
		GasToken:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
		RefundReceiver: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Nonce:          big.NewInt(42),
	}

	// Sign the transaction
	signature, err := SignSafeTransaction(safeAddress, chainId, txData, privateKey)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %v\n", err)
		return
	}

	// Format the signature for use with Safe's execTransaction
	formattedSignature := FormatSignatureBytes(signature)

	// Generate the execTransaction call data
	callData, err := ExecTransactionData(
		txData.To,
		txData.Value,
		txData.Data,
		txData.Operation,
		txData.SafeTxGas,
		txData.BaseGas,
		txData.GasPrice,
		txData.GasToken,
		txData.RefundReceiver,
		formattedSignature,
	)
	if err != nil {
		fmt.Printf("Failed to generate execTransaction call data: %v\n", err)
		return
	}

	fmt.Printf("execTransaction call data: %s\n", hexutil.Encode(callData))

	// Output:
	// execTransaction call data: 0x...
}

// ExampleGetNonce demonstrates how to use GetNonceData and ParseNonceResult functions
func ExampleGetNonce() {
	// Generate contract call data for nonce query
	data, err := GetNonceData()
	if err != nil {
		log.Fatalf("Failed to generate nonce call data: %v", err)
	}

	// Print the generated call data in hex format
	fmt.Printf("Nonce query call data: 0x%s\n", hex.EncodeToString(data))

	// In a real application, you would send this data to the contract and parse the result
	// Example of how to use the result:
	// Assuming you've received the result bytes from the contract call
	mockResult, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000002a")
	nonce, err := ParseNonceResult(mockResult)
	if err != nil {
		log.Fatalf("Failed to parse nonce result: %v", err)
	}
	fmt.Printf("Parsed nonce: %s\n", nonce.String()) // Should print "42"
}

// ExampleUseNonce shows how to use the nonce value in a transaction
func ExampleUseNonce() {
	// In a real application, you would get the nonce from the contract
	nonce := big.NewInt(42)

	// Use the nonce for transaction preparation
	// For example, when preparing a Safe transaction:
	fmt.Printf("Using nonce %s for the next transaction\n", nonce.String())

	// The nonce would typically be used in transaction parameters
	// or included in transaction data for contract interactions
}

// ExampleExecTransaction demonstrates how to use ExecTransactionData and ParseExecTransactionResult functions
func ExampleExecTransaction() {
	// Example parameters for a Safe transaction
	to := common.HexToAddress("0x1234567890123456789012345678901234567890") // Destination address
	value := big.NewInt(1000000000000000000)                                // 1 ETH

	// Example transaction data (e.g., ERC20 transfer)
	txData, _ := hex.DecodeString("a9059cbb000000000000000000000000abcdef0123456789abcdef0123456789abcdef010000000000000000000000000000000000000000000000000de0b6b3a7640000")

	operation := uint8(0) // Call operation (0 = call, 1 = delegatecall)
	safeTxGas := big.NewInt(21000)
	baseGas := big.NewInt(10000)
	gasPrice := big.NewInt(1000000000)                                                  // 1 Gwei
	gasToken := common.HexToAddress("0x0000000000000000000000000000000000000000")       // ETH (zero address)
	refundReceiver := common.HexToAddress("0x0000000000000000000000000000000000000000") // No refund (zero address)

	// In a real application, signatures would be collected from owners
	// This is just a placeholder example signature
	signatures := []byte{}
	// Example: append multiple signatures in format {bytes32 r}{bytes32 s}{uint8 v}
	signatures = append(signatures, make([]byte, 65)...) // Placeholder for a 65-byte signature

	// Generate contract call data
	execData, err := ExecTransactionData(
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
		log.Fatalf("Failed to generate execTransaction call data: %v", err)
	}

	// Print the generated call data in hex format
	fmt.Printf("ExecTransaction call data: 0x%s\n", hex.EncodeToString(execData))

	// In a real application, you would send this data to the contract and parse the result
	// Example of how to use the result:
	// Assuming you've received the result bytes from the contract call
	mockResult, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	success, err := ParseExecTransactionResult(mockResult)
	if err != nil {
		log.Fatalf("Failed to parse execTransaction result: %v", err)
	}

	if success {
		fmt.Println("Transaction executed successfully")
	} else {
		fmt.Println("Transaction execution failed")
	}
}
