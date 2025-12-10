package bitcoinlib

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
)

// MultisigScript m/n multisig scripts,p2wsh
type MultisigScript struct {
	pks          []string
	m            int
	n            int
	network      *chaincfg.Params
	redeemScript []byte
}

func NewMultisigScript(pks []string, m, n int, network *chaincfg.Params) (*MultisigScript, error) {
	builder := txscript.NewScriptBuilder()

	// multisig
	{
		// add the minimum number of needed signatures
		builder.AddOp(byte(txscript.OP_RESERVED + m))

		for _, v := range pks {
			bz, _ := hex.DecodeString(v)
			builder.AddData(bz)
		}
		// add the total number of public keys in the multi-sig screipt
		builder.AddOp(byte(txscript.OP_RESERVED + n))
		// add the check-multi-sig op-code
		builder.AddOp(txscript.OP_CHECKMULTISIG)
	}
	redeemScript, err := builder.Script()
	if err != nil {
		return nil, err
	}
	return &MultisigScript{
		pks:          pks,
		m:            m,
		n:            n,
		network:      network,
		redeemScript: redeemScript,
	}, nil
}

func (ms *MultisigScript) RedeemScript() []byte {
	return ms.redeemScript
}

func (ms *MultisigScript) DisasmString() string {
	redeemStr, err := txscript.DisasmString(ms.redeemScript)
	if err != nil {
		panic(err)
	}
	return redeemStr
}

func (ms *MultisigScript) PubKey() string {
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

func (ms *MultisigScript) Address() string {
	witnessScriptHash := chainhash.HashB(ms.redeemScript)
	scriptAddr, err := btcutil.NewAddressWitnessScriptHash(witnessScriptHash, ms.network) //P2WSH
	if err != nil {
		panic(err)
	}

	return scriptAddr.EncodeAddress()
}
