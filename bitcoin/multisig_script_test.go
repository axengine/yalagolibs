package bitcoinlib

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/axengine/gogolibs/cubist"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func TestNewMultisigScript(t *testing.T) {
	_, pk1, _, _ := CreateBitcoinTaprootAccount("testnet")
	_, pk2, _, _ := CreateBitcoinTaprootAccount("testnet")
	_, pk3, _, _ := CreateBitcoinTaprootAccount("testnet")

	script, err := NewMultisigScript([]string{pk1, pk2, pk3}, 2, 3, &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(script.DisasmString())
	fmt.Println(script.PubKey())
	fmt.Println(script.Address())
	fmt.Println(script.RedeemScript())
}

// withdraw btc from P2WSH
func TestClaimFromScript(t *testing.T) {
	pk1 := "0x041813408b94aad6745f5710e887f3c705f9cce13e3c971f5e3347cfd06bbf64fb1594c4c25392a36284f26ec89e4bc490d3815562043de013795012a782c12125"
	pk2 := "0x04daae4253f0302725c22c9f124e6c0d182c5b4275047db181a567ff955aaebc19789155da6cd830ebc64e7ce13240c921b773681794f60e843d91d21b531584fb"
	pk3 := "0x040e1c9e63a03a4274998089d6491cd5ed789a7e09efbe7994a040228646e0e45d64def988296662eb5480ecab78dda694eda605bb9a56503533f02df3f5441202"
	var pks []string
	{
		bz, _ := hex.DecodeString(pk1[2:])
		pk, _ := btcec.ParsePubKey(bz)
		pks = append(pks, hex.EncodeToString(pk.SerializeCompressed()))
		bz, _ = hex.DecodeString(pk2[2:])
		pk, _ = btcec.ParsePubKey(bz)
		pks = append(pks, hex.EncodeToString(pk.SerializeCompressed()))
		bz, _ = hex.DecodeString(pk3[2:])
		pk, _ = btcec.ParsePubKey(bz)
		pks = append(pks, hex.EncodeToString(pk.SerializeCompressed()))
	}
	script, err := NewMultisigScript(pks, 2, 3, &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(script.Address())
	fmt.Println(hex.EncodeToString(script.RedeemScript()))

	cli := NewSmartClient([]string{"https://mempool.space:443/testnet/api/", "https://blockstream.info/testnet/api/"})
	utxos, err := cli.GetAddressUtxos(context.Background(), script.Address())
	if err != nil {
		t.Fatal(err)
	}

	// build PSBT
	var (
		inputs     []*wire.OutPoint
		nSequences []uint32
		outputs    []*wire.TxOut
	)
	// outout
	{
		decodedAddr, err := btcutil.DecodeAddress("tb1p83qvzymlvg0vafghxe6jdgr3cchwsscnes74d39778u0llqs47mqj7ztey",
			Network("testnet3"))
		if err != nil {
			t.Fatal(err)
		}
		destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			t.Fatal(err)
		}
		// adding the destination address and the amount to the transaction
		redeemTxOut := wire.NewTxOut(113400000, destinationAddrByte)
		outputs = append(outputs, redeemTxOut)
	}

	// input
	for _, utxo := range utxos {
		chainHash, _ := chainhash.NewHashFromStr(utxo.Txid)
		prevOut := wire.NewOutPoint(chainHash, utxo.Vout)
		inputs = append(inputs, prevOut)
		nSequences = append(nSequences, 0xfffffffd)
	}

	packet, err := psbt.New(inputs, outputs, 2, 0, nSequences)
	if err != nil {
		t.Fatal(err)
	}

	// add witness
	updater, err := psbt.NewUpdater(packet)
	if err != nil {
		t.Fatal(err)
	}
	for index, v := range utxos {
		tx, err := cli.GetTransaction(context.Background(), v.Txid)
		if err != nil {
			t.Fatal(err)
		}
		if tx == nil {
			t.Fatal(err)
		}
		if int(v.Vout) >= len(tx.Vout) {
			t.Fatal(err)
		}
		out := tx.Vout[v.Vout]
		pk, err := hex.DecodeString(out.ScriptPubkey) // p2wsh pk
		if err != nil {
			t.Fatal(err)
		}
		if err := updater.AddInWitnessUtxo(wire.NewTxOut(v.Value, pk), index); err != nil {
			t.Fatal(err)
		}
		redeemScript, err := hex.DecodeString("5221031813408b94aad6745f5710e887f3c705f9cce13e3c971f5e3347cfd06bbf64fb2103daae4253f0302725c22c9f124e6c0d182c5b4275047db181a567ff955aaebc1921020e1c9e63a03a4274998089d6491cd5ed789a7e09efbe7994a040228646e0e45d53ae") // redemption script
		if err != nil {
			t.Fatal(err)
		}
		if err := updater.AddInWitnessScript(redeemScript, index); err != nil {
			t.Fatal(err)
		}
	}

	var serializedTx bytes.Buffer
	if err := packet.Serialize(&serializedTx); err != nil {
		t.Fatal(err)
	}
	unsignedPsbt := hex.EncodeToString(serializedTx.Bytes())
	cu := cubist.New(true, "")
	unsignedPsbt, err = cu.PsbtSign(context.Background(), "tb1qnv34n2wcmd94leqqs3ycrwh93pa44adhexkj9l", unsignedPsbt, nil)
	if err != nil {
		t.Fatal(err)
	}
	signedPsbtHex, err := cu.PsbtSign(context.Background(), "tb1qvx32avcf6f6ke8s9h23pg7fr4d9798vs3xxrqr", unsignedPsbt, nil)
	if err != nil {
		t.Fatal(err)
	}

	signedPsbtBz, err := hex.DecodeString(signedPsbtHex)
	if err != nil {
		t.Fatal(err)
	}
	packet, err = psbt.NewFromRawBytes(bytes.NewReader(signedPsbtBz), false)
	if err != nil {
		t.Fatal(err)
	}
	if err = psbt.MaybeFinalizeAll(packet); err != nil {
		t.Fatal(err)
	}
	msgTx, err := psbt.Extract(packet)
	if err != nil {
		t.Fatal(err)
	}

	txid := msgTx.TxHash().String()
	t.Log("txid=", txid)
	var signedTxBuf bytes.Buffer
	if err := msgTx.Serialize(&signedTxBuf); err != nil {
		t.Fatal(err)
	}
	signedTx := hex.EncodeToString(signedTxBuf.Bytes())

	if _, err := cli.PostTransaction(context.Background(), signedTx); err != nil {
		t.Fatal(err)
	}
}
