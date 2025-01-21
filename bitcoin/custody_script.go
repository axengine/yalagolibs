package bitcoinlib

import (
	"encoding/binary"
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
)

// CustodyScript script for self-custody
type CustodyScript struct {
	pks          []string
	m            int
	n            int
	locktime     uint32
	network      *chaincfg.Params
	redeemScript []byte
	pk           []byte
}

func NewCustodyScript(pks []string, pkAdmin string, m, n int, locktime uint32, network *chaincfg.Params) (*CustodyScript, error) {
	builder := txscript.NewScriptBuilder()

	builder.AddOp(txscript.OP_IF)
	// in-locktime
	{
		var bz = make([]byte, 4)
		binary.LittleEndian.PutUint32(bz, locktime) //
		builder.AddData(bz)
		builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY)
		builder.AddOp(txscript.OP_DROP)
	}
	builder.AddOp(txscript.OP_ELSE)
	{
		// musig
		builder.AddOp(byte(txscript.OP_RESERVED + m))

		for _, v := range pks {
			v = strings.TrimPrefix(v, "0x")
			bz, err := hex.DecodeString(v)
			if err != nil {
				return nil, err
			}
			if len(v) > 66 {
				if pk, err := btcec.ParsePubKey(bz); err != nil {
					return nil, err
				} else {
					bz = pk.SerializeCompressed()
				}
			}
			builder.AddData(bz)
		}
		builder.AddOp(byte(txscript.OP_RESERVED + n))
		builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)
	}
	builder.AddOp(txscript.OP_ENDIF)
	// admin
	pkAdmin = strings.TrimPrefix(pkAdmin, "0x")
	bz, err := hex.DecodeString(pkAdmin)
	if err != nil {
		return nil, err
	}
	if len(pkAdmin) > 66 {
		if pk, err := btcec.ParsePubKey(bz); err != nil {
			return nil, err
		} else {
			bz = pk.SerializeCompressed()
		}
	}
	builder.AddData(bz)
	builder.AddOp(txscript.OP_CHECKSIG)

	redeemScript, err := builder.Script()
	if err != nil {
		return nil, err
	}
	witnessScriptHash := chainhash.HashB(redeemScript)
	multisigAddr, _ := btcutil.NewAddressWitnessScriptHash(witnessScriptHash[:], network)
	decodedAddr, _ := btcutil.DecodeAddress(multisigAddr.EncodeAddress(), network)
	destinationAddrByte, _ := txscript.PayToAddrScript(decodedAddr)

	return &CustodyScript{
		pks:          pks,
		m:            m,
		n:            n,
		locktime:     locktime,
		network:      network,
		redeemScript: redeemScript,
		pk:           destinationAddrByte,
	}, nil
}

func (ms *CustodyScript) RedeemScript() []byte {
	return ms.redeemScript
}

func (ms *CustodyScript) DisasmString() string {
	redeemStr, err := txscript.DisasmString(ms.redeemScript)
	if err != nil {
		panic(err)
	}
	return redeemStr
}

func (ms *CustodyScript) PubKey() string {
	hash := chainhash.HashB(ms.redeemScript)
	// Create a script pub key for p2 wsh
	scriptPubKey, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).
		AddData(hash[:]).
		Script()
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(scriptPubKey[:])
}

func (ms *CustodyScript) PK() []byte {
	return ms.pk
}

func (ms *CustodyScript) Address() string {
	witnessScriptHash := chainhash.HashB(ms.redeemScript)
	scriptAddr, err := btcutil.NewAddressWitnessScriptHash(witnessScriptHash, ms.network) //P2WSH
	if err != nil {
		panic(err)
	}

	return scriptAddr.EncodeAddress()
}

func (ms *CustodyScript) Locktime() uint32 {
	return ms.locktime
}

