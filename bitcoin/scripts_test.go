package bitcoinlib

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestNotaryScriptPubKey(t *testing.T) {
	wifKey := "cTzQ9BHFw9Lff3P4HdctAjiLj9HRZEpZBnHaw11PYRA9KcNBcchd"
	key, err := btcutil.DecodeWIF(wifKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(common.Bytes2Hex(key.PrivKey.PubKey().SerializeCompressed()))

	ns, _ := NewNotaryBridgeScript(common.Bytes2Hex(key.PrivKey.PubKey().SerializeCompressed()), "testnet", nil)
	t.Log(ns.P2WSHAddress())
}

func TestNotaryScript(t *testing.T) {
	wifKey := "cTzQ9BHFw9Lff3P4HdctAjiLj9HRZEpZBnHaw11PYRA9KcNBcchd"
	key, err := btcutil.DecodeWIF(wifKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(common.Bytes2Hex(key.PrivKey.PubKey().SerializeCompressed()))

	network := &chaincfg.TestNet3Params

	// Create a P2PK script
	script, err := txscript.NewScriptBuilder().
		AddData(key.PrivKey.PubKey().SerializeCompressed()). // Add a public key
		AddOp(txscript.OP_CHECKSIG).                         // Add a CHECKSIG opcode
		Script()
	if err != nil {
		t.Fatal(err)
	}

	witnessScriptHash := sha256.Sum256(script)
	p2shAddr, err := btcutil.NewAddressWitnessScriptHash(witnessScriptHash[:], network)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("P2WSH Address:", p2shAddr.EncodeAddress())

	// Let's say the key is a private key object that you hold
	// Suppose the network is a Bitcoin network parameter, e.g. BTCUTIL. MainNetParams
	// Let's say utxoHash is the hash of the UTXO transaction, utxoIndex is the output index of the UTXO, and utxoAmount is the amount of the UTXO

	// Destination address
	toAddress, err := btcutil.DecodeAddress("tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp", network)
	if err != nil {
		t.Fatal(err)
	}

	// Structuring the deal
	tx := wire.NewMsgTx(wire.TxVersion)

	utxo, fetcher := GetUnspent()
	utxoAmount := int64(17120)

	// Add UTXO as input
	chainHash, _ := chainhash.NewHash(utxo.Hash[:])
	prevOut := wire.NewOutPoint(chainHash, utxo.Index)
	txIn := wire.NewTxIn(prevOut, nil, nil)
	txIn.Sequence = 0xfffffffd
	tx.AddTxIn(txIn)

	// Add the transfer amount and change
	{
		pkScript, err := txscript.PayToAddrScript(toAddress)
		if err != nil {
			t.Fatal(err)
		}
		txOut := wire.NewTxOut(10000, pkScript)
		tx.AddTxOut(txOut)
	}

	change := utxoAmount - 10000 - 1000 // 1000 is the transaction fee
	if change > 0 {
		changeAddr := p2shAddr
		changeScript, err := txscript.PayToAddrScript(changeAddr)
		if err != nil {
			t.Fatal(err)
		}
		txOut := wire.NewTxOut(change, changeScript)
		tx.AddTxOut(txOut)

	}

	// Signature Transactions
	sig, err := txscript.RawTxInWitnessSignature(
		tx,
		txscript.NewTxSigHashes(tx, fetcher),
		0,
		utxoAmount,
		script,
		txscript.SigHashAll,
		key.PrivKey)
	if err != nil {
		t.Fatal(err)
	}

	// tx.TxIn[0].SignatureScript = sigScript
	// Set the witness
	tx.TxIn[0].Witness = wire.TxWitness{sig, script}
	tx.TxIn[0].SignatureScript = nil

	// Verify the signature
	// bz, _ := hex.DecodeString("5120b9e42233471b8d58e9cb98fda7ae088f16ff6fe71733c061ce73ea96ec613cd6")
	// engine, err := txscript.NewEngine(bz, tx, 0, txscript.StandardVerifyFlags, nil, nil, utxoAmount, fetcher)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if err := engine.Execute(); err != nil {
	// 	t.Fatal(err)
	// }

	// Serialized transactions
	var serializedTx bytes.Buffer
	if err := tx.Serialize(&serializedTx); err != nil {
		// Handling errors
		t.Fatalf("Failed to serialize transaction: %v", err)
	}
	// Converts serialized transactions to hexadecimal strings
	txHex := hex.EncodeToString(serializedTx.Bytes())
	fmt.Println("Serialized Transaction Hex:", txHex)

	// Send a transaction
	url := "https://mempool.space:443/testnet/api/tx" // Modify to your Bitcoin node URL
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(txHex)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func GetUnspent() (*wire.OutPoint, *txscript.MultiPrevOutFetcher) {
	// The hash value of the transaction, and you want to specify the output location
	txHash, _ := chainhash.NewHashFromStr(
		"060ababdc23c14ec5d5dc606f790b9f594365f4ab5876684c5695b1d2af20e2d")
	point := wire.NewOutPoint(txHash, uint32(1))

	// The lock script for the transaction, which corresponds to the ScriptPubKey field
	script, _ := hex.DecodeString("02C8E1863B7C11867BCD261C8646CB2C27EEFC50B15A2D901BB63C8A9305DA2308")
	output := wire.NewTxOut(int64(17120), script)
	prevOuts := make(map[wire.OutPoint]*wire.TxOut)
	prevOuts[*point] = output
	fetcher := txscript.NewMultiPrevOutFetcher(prevOuts)
	// fetcher.AddPrevOut(*point, output)

	return point, fetcher
}

func TestBuildAndSignTx(t *testing.T) {
	wifKey := "cTzQ9BHFw9Lff3P4HdctAjiLj9HRZEpZBnHaw11PYRA9KcNBcchd"
	key, err := btcutil.DecodeWIF(wifKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(common.Bytes2Hex(key.PrivKey.PubKey().SerializeCompressed()))

	ns, _ := NewNotaryBridgeScript(common.Bytes2Hex(key.PrivKey.PubKey().SerializeCompressed()), "testnet", NewMempool(URL_MEMPOOL))
	t.Log(ns.P2WSHAddress())

	utxoOpt := WithUTXOS([]UTXO{UTXO{
		Txid:  "b13393a9c2e9cfdbefb9f02211e9b7c7562da3817592fab1f04bc9049be58e9f",
		Vout:  1,
		Value: 697606,
	}})
	feeRateOpt := WithFeeRate(2)

	txHex, usedUtxos, err := ns.BuildTx(context.Background(), "tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp", 100000, utxoOpt, feeRateOpt)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("txHex:", txHex)
	t.Log("usedUtxos:", usedUtxos)

	signedTx, txid, err := ns.SignTx(context.Background(), wifKey, txHex)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("signedTx:", signedTx)
	t.Log("txid:", txid)

	txId1, err := ns.client.PostTransaction(context.Background(), signedTx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("txId1:", txId1)
}
