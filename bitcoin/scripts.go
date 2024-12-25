package bitcoinlib

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common"
)

// Multisig P2SH script
type NotaryBridgeScript struct {
	network            *chaincfg.Params
	notaryPublicKeyHex string
	script             []byte
	p2wshAddress       string

	client IBitcoinClient

	utxos map[string]*UTXO
}

func NewNotaryBridgeScript(notaryPublicKeyHex string, network string, client IBitcoinClient) (*NotaryBridgeScript, error) {
	n := Network(network)
	notaryPublicKey, err := hex.DecodeString(notaryPublicKeyHex)
	if err != nil {
		return nil, err
	}

	// Resolve the public key
	pubKey, err := btcec.ParsePubKey(notaryPublicKey)
	if err != nil {
		return nil, err
	}

	// Create a P2PK script
	script, err := txscript.NewScriptBuilder().
		AddData(pubKey.SerializeCompressed()). // Add a public key
		AddOp(txscript.OP_CHECKSIG).           // Add a CHECKSIG opcode
		Script()
	if err != nil {
		return nil, err
	}

	witnessScriptHash := sha256.Sum256(script)
	p2wshAddr, err := btcutil.NewAddressWitnessScriptHash(witnessScriptHash[:], n)
	if err != nil {
		return nil, err
	}

	return &NotaryBridgeScript{
		script:             script,
		p2wshAddress:       p2wshAddr.EncodeAddress(),
		notaryPublicKeyHex: notaryPublicKeyHex,
		network:            n,
		client:             client,
		utxos:              make(map[string]*UTXO),
	}, nil
}

func (ns *NotaryBridgeScript) DisasmString() (string, error) {
	return txscript.DisasmString(ns.script)
}

func (ns *NotaryBridgeScript) P2SHAddress() (string, error) {
	p2shAddr, err := btcutil.NewAddressScriptHash(ns.script, ns.network)
	if err != nil {
		return "", err
	}
	return p2shAddr.EncodeAddress(), nil
}

func (ns *NotaryBridgeScript) P2WSHAddress() string {
	return ns.p2wshAddress
}

type BuildTxReq struct {
	utxos   []UTXO
	feeRate int64
}

type TxOption func(*BuildTxReq)

func WithUTXOS(utxos []UTXO) TxOption {
	return func(btr *BuildTxReq) {
		btr.utxos = utxos
	}
}

func WithFeeRate(feeRate int64) TxOption {
	return func(btr *BuildTxReq) {
		btr.feeRate = feeRate
	}
}

type To struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

func (ns *NotaryBridgeScript) utxoPrevOutFetcher(utxos []UTXO) *txscript.MultiPrevOutFetcher {
	prevOuts := make(map[wire.OutPoint]*wire.TxOut)
	for _, utxo := range utxos {
		// The hash value of the transaction, and you want to specify the output location
		txHash, _ := chainhash.NewHashFromStr(utxo.Txid)
		point := wire.NewOutPoint(txHash, uint32(utxo.Vout))

		output := wire.NewTxOut(utxo.Value, common.Hex2Bytes(ns.notaryPublicKeyHex)) // The lock script for the transaction, which corresponds to the ScriptPubKey field
		prevOuts[*point] = output

		ns.utxos[fmt.Sprintf("%s:%d", utxo.Txid, utxo.Vout)] = &utxo
		//fmt.Println("onbuild", utils.JsonPretty(point), utils.JsonPretty(output))
	}

	fetcher := txscript.NewMultiPrevOutFetcher(prevOuts)

	return fetcher
}

func (ns *NotaryBridgeScript) getUtxoPreOutputViaNetwork(ctx context.Context, txid string, vout int) (*wire.TxOut, error) {
	tx, err := ns.client.GetTransaction(ctx, txid)
	if err != nil {
		return nil, err
	}
	out := tx.Vout[vout]
	output := wire.NewTxOut(out.Value, common.Hex2Bytes(out.ScriptPubkey)) //The lock script for the transaction, which corresponds to the ScriptPubKey field
	return output, nil
}