// TxSizeWithinLocktime The estimated size is on the large side.
func (ms *CustodyScript) TxSizeWithinLocktime(input, output int) int {
	var basesize int
	basesize += 4 // version

	basesize += 1 // input count assumes inputCount<=255
	for i := 0; i < input; i++ {
		inputsize := 32 + 4 // outpoint(hash+index)
		inputsize += 1      // len SignatureScript
		inputsize += 0      // SignatureScript
		inputsize += 4      // sequence
		basesize += inputsize
	}

	basesize += 1 // output count assumes outputCount<=255
	for i := 0; i < output; i++ {
		outputsize := 8  // value
		outputsize += 1  // pk len
		outputsize += 34 // pk assumes that pk length is 34
		basesize += outputsize
	}

	basesize += 4 // nlocktime

	witnesssize := 0
	witnesssize += 2 // flag
	for i := 0; i < input; i++ {
		witnesssizePerInput := 0
		witnesssizePerInput += 1 // witness len, assuming nOfMusig<=255

		for j := 0; j < ms.n-ms.m; j++ { // There is no signed public key
			witnesssizePerInput += 1 // witness buf len
			witnesssizePerInput += 0
		}
		for j := 0; j < ms.m+1; j++ { // The public key of the signatureublic key of the signature
			witnesssizePerInput += 1  // witness buf len
			witnesssizePerInput += 64 // signature
		}

		witnesssizePerInput += 1 // false
		witnesssizePerInput += 1 // false

		witnesssizePerInput += 1 // witness buf len, assuming scriptLen《=255
		witnesssizePerInput += len(ms.redeemScript)

		witnesssizePerInput += 1 // script pk
		witnesssizePerInput += 33

		witnesssize += witnesssizePerInput
	}

	// doc：https://bitcoin.stackexchange.com/questions/114375/how to accurately calculate the vsize of a transaction
	totalsize := basesize + witnesssize

	//fmt.Println("basesize=", basesize, "witnesssize=", witnesssize)
	weight := 3*basesize + totalsize // vsize = basesize + (totalsize-basesize)/4
	vsize := (weight + 4 - 1) / 4    // Not so precise
	return vsize
}

// TxSizeOutsideLocktime The estimated size is on the large side.
func (ms *CustodyScript) TxSizeOutsideLocktime(input, output int) int {
	var basesize int
	basesize += 4 // version

	basesize += 1 // input count assumes inputCount<=255
	for i := 0; i < input; i++ {
		inputsize := 32 + 4 // outpoint(hash+index)
		inputsize += 1      // len SignatureScript
		inputsize += 0      // SignatureScript
		inputsize += 4      // sequence
		basesize += inputsize
	}

	basesize += 1 // output count assumes outputCount<=255
	for i := 0; i < output; i++ {
		outputsize := 8  // value
		outputsize += 1  // pk len
		outputsize += 34 // pk assumes that pk length is 34
		basesize += outputsize
	}

	basesize += 4 // nlocktime

	witnesssize := 0
	witnesssize += 2 // flag
	for i := 0; i < input; i++ {
		witnesssizePerInput := 0
		witnesssizePerInput += 1 // witness len, assuming nOfMusig<=255

		for j := 0; j < ms.n-ms.m; j++ { // There is no signed public key
			witnesssizePerInput += 1 // witness buf len
			witnesssizePerInput += 0
		}

		// admin signature
		witnesssizePerInput += 1  // witness buf len
		witnesssizePerInput += 64 // signature

		witnesssizePerInput += 1 // true

		witnesssizePerInput += 1 // witness buf len, assuming scriptLen《=255
		witnesssizePerInput += len(ms.redeemScript)

		witnesssizePerInput += 1 // script pk
		witnesssizePerInput += 33

		witnesssize += witnesssizePerInput
	}

	// doc：https://bitcoin.stackexchange.com/questions/114375/how to accurately calculate the vsize of a transaction
	totalsize := basesize + witnesssize

	//fmt.Println("basesize=", basesize, "witnesssize=", witnesssize)
	weight := 3*basesize + totalsize // vsize = basesize + (totalsize-basesize)/4
	vsize := (weight + 4 - 1) / 4    // Not so precise
	return vsize
}
