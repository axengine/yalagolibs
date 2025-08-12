package safewallet

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/yalaorg/golibs/cubist"
	"golang.org/x/crypto/sha3"
)

func CalcSafeAddress(
	safe,
	safeL2,
	compatibilityFallbackHandler,
	safeToL2Setup,
	safeProxyFactory common.Address,
	creationCode string,
	accounts []common.Address,
	threshold *big.Int,
	nonce *big.Int,
) (common.Address, error) {
	setupToL2ABIJson := `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"singleton","type":"address"}],"name":"ChangedMasterCopy","type":"event"},{"inputs":[{"internalType":"address","name":"l2Singleton","type":"address"}],"name":"setupToL2","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	setupToL2ABI, _ := abi.JSON(strings.NewReader(setupToL2ABIJson))
	data, _ := setupToL2ABI.Pack("setupToL2", safeL2)

	safeL2ABIJson := `[{"inputs":[{"internalType":"address[]","name":"_owners","type":"address[]"},{"internalType":"uint256","name":"_threshold","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"bytes","name":"data","type":"bytes"},{"internalType":"address","name":"fallbackHandler","type":"address"},{"internalType":"address","name":"paymentToken","type":"address"},{"internalType":"uint256","name":"payment","type":"uint256"},{"internalType":"address payable","name":"paymentReceiver","type":"address"}],"name":"setup","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	safeL2ABI, _ := abi.JSON(strings.NewReader(safeL2ABIJson))

	initializer, err := safeL2ABI.Pack("setup",
		accounts,
		threshold,
		safeToL2Setup,
		data,
		compatibilityFallbackHandler,
		common.Address{},
		big.NewInt(0),
		common.HexToAddress("0x5afe7A11E7000000000000000000000000000000"),
	)
	if err != nil {
		return common.Address{}, err
	}

	safeProxyCreationCode, _ := hexutil.Decode(creationCode)
	_singleton := safe
	deployCode := encodePacked(
		common.LeftPadBytes(safeProxyCreationCode, 32),
		common.LeftPadBytes(_singleton.Bytes(), 32),
	)
	deployCodeHash := crypto.Keccak256Hash(deployCode)

	salt := keccak256(encodePacked(keccak256(initializer), nonce))
	addressCode := encodePacked(
		[]byte{0xff},
		safeProxyFactory.Bytes(),
		salt,
		deployCodeHash.Bytes(),
	)

	addrHash := crypto.Keccak256Hash(addressCode)
	return common.BytesToAddress(addrHash[12:]), nil
}

func encodePacked(args ...interface{}) []byte {
	bz := make([]byte, 0)
	for _, arg := range args {
		switch val := arg.(type) {
		case *big.Int:
			bz = append(bz, common.LeftPadBytes(val.Bytes(), 32)...)
		case bool:
			if val {
				bz = append(bz, []byte{0x0, 0x1}...)
			}
		case common.Hash:
			bz = append(bz, val[:]...)
		case []byte:
			bz = append(bz, val...)
		case common.Address:
			bz = append(bz, val[:]...)
		case uint8, uint16, uint32:
			buf := new(bytes.Buffer)
			_ = binary.Write(buf, binary.BigEndian, val)
			bz = append(bz, buf.Bytes()...)
		default:
			panic(fmt.Sprintf("unsupport type %T", arg))
		}
	}
	return bz
}

// Function to calculate keccak256 hash
func keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256() // Ethereum uses Keccak256
	hash.Write(data)
	return hash.Sum(nil)
}

var SafeTypedData = cubist.Types{
	"EIP712Domain": {
		{
			Name: "chainId",
			Type: "uint256",
		},
		{
			Name: "verifyingContract",
			Type: "address",
		},
	},
	"SafeTx": {
		{
			Name: "to",
			Type: "address",
		},
		{
			Name: "value",
			Type: "uint256",
		},
		{
			Name: "data",
			Type: "bytes",
		},
		{
			Name: "operation",
			Type: "uint8",
		},
		{
			Name: "safeTxGas",
			Type: "uint256",
		},
		{
			Name: "baseGas",
			Type: "uint256",
		},
		{
			Name: "gasPrice",
			Type: "uint256",
		},
		{
			Name: "gasToken",
			Type: "address",
		},
		{
			Name: "refundReceiver",
			Type: "address",
		},
		{
			Name: "nonce",
			Type: "uint256",
		},
	},
}

// GetNonceData generates contract call data for nonce function
// This function is used to query the current nonce of a Safe wallet
func GetNonceData() ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ABI_SafeWallet))
	if err != nil {
		return nil, err
	}

	data, err := parsedABI.Pack("nonce")
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ParseNonceResult parses the result of a nonce call and returns the nonce as a big.Int
// result: the raw result bytes from the contract call
func ParseNonceResult(result []byte) (*big.Int, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ABI_SafeWallet))
	if err != nil {
		return nil, err
	}

	var nonce *big.Int
	err = parsedABI.UnpackIntoInterface(&nonce, "nonce", result)
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

