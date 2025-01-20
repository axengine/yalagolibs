package bitcoinlib

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"testing"
)

func TestSchoorPk2Taproot(t *testing.T) {
	bz, _ := hexutil.Decode("0xf84a799e9873f9a8c19cd34007bc4c2173e48be882f3a50c949210903e314a4e")
	addr, err := SchnorrPk2Taproot(bz, &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	if addr != "tb1ps2zjfxsvalaqjyge9s4cz2c5l2mxf6augu7vez40m966qxfamqjsdjwk2l" {
		t.Fatal("not equals")
	}
}
