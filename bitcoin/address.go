package bitcoinlib

import (
	"errors"
	"github.com/btcsuite/btcd/btcutil"
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
