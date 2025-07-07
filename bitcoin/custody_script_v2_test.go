package bitcoinlib

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/yalaorg/golibs/cubist"
)

func createCustodyV2Script() *CustodyScriptV2 {
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

	script, _ := NewCustodyScriptV2([]string{hex.EncodeToString(pk1), hex.EncodeToString(pk2), hex.EncodeToString(pk3)},
		hex.EncodeToString(pkAdmin), 2, 3, uint32(locktime), &chaincfg.TestNet3Params)
	return script
}

func TestNewCustodyScriptV2(t *testing.T) {
	script := createCustodyV2Script()
	t.Log(script.Address())

	t.Log(hex.EncodeToString(script.pk))
	t.Log(script.PubKey())
}

func TestWithdrawV2(t *testing.T) {
	script := createCustodyV2Script()
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
	utxoHash, _ := chainhash.NewHashFromStr("3fc3753dbd4796abcbf341ca0ab04c056fb180a098c689ea2d3369e0908de270")
	outPoint := wire.NewOutPoint(utxoHash, 0)
	output := wire.NewTxOut(99000, script.pk)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	txIn.Sequence = 0xfffffffd // Note: nSequence MUST be <= 0xfffffffe otherwise OP_CHECKLOCKTIMEVERIFY will fail.
	redeemTx.AddTxIn(txIn)

	fetcherMap[*outPoint] = output

	// adding the output to tx
	decodedAddr, err := btcutil.DecodeAddress(script.Address(), &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		t.Fatal(err)
	}

	// adding the destination address and the amount to the transaction
	redeemTxOut := wire.NewTxOut(98000, destinationAddrByte)
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

		// asm: '<sig(U)> 0 0 0'
		// redeemTx.TxIn[index].Witness = wire.TxWitness{
		// 	sigAdmin,
		// 	{}, // false
		// 	{}, // null signature
		// 	{}, // null signature
		// 	script.redeemScript,
		// }

		// asm: '<sig(U)> 0 <sig(A)> <sig(B)>'
		redeemTx.TxIn[index].Witness = wire.TxWitness{
			sigAdmin,
			{},
			sig1,
			sig2,
			script.redeemScript,
		}
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

func TestWithdrawV2PSBT(t *testing.T) {
	pk1 := "0x041813408b94aad6745f5710e887f3c705f9cce13e3c971f5e3347cfd06bbf64fb1594c4c25392a36284f26ec89e4bc490d3815562043de013795012a782c12125"     // tb1qnv34n2wcmd94leqqs3ycrwh93pa44adhexkj9l
	pk2 := "0x04daae4253f0302725c22c9f124e6c0d182c5b4275047db181a567ff955aaebc19789155da6cd830ebc64e7ce13240c921b773681794f60e843d91d21b531584fb"     // tb1qvx32avcf6f6ke8s9h23pg7fr4d9798vs3xxrqr
	pk3 := "0x040e1c9e63a03a4274998089d6491cd5ed789a7e09efbe7994a040228646e0e45d64def988296662eb5480ecab78dda694eda605bb9a56503533f02df3f5441202"     // tb1qpgw555fq567k098wtk6346ejt4q0jxhtxeyepc
	pkAdmin := "0x043299eade0becea53c3e3943ccfbd647a103cf7654d8d7b79a11a6efeb81d8a31c396c7fa787f359608fe5b18debf4e8f242735f6049e880dd83480b403309a2a" // tb1qm39c3gupc2rfrrr6g8y8ely0ad26992lypcz5l
	locktime := time.Date(2024, 12, 21, 9, 0, 0, 0, time.Local).Unix()

	script, err := NewCustodyScriptV2([]string{pk1, pk2, pk3}, pkAdmin, 2, 3, uint32(locktime), &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(script.DisasmString())
	t.Log(script.Address())
	fmt.Println("address:", script.Address())

	cli := NewSmartClient([]string{"https://mempool.space:443/testnet/api/"})
	utxos, err := cli.GetAddressUtxos(context.Background(), script.Address())
	if err != nil {
		t.Fatal(err)
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)
	redeemTx.LockTime = uint32(script.locktime)

	// outout
	{
		decodedAddr, err := btcutil.DecodeAddress(script.Address(),
			Network("testnet3"))
		if err != nil {
			t.Fatal(err)
		}
		destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			t.Fatal(err)
		}
		// adding the destination address and the amount to the transaction
		redeemTxOut := wire.NewTxOut(99000, destinationAddrByte)
		redeemTx.AddTxOut(redeemTxOut)
	}

	// input
	var fetcherMap = make(map[wire.OutPoint]*wire.TxOut)

	for _, utxo := range utxos {
		if utxo.Value > 100000 {
			continue
		}
		chainHash, _ := chainhash.NewHashFromStr(utxo.Txid)
		prevOut := wire.NewOutPoint(chainHash, utxo.Vout)
		in := wire.NewTxIn(prevOut, nil, nil)
		in.Sequence = 0xfffffffd
		redeemTx.AddTxIn(in)

		// 为 PSBT 准备 UTXO 信息
		fetcherMap[*prevOut] = wire.NewTxOut(utxo.Value, script.pk)
	}

	packet, err := psbt.NewFromUnsignedTx(redeemTx)
	if err != nil {
		t.Fatal(err)
	}
	// 添加输入信息
	for i, in := range redeemTx.TxIn {
		utxo := fetcherMap[in.PreviousOutPoint]
		packet.Inputs[i].WitnessScript = script.redeemScript
		packet.Inputs[i].WitnessUtxo = utxo
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
	unsignedPsbt, err = cu.PsbtSign(context.Background(), "tb1qvx32avcf6f6ke8s9h23pg7fr4d9798vs3xxrqr", unsignedPsbt, nil)
	if err != nil {
		t.Fatal(err)
	}

	// admin
	signedPsbtHex, err := cu.PsbtSign(context.Background(), "tb1qm39c3gupc2rfrrr6g8y8ely0ad26992lypcz5l", unsignedPsbt, nil)
	if err != nil {
		t.Fatal(err)
	}

	// 完成 PSBT
	signedPsbtBz, err := hex.DecodeString(signedPsbtHex)
	if err != nil {
		t.Fatal(err)
	}
	packet, err = psbt.NewFromRawBytes(bytes.NewReader(signedPsbtBz), false)
	if err != nil {
		t.Fatal(err)
	}
	if err := finalizePSBTInputV2(packet, 0); err != nil {
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

func finalizePSBTInputV2(packet *psbt.Packet, index int) error {
	// 获取签名
	userSig := packet.Inputs[index].PartialSigs[0].Signature
	notary1Sig := packet.Inputs[index].PartialSigs[1].Signature
	notary2Sig := packet.Inputs[index].PartialSigs[2].Signature

	//   { asm: '<sig(U)> 0 <sig(A)> <sig(B)>' },
	//   { asm: '<sig(U)> 0 <sig(A)> <sig(C)>' },
	//   { asm: '<sig(U)> 0 <sig(B)> <sig(C)>' }
	witness := wire.TxWitness{
		userSig,
		{},
		notary1Sig,
		notary2Sig,
		packet.Inputs[index].WitnessScript,
	}

	// 设置见证数据
	packet.Inputs[index].FinalScriptWitness = serializeWitness(witness)

	// 清除不需要的字段
	packet.Inputs[index].PartialSigs = nil
	packet.Inputs[index].WitnessScript = nil

	return nil
}
