package bitcoinlib

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
)

func TestSign(t *testing.T) {
	wif, err := btcutil.DecodeWIF("cVgvJgH6mcjh9qcoMHvk5AubE7G3xEtBoHiFDnjMh7GArhEiDPEB")
	if err != nil {
		t.Fatal(err)
	}
	msg := "Hello,YALA\nd16cdc32036de0bac164f7b2c96bc6554857a800a0f96f4d42a8fba3f02288d4"
	signature := SignMessage(wif.PrivKey, []byte(msg))
	t.Log(hex.EncodeToString(signature))

	pk, err := VerifyMessage([]byte(msg), signature, "tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp", &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(pk.SerializeCompressed()))
}
