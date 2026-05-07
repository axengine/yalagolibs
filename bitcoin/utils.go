package bitcoinlib

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

func SchnorrPk2SECPPk(bz []byte) (*btcec.PublicKey, error) {
	if len(bz) != 32 {
		return nil, fmt.Errorf("invalid schnorr public key")
	}
	pk, err := schnorr.ParsePubKey(bz)
	return pk, err
}

func SchnorrPk2Taproot(bz []byte, network *chaincfg.Params) (string, error) {
	pk, err := SchnorrPk2SECPPk(bz)
	if err != nil {
		return "", err
	}

	pubKey := txscript.ComputeTaprootKeyNoScript(pk)

	witnessProg := schnorr.SerializePubKey(pubKey)
	tapAddr, err := btcutil.NewAddressTaproot(witnessProg, network)
	if err != nil {
		return "", err
	}
	return tapAddr.EncodeAddress(), nil
}
