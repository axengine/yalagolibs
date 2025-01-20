package safewallet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
	"math/big"
	"strings"
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