// ExecTransactionData generates contract call data for execTransaction function
// This function is used to execute a transaction through a Safe wallet
// to: destination address of the transaction
// value: ether value of the transaction
// txData: data payload of the transaction
// operation: operation type (0 = call, 1 = delegatecall)
// safeTxGas: gas that should be used for the safe transaction
// baseGas: gas costs for data used to trigger the safe transaction
// gasPrice: gas price that should be used for the payment calculation
// gasToken: token address (or 0 if ETH) that is used for the payment
// refundReceiver: address of receiver of gas payment (or 0 if tx.origin)
// signatures: packed signature data ({bytes32 r}{bytes32 s}{uint8 v})
func ExecTransactionData(
	to common.Address,
	value *big.Int,
	txData []byte,
	operation uint8,
	safeTxGas *big.Int,
	baseGas *big.Int,
	gasPrice *big.Int,
	gasToken common.Address,
	refundReceiver common.Address,
	signatures []byte,
) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ABI_SafeWallet))
	if err != nil {
		return nil, err
	}

	packedData, err := parsedABI.Pack(
		"execTransaction",
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
		return nil, err
	}

	return packedData, nil
}

// ParseExecTransactionResult parses the result of an execTransaction call and returns the success status
// result: the raw result bytes from the contract call
func ParseExecTransactionResult(result []byte) (bool, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ABI_SafeWallet))
	if err != nil {
		return false, err
	}

	var success bool
	err = parsedABI.UnpackIntoInterface(&success, "execTransaction", result)
	if err != nil {
		return false, err
	}

	return success, nil
}

// SafeTransactionData represents the data of a Safe transaction
type SafeTransactionData struct {
	To             common.Address
	Value          *big.Int
	Data           []byte
	Operation      uint8
	SafeTxGas      *big.Int
	BaseGas        *big.Int
	GasPrice       *big.Int
	GasToken       common.Address
	RefundReceiver common.Address
	Nonce          *big.Int
}

// EIP712 domain and type constants for Safe transactions
const (
	EIP712_SAFE_TX_TYPE = "SafeTx(address to,uint256 value,bytes data,uint8 operation,uint256 safeTxGas,uint256 baseGas,uint256 gasPrice,address gasToken,address refundReceiver,uint256 nonce)"
)

