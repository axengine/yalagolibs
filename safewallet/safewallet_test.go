package safewallet

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
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
	if addr.Hex() != "0xB6A5246A7952E50956a5C955095a8C78473EFAf4" {
		t.Fatal("not equal")
	}
}
