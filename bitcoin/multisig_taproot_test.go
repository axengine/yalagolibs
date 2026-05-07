package bitcoinlib

import (
	"bytes"
	"context"
	"encoding/hex"
	"sort"
	"strings"
	"testing"

	"github.com/axengine/yalagolibs/cubist"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func TestGenTapMulsigTx2of3(t *testing.T) {
	signpks := []string{
		"tb1pcxvfy4ezehsha2h30dr9lpa8psg8axat6c0fssflwkpanlgwr7csym0cye",
		"tb1pzztghz3vlg7c07pd6fn48lchw9xsvux7gc2tluqaudg78hr97fmqews5sc",
		"tb1ps2zjfxsvalaqjyge9s4cz2c5l2mxf6augu7vez40m966qxfamqjsdjwk2l",
	}
	_ = signpks
	pks := []string{
		"be7c61b415dcf3117992b2c7941293db04f1723fea9a216d4427ee3d54d21d32",
		"c81b336d1fdef81309eb4b6b9e073d9b70638cc76700cdb5276cd1078800ef0d",
		"f84a799e9873f9a8c19cd34007bc4c2173e48be882f3a50c949210903e314a4e"}
	tapscript, err := NewMultisigTaprootScript(pks, 2, 3, &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tapscript.address)
	t.Log(tapscript.Output())
	t.Log(tapscript.ControlBlock())
	t.Log(tapscript.RedeemScript())
	t.Log(tapscript.DisasmString())
	t.Log("tapleafHash:", hex.EncodeToString(tapscript.tapleafHash))

	// Construct a transfer transaction
	const utxoHash = "bce392e4a5946e7e1f572e8234b5cc4594c71697267f2e822c0460687c0c150d"
	const utxoVout = 0
	const utxoAmount = 1000000
	const sendAmount = 1000
	const fee = 15000
	const changeAmount = utxoAmount - sendAmount - fee

	var (
		inputs     []*wire.OutPoint
		nSequences []uint32
		outputs    []*wire.TxOut
	)
	{ // output
		decodedAddr, err := btcutil.DecodeAddress("tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp",
			Network("testnet"))
		if err != nil {
			t.Fatal(err)
		}
		destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			t.Fatal(err)
		}
		// adding the destination address and the amount to the transaction
		redeemTxOut := wire.NewTxOut(sendAmount, destinationAddrByte)
		outputs = append(outputs, redeemTxOut)
	}

	{ // change
		decodedAddr, err := btcutil.DecodeAddress("tb1p4jkyr5fvkw5q36066a74xyss9x95ch6w8g253mpak2zpef6q4v7s54kywf",
			Network("testnet"))
		if err != nil {
			t.Fatal(err)
		}
		destinationAddrByte, _ := txscript.PayToAddrScript(decodedAddr)
		// adding the destination address and the amount to the transaction
		redeemTxOut := wire.NewTxOut(changeAmount, destinationAddrByte)
		outputs = append(outputs, redeemTxOut)
	}

	{ // input
		chainHash, _ := chainhash.NewHashFromStr(utxoHash)
		prevOut := wire.NewOutPoint(chainHash, utxoVout)
		inputs = append(inputs, prevOut)
		nSequences = append(nSequences, 0xffffffff)
	}

	// Build PSBT
	packet, err := psbt.New(inputs, outputs, 2, 0, nSequences)
	if err != nil {
		t.Fatal(err)
	}
	packet.Inputs[0].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
		{
			ControlBlock: tapscript.controlBlock,
			Script:       tapscript.leafScript,
			LeafVersion:  0xc0,
		},
	}

	// Add witness data
	updater, err := psbt.NewUpdater(packet)
	if err != nil {
		t.Fatal(err)
	}
	if err := updater.AddInWitnessUtxo(wire.NewTxOut(utxoAmount, tapscript.output), 0); err != nil {
		t.Fatal(err)
	}

	var serializedTx bytes.Buffer
	if err := packet.Serialize(&serializedTx); err != nil {
		t.Fatal(err)
	}
	psbtHex := hex.EncodeToString(serializedTx.Bytes())
	t.Log("before sign:", psbtHex)

	// sign
	cu := cubist.New(true, "")
	signedPsbtHex, err := cu.PsbtSign(context.Background(), signpks[1], psbtHex, nil)
	if err != nil {
		t.Fatal(err)
	}
	signedPsbtHex, err = cu.PsbtSign(context.Background(), signpks[2], signedPsbtHex, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("after sign:", signedPsbtHex)

	// Adjustment
	psbtBz, _ := hex.DecodeString(signedPsbtHex)
	packet, err = psbt.NewFromRawBytes(bytes.NewReader(psbtBz), false)
	if err != nil {
		t.Fatal(err)
	}

	for index, input := range packet.Inputs {
		for _, pk := range pks {
			if strings.HasPrefix(pk, "0x") {
				pk = pk[2:]
			}
			pkbz, _ := hex.DecodeString(pk)
			var signed bool
			for _, x := range input.TaprootScriptSpendSig {
				if bytes.Equal(pkbz, x.XOnlyPubKey) {
					signed = true
					break
				}
			}
			if !signed {
				packet.Inputs[index].TaprootScriptSpendSig = append(packet.Inputs[index].TaprootScriptSpendSig, &psbt.TaprootScriptSpendSig{
					XOnlyPubKey: pkbz,
					LeafHash:    tapscript.tapleafHash,
					Signature:   make([]byte, 0),
					SigHash:     txscript.SigHashDefault,
				})
			}
		}

		sigs := packet.Inputs[index].TaprootScriptSpendSig
		sort.Slice(sigs, func(i, j int) bool {
			return bytes.Compare(sigs[i].XOnlyPubKey, sigs[j].XOnlyPubKey) > 0
		})
		packet.Inputs[index].TaprootScriptSpendSig = sigs
	}

	for _, input := range packet.Inputs {
		for _, x := range input.TaprootScriptSpendSig {
			t.Log(x.XOnlyPubKey)
		}
	}

	{
		var serializedTx bytes.Buffer
		if err := packet.Serialize(&serializedTx); err != nil {
			t.Fatal(err)
		}
		t.Log("before finalize", hex.EncodeToString(serializedTx.Bytes()))
	}
	if err = psbt.MaybeFinalizeAll(packet); err != nil {
		t.Fatal(err)
	}
	{
		var serializedTx bytes.Buffer
		if err := packet.Serialize(&serializedTx); err != nil {
			t.Fatal(err)
		}
		t.Log("after finalize", hex.EncodeToString(serializedTx.Bytes())) // not eqaul
		msgTx, err := psbt.Extract(packet)
		if err != nil {
			t.Fatal(err)
		}
		txid := msgTx.TxHash().String()
		var signedTxBuf bytes.Buffer
		if err := msgTx.Serialize(&signedTxBuf); err != nil {
			t.Fatal(err)
		}
		signedTx := hex.EncodeToString(signedTxBuf.Bytes())
		t.Log("txid", txid)
		t.Log("signedTx", signedTx)

		cli := NewSmartClient([]string{"https://mempool.space:443/testnet/api/"})
		if _, err = cli.PostTransaction(context.Background(), signedTx); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTxSize(t *testing.T) {
	pks := []string{
		"be7c61b415dcf3117992b2c7941293db04f1723fea9a216d4427ee3d54d21d32",
		"c81b336d1fdef81309eb4b6b9e073d9b70638cc76700cdb5276cd1078800ef0d",
		"f84a799e9873f9a8c19cd34007bc4c2173e48be882f3a50c949210903e314a4e"}
	tapscript, err := NewMultisigTaprootScript(pks, 2, 3, &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(tapscript.LeafScript()))
	t.Log(tapscript.TxSize(2, 2))
}

func TestGenTapMulsigTx3of5(t *testing.T) {
	signpks := []string{
		"tb1pcxvfy4ezehsha2h30dr9lpa8psg8axat6c0fssflwkpanlgwr7csym0cye",
		"tb1pzztghz3vlg7c07pd6fn48lchw9xsvux7gc2tluqaudg78hr97fmqews5sc",
		"tb1ps2zjfxsvalaqjyge9s4cz2c5l2mxf6augu7vez40m966qxfamqjsdjwk2l",
		"tb1pvald49wpgpfqxyzqeh3hh6xy53f6k9f6jds7hln2nd8wd9wcl3gsp0kvje",
		"tb1pegq9cl3fg429ar8ultsxg9tc9hdmyfcnsp7dwz0ep2emkjy6w29qzm38sm",
	}
	_ = signpks
	pks := []string{
		"be7c61b415dcf3117992b2c7941293db04f1723fea9a216d4427ee3d54d21d32",
		"c81b336d1fdef81309eb4b6b9e073d9b70638cc76700cdb5276cd1078800ef0d",
		"f84a799e9873f9a8c19cd34007bc4c2173e48be882f3a50c949210903e314a4e",
		"0xd2015f75bf3fb0d85d33cf80e37a46b0eabf434bc206d80ca3071ffb4db5c23b",
		"0x7440fb0f2b658c7d4290a6c2cb31374e1333d2ee71206ba570dfbc3b6b37d7bb",
	}
	tapscript, err := NewMultisigTaprootScript(pks, 3, 5, &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tapscript.address)

	var changeAddress = tapscript.address

	// Construct a transfer transaction
	const utxoHash = "0c1e996de92d58e1816bbc6ddafdf7fe35dac158ad9aab325441edd35040fe3d"
	const utxoVout = 0
	const utxoAmount = 1000000
	const sendAmount = 1000
	const fee = 240
	const changeAmount = utxoAmount - sendAmount - fee
	const receiver = "tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp"

	var (
		inputs     []*wire.OutPoint
		nSequences []uint32
		outputs    []*wire.TxOut
	)
	{ // output
		decodedAddr, err := btcutil.DecodeAddress(receiver, Network("testnet"))
		if err != nil {
			t.Fatal(err)
		}
		destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			t.Fatal(err)
		}
		// adding the destination address and the amount to the transaction
		redeemTxOut := wire.NewTxOut(sendAmount, destinationAddrByte)
		outputs = append(outputs, redeemTxOut)
	}

	{ // change
		decodedAddr, err := btcutil.DecodeAddress(changeAddress, Network("testnet"))
		if err != nil {
			t.Fatal(err)
		}
		destinationAddrByte, _ := txscript.PayToAddrScript(decodedAddr)
		// adding the destination address and the amount to the transaction
		redeemTxOut := wire.NewTxOut(changeAmount, destinationAddrByte)
		outputs = append(outputs, redeemTxOut)
	}

	{ // input
		chainHash, _ := chainhash.NewHashFromStr(utxoHash)
		prevOut := wire.NewOutPoint(chainHash, utxoVout)
		inputs = append(inputs, prevOut)
		nSequences = append(nSequences, 0xffffffff)
	}

	// Build PSBT
	packet, err := psbt.New(inputs, outputs, 2, 0, nSequences)
	if err != nil {
		t.Fatal(err)
	}
	packet.Inputs[0].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
		{
			ControlBlock: tapscript.controlBlock,
			Script:       tapscript.leafScript,
			LeafVersion:  0xc0,
		},
	}

	// Add witness data
	updater, err := psbt.NewUpdater(packet)
	if err != nil {
		t.Fatal(err)
	}
	if err := updater.AddInWitnessUtxo(wire.NewTxOut(utxoAmount, tapscript.output), 0); err != nil {
		t.Fatal(err)
	}

	var serializedTx bytes.Buffer
	if err := packet.Serialize(&serializedTx); err != nil {
		t.Fatal(err)
	}
	psbtHex := hex.EncodeToString(serializedTx.Bytes())
	t.Log("before sign:", psbtHex)

	// sign
	cu := cubist.New(true, "")
	signedPsbtHex, err := cu.PsbtSign(context.Background(), signpks[1], psbtHex, nil)
	if err != nil {
		t.Fatal(err)
	}
	signedPsbtHex, err = cu.PsbtSign(context.Background(), signpks[2], signedPsbtHex, nil)
	if err != nil {
		t.Fatal(err)
	}
	signedPsbtHex, err = cu.PsbtSign(context.Background(), signpks[4], signedPsbtHex, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("after sign:", signedPsbtHex)

	// Adjustment
	psbtBz, _ := hex.DecodeString(signedPsbtHex)
	packet, err = psbt.NewFromRawBytes(bytes.NewReader(psbtBz), false)
	if err != nil {
		t.Fatal(err)
	}

	for index, input := range packet.Inputs {
		for _, pk := range pks {
			if strings.HasPrefix(pk, "0x") {
				pk = pk[2:]
			}
			pkbz, _ := hex.DecodeString(pk)
			var signed bool
			for _, x := range input.TaprootScriptSpendSig {
				if bytes.Equal(pkbz, x.XOnlyPubKey) {
					signed = true
					break
				}
			}
			if !signed {
				packet.Inputs[index].TaprootScriptSpendSig = append(packet.Inputs[index].TaprootScriptSpendSig, &psbt.TaprootScriptSpendSig{
					XOnlyPubKey: pkbz,
					LeafHash:    tapscript.tapleafHash,
					Signature:   make([]byte, 0),
					SigHash:     txscript.SigHashDefault,
				})
			}
		}

		sigs := packet.Inputs[index].TaprootScriptSpendSig
		sort.Slice(sigs, func(i, j int) bool {
			return bytes.Compare(sigs[i].XOnlyPubKey, sigs[j].XOnlyPubKey) > 0
		})
		packet.Inputs[index].TaprootScriptSpendSig = sigs
	}

	for _, input := range packet.Inputs {
		for _, x := range input.TaprootScriptSpendSig {
			t.Log(x.XOnlyPubKey)
		}
	}

	{
		var serializedTx bytes.Buffer
		if err := packet.Serialize(&serializedTx); err != nil {
			t.Fatal(err)
		}
		t.Log("before finalize", hex.EncodeToString(serializedTx.Bytes()))
	}
	if err = psbt.MaybeFinalizeAll(packet); err != nil {
		t.Fatal(err)
	}
	{
		var serializedTx bytes.Buffer
		if err := packet.Serialize(&serializedTx); err != nil {
			t.Fatal(err)
		}
		t.Log("after finalize", hex.EncodeToString(serializedTx.Bytes())) // not eqaul
		msgTx, err := psbt.Extract(packet)
		if err != nil {
			t.Fatal(err)
		}
		txid := msgTx.TxHash().String()
		var signedTxBuf bytes.Buffer
		if err := msgTx.Serialize(&signedTxBuf); err != nil {
			t.Fatal(err)
		}
		signedTx := hex.EncodeToString(signedTxBuf.Bytes())
		t.Log("txid", txid)
		t.Log("signedTx", signedTx)

		cli := NewSmartClient([]string{"https://mempool.space:443/testnet/api/"})
		if _, err = cli.PostTransaction(context.Background(), signedTx); err != nil {
			t.Fatal(err)
		}
	}
}
