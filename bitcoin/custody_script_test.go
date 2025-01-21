package bitcoinlib

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/axengine/utils"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/yalaorg/golibs/cubist"
)

func createScript() *CustodyScript {
	sk1 := "cMv5A7pRrcFsy53DV2xrQmM4dSaiSGk5PhjPadQQ2hoNo633gdXN"
	sk2 := "cRexmaiBZwFdswYfy9itFB6qLcM4ZVMKDUrk6FxP8WFghmyj2T2K"
	sk3 := "cScfkGjbzzoeewVWmU2hYPUHeVGJRDdFt7WhmrVVGkxpmPP8BHWe"
	skAdmin := "cMkopUXKWsEzAjfa1zApksGRwjVpJRB3831qM9W4gKZsLwjHXA9x"
	locktime := time.Date(2024, 12, 21, 9, 0, 0, 0, time.Local).Unix()

	wif1, err := btcutil.DecodeWIF(sk1)
	if err != nil {
		panic(err)
	}
	pk1 := wif1.PrivKey.PubKey().SerializeCompressed()

	wif2, err := btcutil.DecodeWIF(sk2)
	if err != nil {
		panic(err)
	}
	pk2 := wif2.PrivKey.PubKey().SerializeCompressed()

	wif3, err := btcutil.DecodeWIF(sk3)
	if err != nil {
		panic(err)
	}
	pk3 := wif3.PrivKey.PubKey().SerializeCompressed()

	wifAdmin, err := btcutil.DecodeWIF(skAdmin)
	if err != nil {
		panic(err)
	}
	pkAdmin := wifAdmin.PrivKey.PubKey().SerializeCompressed()

	script, _ := NewCustodyScript([]string{hex.EncodeToString(pk1), hex.EncodeToString(pk2), hex.EncodeToString(pk3)},
		hex.EncodeToString(pkAdmin), 2, 3, uint32(locktime), &chaincfg.TestNet3Params)
	return script
}

func TestCreateScript(t *testing.T) {
	script := createScript()
	t.Log(script.Address())

	t.Log(hex.EncodeToString(script.pk))
	t.Log(script.PubKey())
}

