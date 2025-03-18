package bitcoinlib

import (
	"errors"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/bech32"
)

func EncodeAddress(address string, network string) (string, error) {
	net := Network(network)
	decodedAddr, err := btcutil.DecodeAddress(address, net)
	if err != nil {
		return "", err
	}
	if !decodedAddr.IsForNet(net) {
		return "", errors.New("address is not For Network")
	}
	return decodedAddr.EncodeAddress(), nil
}

// IsTaprootAddress Determine whether the address is a Taproot address and return the network type
func IsTaprootAddress(address string) (bool, string) {
	// Check if the address starts with "bc1p" or "tb1p"
	if !strings.HasPrefix(address, "bc1p") && !strings.HasPrefix(address, "tb1p") {
		return false, ""
	}

	// Decode Bech32m address
	hrp, _, err := bech32.Decode(address)
	if err != nil {
		return false, ""
	}

	// Determine network type based on HRP
	switch hrp {
	case "bc":
		return true, "mainnet"
	case "tb":
		return true, "testnet"
	default:
		return false, ""
	}
}

// IsValidTaprootAddressForNetwork checks if the address is a valid Taproot address for the specified network
func IsValidTaprootAddressForNetwork(address string, network string) bool {
	isTaproot, addrNetwork := IsTaprootAddress(address)
	if !isTaproot {
		return false
	}
	return addrNetwork == network
}