// GetSafeTransactionHash calculates the EIP712 hash of a Safe transaction
// safeAddress: the address of the Safe contract
// chainId: the chain ID of the network
// txData: the Safe transaction data
func GetSafeTransactionHash(safeAddress common.Address, chainId *big.Int, txData SafeTransactionData) (common.Hash, error) {
	// Create the domain separator hash
	domainSeparator := crypto.Keccak256Hash(
		[]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
		crypto.Keccak256Hash([]byte("Gnosis Safe")).Bytes(),
		crypto.Keccak256Hash([]byte("1.3.0")).Bytes(),
		common.LeftPadBytes(chainId.Bytes(), 32),
		common.LeftPadBytes(safeAddress.Bytes(), 32),
	)

	// Create the SafeTx type hash
	safeTxTypeHash := crypto.Keccak256Hash([]byte(EIP712_SAFE_TX_TYPE))

	// Pack the transaction data
	txDataHash := crypto.Keccak256Hash(
		safeTxTypeHash.Bytes(),
		common.LeftPadBytes(txData.To.Bytes(), 32),
		common.LeftPadBytes(txData.Value.Bytes(), 32),
		crypto.Keccak256Hash(txData.Data).Bytes(),
		common.LeftPadBytes([]byte{txData.Operation}, 32),
		common.LeftPadBytes(txData.SafeTxGas.Bytes(), 32),
		common.LeftPadBytes(txData.BaseGas.Bytes(), 32),
		common.LeftPadBytes(txData.GasPrice.Bytes(), 32),
		common.LeftPadBytes(txData.GasToken.Bytes(), 32),
		common.LeftPadBytes(txData.RefundReceiver.Bytes(), 32),
		common.LeftPadBytes(txData.Nonce.Bytes(), 32),
	)

	// Combine domain separator and transaction hash according to EIP-712
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, domainSeparator.Bytes(), txDataHash.Bytes()), nil
}

// SignSafeTransactionHash signs a Safe transaction hash using a private key
// txHash: the EIP712 hash of the Safe transaction
// privateKey: the private key to sign with
// Returns the signature in the format required by the Safe contract (65 bytes: r, s, v)
func SignSafeTransactionHash(txHash common.Hash, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	// Sign the hash with the private key
	signature, err := crypto.Sign(txHash.Bytes(), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction hash: %v", err)
	}

	// Convert the signature to the format expected by the Safe contract
	// Safe expects signatures in the format: {bytes32 r}{bytes32 s}{uint8 v}
	// The 'v' value needs to be adjusted: v = 27 + v'
	signature[64] += 27 // Transform V from 0/1 to 27/28

	return signature, nil
}

// SignSafeTransaction signs a Safe transaction using a private key
// safeAddress: the address of the Safe contract
// chainId: the chain ID of the network
// txData: the Safe transaction data
// privateKey: the private key to sign with
// Returns the signature in the format required by the Safe contract (65 bytes: r, s, v)
func SignSafeTransaction(safeAddress common.Address, chainId *big.Int, txData SafeTransactionData, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	// Calculate the transaction hash
	txHash, err := GetSafeTransactionHash(safeAddress, chainId, txData)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction hash: %v", err)
	}

	// Sign the transaction hash
	return SignSafeTransactionHash(txHash, privateKey)
}

// FormatSignatureBytes formats multiple signatures into the format expected by the Safe contract
// Each signature should be 65 bytes (r, s, v)
// The signatures should be sorted by signer address (ascending order)
func FormatSignatureBytes(signatures ...[]byte) []byte {
	result := []byte{}
	for _, sig := range signatures {
		if len(sig) != 65 {
			continue // Skip invalid signatures
		}
		result = append(result, sig...)
	}
	return result
}

// GetSignerAddress recovers the signer address from a signature and transaction hash
// txHash: the EIP712 hash of the Safe transaction
// signature: the 65-byte signature (r, s, v)
// Returns the address of the signer
func GetSignerAddress(txHash common.Hash, signature []byte) (common.Address, error) {
	if len(signature) != 65 {
		return common.Address{}, fmt.Errorf("invalid signature length: got %d, want 65", len(signature))
	}

	// Make a copy of the signature to avoid modifying the original
	sigCopy := make([]byte, 65)
	copy(sigCopy, signature)

	// Adjust v value back to 0/1 for recovery
	if sigCopy[64] >= 27 {
		sigCopy[64] -= 27
	}

	// Recover the public key
	pubKey, err := crypto.SigToPub(txHash.Bytes(), sigCopy)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to recover public key: %v", err)
	}

	// Derive the address from the public key
	return crypto.PubkeyToAddress(*pubKey), nil
}

func GetNonce(ctx context.Context, client *ethclient.Client, safeAddress common.Address, block *big.Int) (*big.Int, error) {
	calldata, err := GetNonceData()
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce data: %v", err)
	}

	respData, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &safeAddress,
		Data: calldata,
	}, block)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}
	nonce, err := ParseNonceResult(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nonce result: %v", err)
	}
	return nonce, nil
}