func TestWIthdraw(t *testing.T) {
	script := createScript()
	fmt.Println(script.DisasmString())
	sk1 := "cMv5A7pRrcFsy53DV2xrQmM4dSaiSGk5PhjPadQQ2hoNo633gdXN"
	sk2 := "cRexmaiBZwFdswYfy9itFB6qLcM4ZVMKDUrk6FxP8WFghmyj2T2K"
	sk3 := "cScfkGjbzzoeewVWmU2hYPUHeVGJRDdFt7WhmrVVGkxpmPP8BHWe"
	skAdmin := "cMkopUXKWsEzAjfa1zApksGRwjVpJRB3831qM9W4gKZsLwjHXA9x"

	wif1, err := btcutil.DecodeWIF(sk1)
	if err != nil {
		panic(err)
	}

	wif2, err := btcutil.DecodeWIF(sk2)
	if err != nil {
		panic(err)
	}

	wif3, err := btcutil.DecodeWIF(sk3)
	if err != nil {
		panic(err)
	}

	wifAdmin, err := btcutil.DecodeWIF(skAdmin)
	if err != nil {
		panic(err)
	}

	// Construct transaction
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	var fetcherMap = make(map[wire.OutPoint]*wire.TxOut)

	// input
	utxoHash, _ := chainhash.NewHashFromStr("1091fa0c10c3c979bb4faf3fceeb86930b95d4ff4ded47c5d9cd10c2cf2506e6")
	outPoint := wire.NewOutPoint(utxoHash, 0)
	output := wire.NewTxOut(100000, script.pk)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	txIn.Sequence = 0xfffffffd // Note: nSequence MUST be <= 0xfffffffe otherwise OP_CHECKLOCKTIMEVERIFY will fail.
	redeemTx.AddTxIn(txIn)

	fetcherMap[*outPoint] = output

	// adding the output to tx
	decodedAddr, err := btcutil.DecodeAddress("tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp", &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		t.Fatal(err)
	}

	// adding the destination address and the amount to the transaction
	redeemTxOut := wire.NewTxOut(90000, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	redeemTx.LockTime = uint32(script.locktime)

	// sign
	fetcher := txscript.NewMultiPrevOutFetcher(fetcherMap)
	txSigHashes := txscript.NewTxSigHashes(redeemTx, fetcher)
	for index, v := range redeemTx.TxIn {
		x := fetcherMap[v.PreviousOutPoint]
		sig1, err := txscript.RawTxInWitnessSignature(redeemTx, txSigHashes, index, x.Value, script.redeemScript, txscript.SigHashAll, wif1.PrivKey)
		if err != nil {
			t.Fatal(err)
		}
		_ = sig1
		sig2, err := txscript.RawTxInWitnessSignature(redeemTx, txSigHashes, index, x.Value, script.redeemScript, txscript.SigHashAll, wif2.PrivKey)
		if err != nil {
			t.Fatal(err)
		}
		_ = sig2

		sig3, err := txscript.RawTxInWitnessSignature(redeemTx, txSigHashes, index, x.Value, script.redeemScript, txscript.SigHashAll, wif3.PrivKey)
		if err != nil {
			t.Fatal(err)
		}
		_ = sig3

		sigAdmin, err := txscript.RawTxInWitnessSignature(redeemTx, txSigHashes, index, x.Value, script.redeemScript, txscript.SigHashAll, wifAdmin.PrivKey)
		if err != nil {
			fmt.Println("got error in constructing sigAdmin")
			t.Fatal(err)
		}
		_ = sigAdmin

		// CASE IF
		redeemTx.TxIn[index].Witness = wire.TxWitness{
			sigAdmin,
			{0x01},
			script.redeemScript,
		}

		// CASE ELSE
		//redeemTx.TxIn[index].Witness = wire.TxWitness{
		//	sigAdmin,
		//	{},
		//	sig1,
		//	sig2,
		//	{}, // false
		//	redeemScript,
		//}
	}

	var signedTx bytes.Buffer
	if err := redeemTx.Serialize(&signedTx); err != nil {
		t.Fatal(err)
	}
	fmt.Println("TxHash:", redeemTx.TxHash().String())

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())
	fmt.Println("signedTx:", hexSignedTx)

	var multi_expporer_api = []string{"https://mempool.space:443/testnet/api/"}
	cli := NewSmartClient(multi_expporer_api)
	txid, err := cli.PostTransaction(context.Background(), hexSignedTx)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("txid:", txid)
}

