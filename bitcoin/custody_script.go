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
