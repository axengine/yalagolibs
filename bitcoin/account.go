package bitcoinlib

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
)

// CreateBitcoinTaprootAccount create bitcoin taproot account
func CreateBitcoinTaprootAccount(network string) (string, string, string, error) {
	netParams := Network(network)

	sk, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", "", err
	}

	pk := sk.PubKey()

	pubKeyAddr, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(
		txscript.ComputeTaprootKeyNoScript(
			pk)), netParams)
	if err != nil {
		return "", "", "", err
	}
	wif, err := btcutil.NewWIF(sk, netParams, true)
	if err != nil {
		return "", "", "", err
	}
	return wif.String(), hex.EncodeToString(pk.SerializeCompressed()), pubKeyAddr.EncodeAddress(), nil
}
