package bitcoinlib

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
)

func TestSign(t *testing.T) {
	wif, err := btcutil.DecodeWIF("--")
	if err != nil {
		t.Fatal(err)
	}
	msg := "Hello,YALA\nd16cdc32036de0bac164f7b2c96bc6554857a800a0f96f4d42a8fba3f02288d4"
	signature := SignBitcoinMessage(wif.PrivKey, []byte(msg))
	t.Log(hex.EncodeToString(signature))
}
