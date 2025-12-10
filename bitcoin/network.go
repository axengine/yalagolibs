package bitcoinlib

import "github.com/btcsuite/btcd/chaincfg"

// Network returns the bitcoin network based on the string, with testnet3 by default
func Network(network string) *chaincfg.Params {
	var param = &chaincfg.TestNet3Params
	if network == "mainnet" {
		param = &chaincfg.MainNetParams
	} else if network == "regression" {
		param = &chaincfg.RegressionNetParams
	}
	return param
}