// BuildTx builds transactions, returns unsigned transactions, and used UTXOs
func (ns *NotaryBridgeScript) BuildTx(ctx context.Context, to string, amount int64, opts ...TxOption) (string, []UTXO, error) {
	var req = BuildTxReq{}
	for _, opt := range opts {
		opt(&req)
	}

	if req.feeRate == 0 {
		feeRate, err := ns.client.GetFeeRate(ctx)
		if err != nil {
			return "", nil, err
		}
		req.feeRate = feeRate
	}
	if len(req.utxos) == 0 {
		utxos, err := ns.client.GetAddressUtxos(ctx, ns.p2wshAddress)
		if err != nil {
			return "", nil, err
		}
		req.utxos = utxos
	}

	toAddress, err := btcutil.DecodeAddress(to, ns.network)
	if err != nil {
		return "", nil, err
	}

	// Structuring the deal
	tx := wire.NewMsgTx(wire.TxVersion)
	var inputAmount = int64(0)
	var estimatedSize = 10 // The fixed part is about 10 bytes: version number (4 bytes) + input count (1 byte) + output count (1 byte) + lock time (4 bytes)

	// The size of each input
	const p2wshInputSize = 32 + 4 + 4 + 1 + 32 // Each input includes the transaction hash of the previous output 32, the output index 4, the serial number 4, the unlock script length 1 9, and the unlock script
	const outputSize = 8 + 1 + 32              // Amount 8 + Lock Script Long 1 9 + Lock Script

	estimatedSize += outputSize * 2 // There are two outputs by default

	// What this step does is cache the UTXO for Fetcher to use
	fetcher := ns.utxoPrevOutFetcher(req.utxos)
	_ = fetcher

	var usedUtxos []UTXO
	// Add all UTXOs as inputs and estimate the cost
	for _, utxo := range req.utxos {
		usedUtxos = append(usedUtxos, utxo)
		estimatedSize += p2wshInputSize
		inputAmount += utxo.Value
		currentFee := int64(estimatedSize) * req.feeRate

		// Add UTXO as input
		chainHash, _ := chainhash.NewHashFromStr(utxo.Txid)
		prevOut := wire.NewOutPoint(chainHash, utxo.Vout)
		txIn := wire.NewTxIn(prevOut, nil, nil)
		txIn.Sequence = 0xfffffffd
		tx.AddTxIn(txIn)

		if inputAmount >= currentFee+amount {
			break
		}
	}
	// Make sure that the total amount entered is sufficient to cover the amount and the estimated cost
	if inputAmount < amount+int64(estimatedSize)*req.feeRate {
		return "", nil, errors.New("not enough funds to cover the transaction and fees")
	}

	// Add a receiver output
	{
		pkScript, err := txscript.PayToAddrScript(toAddress)
		if err != nil {
			return "", nil, err
		}
		txOut := wire.NewTxOut(amount, pkScript)
		tx.AddTxOut(txOut)
	}

	// Calculate the change, and add the change output (if needed)
	changeAmount := inputAmount - amount - int64(estimatedSize)*req.feeRate
	if changeAmount > 0 {
		witnessScriptHash := sha256.Sum256(ns.script)
		p2wshAddr, _ := btcutil.NewAddressWitnessScriptHash(witnessScriptHash[:], ns.network)

		changeScript, err := txscript.PayToAddrScript(p2wshAddr)
		if err != nil {
			return "", nil, err
		}
		txOut := wire.NewTxOut(changeAmount, changeScript)
		tx.AddTxOut(txOut)
		log.Printf("changeAmount:%d", changeAmount)
	}

	log.Printf("estimatedSize %d feeTotal %d\n", estimatedSize, int64(estimatedSize)*req.feeRate)
	var serializedTx bytes.Buffer
	if err := tx.Serialize(&serializedTx); err != nil {
		return "", nil, err
	}
	// Converts serialized transactions to hexadecimal strings
	return hex.EncodeToString(serializedTx.Bytes()), usedUtxos, nil
}

