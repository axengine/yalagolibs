package bitcoinlib

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/txscript"
)

// CreateBitcoinTaprootAccount create bitcoin taproot account
func CreateBitcoinTaprootAccount(network string) (string, string, string, error) {
	netParams := Network(network)
	// generate hd wallet seed
	seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	if err != nil {
		return "", "", "", err
	}
	masterKey, err := hdkeychain.NewMaster(seed, netParams)
	if err != nil {
		return "", "", "", err
	}

	// derive. key
	privateKey, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return "", "", "", err
	}

	privKeyBytes, err := privateKey.ECPrivKey()
	if err != nil {
		return "", "", "", err
	}

	pubKeyBytes, err := privateKey.ECPubKey()
	if err != nil {
		return "", "", "", err
	}

	pubKeyAddr, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(
		txscript.ComputeTaprootKeyNoScript(
			pubKeyBytes)), netParams)
	if err != nil {
		return "", "", "", err
	}
	wif, err := btcutil.NewWIF(privKeyBytes, netParams, true)
	if err != nil {
		return "", "", "", err
	}
	return wif.String(), hex.EncodeToString(pubKeyBytes.SerializeCompressed()), pubKeyAddr.EncodeAddress(), nil
}
