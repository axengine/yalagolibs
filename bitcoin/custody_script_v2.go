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

type CustodyScriptV2 struct {
	pks          []string
	m            int
	n            int
	locktime     uint32
	network      *chaincfg.Params
	redeemScript []byte
	pk           []byte
}

// policy: and(or(after(1751528180),thresh(2,pk(A),pk(B),pk(C))),pk(U))
// miniscript: and_v(or_c(multi(2,A,B,C),v:after(1751528180)),pk(U))
// asm: <A> <B> <C> 3 OP_CHECKMULTISIG OP_NOTIF <f4326668> OP_CHECKLOCKTIMEVERIFY OP_VERIFY OP_ENDIF <U> OP_CHECKSIG
// nonMalleableSats [
//
//	{ nLockTime: 1751528180, asm: '<sig(U)> 0 0 0' },
//	{ asm: '<sig(U)> 0 <sig(A)> <sig(B)>' },
//	{ asm: '<sig(U)> 0 <sig(A)> <sig(C)>' },
//	{ asm: '<sig(U)> 0 <sig(B)> <sig(C)>' }
//
// ]
// NewCustodyScriptV2 creates a new custody script with the following format:
// asm: "2 <A> <B> <C> 3 OP_CHECKMULTISIG OP_NOTIF <f4326668> OP_CHECKLOCKTIMEVERIFY OP_VERIFY OP_ENDIF <U> OP_CHECKSIG"
func NewCustodyScriptV2(pks []string, pkAdmin string, m, n int, locktime uint32, network *chaincfg.Params) (*CustodyScriptV2, error) {
	builder := txscript.NewScriptBuilder()

	// Add m value for multisig (2 in this case)
	builder.AddOp(byte(txscript.OP_RESERVED + m))

	// Add public keys A, B, C
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

	// Add n value for multisig (3 in this case)
	builder.AddOp(byte(txscript.OP_RESERVED + n))
	builder.AddOp(txscript.OP_CHECKMULTISIG)

	// Add OP_NOTIF for timelock branch
	builder.AddOp(txscript.OP_NOTIF)

	// Add locktime value and verification
	var bz = make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, locktime)
	builder.AddData(bz)
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY)
	builder.AddOp(txscript.OP_VERIFY)

	// Close timelock branch
	builder.AddOp(txscript.OP_ENDIF)

	// Add admin pubkey and checksig
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

	return &CustodyScriptV2{
		pks:          pks,
		m:            m,
		n:            n,
		locktime:     locktime,
		network:      network,
		redeemScript: redeemScript,
		pk:           destinationAddrByte,
	}, nil
}

func (ms *CustodyScriptV2) RedeemScript() []byte {
	return ms.redeemScript
}

func (ms *CustodyScriptV2) DisasmString() string {
	redeemStr, err := txscript.DisasmString(ms.redeemScript)
	if err != nil {
		panic(err)
	}
	return redeemStr
}

func (ms *CustodyScriptV2) PubKey() string {
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

func (ms *CustodyScriptV2) PK() []byte {
	return ms.pk
}

func (ms *CustodyScriptV2) Address() string {
	witnessScriptHash := chainhash.HashB(ms.redeemScript)
	scriptAddr, err := btcutil.NewAddressWitnessScriptHash(witnessScriptHash, ms.network) //P2WSH
	if err != nil {
		panic(err)
	}

	return scriptAddr.EncodeAddress()
}

func (ms *CustodyScriptV2) Locktime() uint32 {
	return ms.locktime
}