func TestClaimFromScriptx(t *testing.T) {
	pk1 := "0x041813408b94aad6745f5710e887f3c705f9cce13e3c971f5e3347cfd06bbf64fb1594c4c25392a36284f26ec89e4bc490d3815562043de013795012a782c12125"     // tb1qnv34n2wcmd94leqqs3ycrwh93pa44adhexkj9l
	pk2 := "0x04daae4253f0302725c22c9f124e6c0d182c5b4275047db181a567ff955aaebc19789155da6cd830ebc64e7ce13240c921b773681794f60e843d91d21b531584fb"     // tb1qvx32avcf6f6ke8s9h23pg7fr4d9798vs3xxrqr
	pk3 := "0x040e1c9e63a03a4274998089d6491cd5ed789a7e09efbe7994a040228646e0e45d64def988296662eb5480ecab78dda694eda605bb9a56503533f02df3f5441202"     // tb1qpgw555fq567k098wtk6346ejt4q0jxhtxeyepc
	pkAdmin := "0x043299eade0becea53c3e3943ccfbd647a103cf7654d8d7b79a11a6efeb81d8a31c396c7fa787f359608fe5b18debf4e8f242735f6049e880dd83480b403309a2a" // tb1qm39c3gupc2rfrrr6g8y8ely0ad26992lypcz5l
	locktime := time.Date(2024, 12, 21, 9, 0, 0, 0, time.Local).Unix()

	script, err := NewCustodyScript([]string{pk1, pk2, pk3}, pkAdmin, 2, 3, uint32(locktime), &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(script.DisasmString())
	t.Log(script.Address())

	cli := NewSmartClient([]string{"https://mempool.space:443/testnet/api/"})
	utxos, err := cli.GetAddressUtxos(context.Background(), script.Address())
	if err != nil {
		t.Fatal(err)
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)
	redeemTx.LockTime = uint32(script.locktime)

	bitcointx := cubist.BitcoinTx{
		Version:  wire.TxVersion,
		Locktime: script.locktime,
		Input:    nil,
		Output:   nil,
	}

	// outout
	{
		decodedAddr, err := btcutil.DecodeAddress("tb1qwszgf7yzj2ffqd303n6xmrpypdh4ax6n6sdmdnyf0hqrjdawl4mq79tg29",
			Network("testnet3"))
		if err != nil {
			t.Fatal(err)
		}
		destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			t.Fatal(err)
		}
		// adding the destination address and the amount to the transaction
		redeemTxOut := wire.NewTxOut(96000, destinationAddrByte)
		redeemTx.AddTxOut(redeemTxOut)

		bitcointx.Output = append(bitcointx.Output, cubist.Output{
			ScriptPubkey: hex.EncodeToString(redeemTxOut.PkScript),
			Value:        redeemTxOut.Value,
		})
	}

	// input
	for _, utxo := range utxos {
		chainHash, _ := chainhash.NewHashFromStr(utxo.Txid)
		prevOut := wire.NewOutPoint(chainHash, utxo.Vout)
		in := wire.NewTxIn(prevOut, nil, nil)
		in.Sequence = 0xfffffffd
		redeemTx.AddTxIn(in)

		bitcointx.Input = append(bitcointx.Input, cubist.Input{
			PreviousOutput: prevOut.String(),
			ScriptSig:      "",
			Sequence:       in.Sequence,
			Witness:        []string{},
		})
	}

	ro := cubist.SegwitSignRo{
		SigKind: cubist.SigKind{Segwit: cubist.Segwit{
			InputIndex:  0,
			ScriptCode:  hexutil.Encode(script.redeemScript),
			SighashType: "All",
			Value:       100000,
		}},
		Tx: bitcointx,
	}

	utils.JsonPrettyToStdout(ro)

	cu := cubist.New(true)
	sig1, err := cu.SegwitSign(context.Background(), "tb1qnv34n2wcmd94leqqs3ycrwh93pa44adhexkj9l", &ro, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = sig1
	sig2, err := cu.SegwitSign(context.Background(), "tb1qvx32avcf6f6ke8s9h23pg7fr4d9798vs3xxrqr", &ro, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = sig2

	sigAdmin, err := cu.SegwitSign(context.Background(), "tb1qm39c3gupc2rfrrr6g8y8ely0ad26992lypcz5l", &ro, nil)
	if err != nil {
		t.Fatal(err)
	}

	hexdecode := func(str string) []byte {
		str = strings.TrimPrefix(str, "0x")
		b, _ := hex.DecodeString(str)
		r := &secp256k1.ModNScalar{}
		r.SetByteSlice(b[:32])
		s := &secp256k1.ModNScalar{}
		s.SetByteSlice(b[32:64])

		sig := ecdsa.NewSignature(r, s)

		sigb := append(sig.Serialize(), byte(0x01))
		fmt.Println(len(sigb), hex.EncodeToString(sigb))
		return sigb
	}

	redeemTx.TxIn[0].Witness = wire.TxWitness{
		hexdecode(sigAdmin),
		{0x01},
		script.redeemScript,
	}

	// CASE ELSE
	redeemTx.TxIn[0].Witness = wire.TxWitness{
		hexdecode(sigAdmin),
		{},
		hexdecode(sig1),
		hexdecode(sig2),
		{}, // false
		script.redeemScript,
	}

	txid := redeemTx.TxHash().String()
	t.Log("txid=", txid)
	var signedTxBuf bytes.Buffer
	if err := redeemTx.Serialize(&signedTxBuf); err != nil {
		t.Fatal(err)
	}
	signedTx := hex.EncodeToString(signedTxBuf.Bytes())

	if _, err := cli.PostTransaction(context.Background(), signedTx); err != nil {
		t.Fatal(err)
	}
}

// https://mempool.space/zh/testnet/tx/c985e52360cbb27085c7a8c7367c6c64376dcde5ce06024c2b908e835894555d
func TestTxSizeWithinLocktime(t *testing.T) {
	pk1 := "0x041813408b94aad6745f5710e887f3c705f9cce13e3c971f5e3347cfd06bbf64fb1594c4c25392a36284f26ec89e4bc490d3815562043de013795012a782c12125"     // tb1qnv34n2wcmd94leqqs3ycrwh93pa44adhexkj9l
	pk2 := "0x04daae4253f0302725c22c9f124e6c0d182c5b4275047db181a567ff955aaebc19789155da6cd830ebc64e7ce13240c921b773681794f60e843d91d21b531584fb"     // tb1qvx32avcf6f6ke8s9h23pg7fr4d9798vs3xxrqr
	pk3 := "0x040e1c9e63a03a4274998089d6491cd5ed789a7e09efbe7994a040228646e0e45d64def988296662eb5480ecab78dda694eda605bb9a56503533f02df3f5441202"     // tb1qpgw555fq567k098wtk6346ejt4q0jxhtxeyepc
	pkAdmin := "0x043299eade0becea53c3e3943ccfbd647a103cf7654d8d7b79a11a6efeb81d8a31c396c7fa787f359608fe5b18debf4e8f242735f6049e880dd83480b403309a2a" // tb1qm39c3gupc2rfrrr6g8y8ely0ad26992lypcz5l
	locktime := time.Date(2024, 12, 21, 9, 0, 0, 0, time.Local).Unix()

	script, err := NewCustodyScript([]string{pk1, pk2, pk3}, pkAdmin, 2, 3, uint32(locktime), &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(script.DisasmString())
	t.Log(script.Address())

	t.Log(script.TxSizeWithinLocktime(1, 2))
}

// https://mempool.space/zh/testnet/tx/c67b7ba14cd1f64f3e7336e19548c0ac416be6d1b20a85d19afcc53a51061aee
func TestTxSizeOutsideLocktime(t *testing.T) {
	pk1 := "0x041813408b94aad6745f5710e887f3c705f9cce13e3c971f5e3347cfd06bbf64fb1594c4c25392a36284f26ec89e4bc490d3815562043de013795012a782c12125"     // tb1qnv34n2wcmd94leqqs3ycrwh93pa44adhexkj9l
	pk2 := "0x04daae4253f0302725c22c9f124e6c0d182c5b4275047db181a567ff955aaebc19789155da6cd830ebc64e7ce13240c921b773681794f60e843d91d21b531584fb"     // tb1qvx32avcf6f6ke8s9h23pg7fr4d9798vs3xxrqr
	pk3 := "0x040e1c9e63a03a4274998089d6491cd5ed789a7e09efbe7994a040228646e0e45d64def988296662eb5480ecab78dda694eda605bb9a56503533f02df3f5441202"     // tb1qpgw555fq567k098wtk6346ejt4q0jxhtxeyepc
	pkAdmin := "0x043299eade0becea53c3e3943ccfbd647a103cf7654d8d7b79a11a6efeb81d8a31c396c7fa787f359608fe5b18debf4e8f242735f6049e880dd83480b403309a2a" // tb1qm39c3gupc2rfrrr6g8y8ely0ad26992lypcz5l
	locktime := time.Date(2024, 12, 21, 9, 0, 0, 0, time.Local).Unix()

	script, err := NewCustodyScript([]string{pk1, pk2, pk3}, pkAdmin, 2, 3, uint32(locktime), &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(script.DisasmString())
	t.Log(script.Address())

	t.Log(script.TxSizeOutsideLocktime(1, 1))
}