// BuildBatchTx builds the transaction, returns the unsigned transaction, the UTXO used, and tos.key=receive tos.value=satoshi
func (ns *NotaryBridgeScript) BuildBatchTx(ctx context.Context, toAddressList []string, toAmountList []int64, opts ...TxOption) (string, []UTXO, error) {
	if len(toAddressList) != len(toAmountList) {
		return "", nil, errors.New("not enough addresses to build the transaction")
	}
	var req = BuildTxReq{}
	for _, opt := range opts {
		opt(&req)
	}

	if req.feeRate == 0 {
		feeRate, err := ns.client.GetFeeRate(ctx)
		if err != nil {
			return "", nil, err
		}
		req.feeRate = feeRate
	}
	if len(req.utxos) == 0 {
		utxos, err := ns.client.GetAddressUtxos(ctx, ns.p2wshAddress)
		if err != nil {
			return "", nil, err
		}
		req.utxos = utxos
	}

	// The total amount of output
	var amount int64
	for _, v := range toAmountList {
		amount += v
	}

	// Structuring the deal
	tx := wire.NewMsgTx(wire.TxVersion)
	var inputAmount = int64(0)
	var estimatedSize = 10 // The fixed part is about 10 bytes: version number (4 bytes) + input count (1 byte) + output count (1 byte) + lock time (4 bytes)

	// The size of each input
	const p2wshInputSize = 32 + 4 + 4 + 1 + 32 // Each input includes the transaction hash of the previous output 32, the output index 4, the serial number 4, the unlock script length 1 9, and the unlock script
	const outputSize = 8 + 1 + 32              // Amount 8 + Lock Script Long 1 9 + Lock Script

	estimatedSize += outputSize * (len(toAmountList) + 1) // Number of outputs + change

	// What this step does is cache the UTXO for Fetcher to use
	fetcher := ns.utxoPrevOutFetcher(req.utxos)
	_ = fetcher

	var usedUtxos []UTXO
	// Add all UTXOs as inputs and estimate the cost
	for _, utxo := range req.utxos {
		//fmt.Println("use utxo:", utils.JsonPretty(utxo))
		usedUtxos = append(usedUtxos, utxo)
		estimatedSize += p2wshInputSize
		inputAmount += utxo.Value
		currentFee := int64(estimatedSize) * req.feeRate

		// Add UTXO as input
		chainHash, _ := chainhash.NewHashFromStr(utxo.Txid)
		prevOut := wire.NewOutPoint(chainHash, utxo.Vout)
		txIn := wire.NewTxIn(prevOut, nil, nil)
		txIn.Sequence = 0xfffffffd
		tx.AddTxIn(txIn)

		if inputAmount >= currentFee+amount {
			break
		}
	}
	// Make sure that the total amount entered is sufficient to cover the amount and the estimated cost
	if inputAmount < amount+int64(estimatedSize)*req.feeRate {
		return "", nil, errors.New("not enough funds to cover the transaction and fees")
	}

	// Add a receiver output
	{
		for i, v := range toAddressList {
			toAddress, err := btcutil.DecodeAddress(v, ns.network)
			if err != nil {
				return "", nil, err
			}
			pkScript, err := txscript.PayToAddrScript(toAddress)
			if err != nil {
				return "", nil, err
			}
			txOut := wire.NewTxOut(toAmountList[i], pkScript)
			tx.AddTxOut(txOut)
		}
	}

	// Calculate the change, and add the change output (if needed)
	changeAmount := inputAmount - amount - int64(estimatedSize)*req.feeRate
	if changeAmount > 0 {
		witnessScriptHash := sha256.Sum256(ns.script)
		p2wshAddr, _ := btcutil.NewAddressWitnessScriptHash(witnessScriptHash[:], ns.network)

		changeScript, err := txscript.PayToAddrScript(p2wshAddr)
		if err != nil {
			return "", nil, err
		}
		txOut := wire.NewTxOut(changeAmount, changeScript)
		tx.AddTxOut(txOut)
		log.Printf("changeAmount:%d", changeAmount)
	}

	log.Printf("estimatedSize %d feeTotal %d\n", estimatedSize, int64(estimatedSize)*req.feeRate)
	var serializedTx bytes.Buffer
	if err := tx.Serialize(&serializedTx); err != nil {
		return "", nil, err
	}
	// Converts serialized transactions to hexadecimal strings
	return hex.EncodeToString(serializedTx.Bytes()), usedUtxos, nil
}

// SignTx uses wifKey to sign txHex and returns the signed txHex and txId
func (ns *NotaryBridgeScript) SignTx(ctx context.Context, wifKey string, txHex string) (string, string, error) {
	key, err := btcutil.DecodeWIF(wifKey)
	if err != nil {
		return "", "", err
	}
	// Decode hexadecimal strings into byte streams
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return "", "", err
	}

	// Deserialize the byte stream as a transaction
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return "", "", err
	}

	// Build fetcher
	var fetcher *txscript.MultiPrevOutFetcher
	{
		prevOuts := make(map[wire.OutPoint]*wire.TxOut)
		for _, in := range tx.TxIn {
			// Output point
			outPoint := wire.NewOutPoint(&in.PreviousOutPoint.Hash, uint32(in.PreviousOutPoint.Index))

			// Build fetcher
			utxo := ns.utxos[in.PreviousOutPoint.Hash.String()]
			if utxo != nil {
				output := wire.NewTxOut(utxo.Value, common.Hex2Bytes(ns.notaryPublicKeyHex)) //The lock script for the transaction, which corresponds to the ScriptPubKey field
				prevOuts[*outPoint] = output
			} else {
				output, err := ns.getUtxoPreOutputViaNetwork(ctx, in.PreviousOutPoint.Hash.String(), int(in.PreviousOutPoint.Index))
				if err != nil {
					return "", "", err
				}
				prevOuts[*outPoint] = output
			}
		}
		fetcher = txscript.NewMultiPrevOutFetcher(prevOuts)
	}

	for idx, in := range tx.TxIn {
		// Output point
		outPoint := wire.NewOutPoint(&in.PreviousOutPoint.Hash, uint32(in.PreviousOutPoint.Index))

		txOut := fetcher.FetchPrevOutput(*outPoint)

		// Signature Transactions
		sig, err := txscript.RawTxInWitnessSignature(
			&tx,
			txscript.NewTxSigHashes(&tx, fetcher),
			idx,
			txOut.Value,
			ns.script,
			txscript.SigHashAll,
			key.PrivKey)
		if err != nil {
			return "", "", err
		}

		in.Witness = wire.TxWitness{sig, ns.script}
		in.SignatureScript = nil
	}

	txid := tx.TxID()

	var serializedTx bytes.Buffer
	if err := tx.Serialize(&serializedTx); err != nil {
		return "", "", err
	}
	// Converts serialized transactions to hexadecimal strings
	return hex.EncodeToString(serializedTx.Bytes()), txid, nil
}
