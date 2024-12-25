package bitcoinlib

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// MultisigTaprootScript musig P2TR script
type MultisigTaprootScript struct {
	pks     []string
	m       int
	n       int
	network *chaincfg.Params

	leafScript   []byte
	tapleafHash  []byte
	publicKey    *btcec.PublicKey
	address      string
	output       []byte
	controlBlock []byte
}

// NewMultisigTaprootScript Generate a multi-signature taproot script The input public key must be XOnly, m<=n;
// There is no order in which the signatures are signed
func NewMultisigTaprootScript(pks []string, m, n int, network *chaincfg.Params) (*MultisigTaprootScript, error) {
	if len(pks) != n {
		return nil, fmt.Errorf("number of pks does not match n")
	}
	if m > n {
		return nil, fmt.Errorf("m must lgt n")
	}
	var leafPubkeys [][]byte
	for _, v := range pks {
		if strings.HasPrefix(v, "0x") {
			v = v[2:]
		}
		if bz, err := hex.DecodeString(v); err != nil {
			return nil, err
		} else {
			leafPubkeys = append(leafPubkeys, bz)
		}
	}
	// The public key must be sorted first
	sort.Slice(leafPubkeys, func(i, j int) bool {
		return bytes.Compare(leafPubkeys[i], leafPubkeys[j]) == -1
	})

	// Build a multi-signature script
	builder := txscript.NewScriptBuilder()
	for i, pk := range leafPubkeys {
		if i == 0 {
			builder.AddData(pk)
			builder.AddOp(byte(txscript.OP_CHECKSIG))
		} else {
			builder.AddData(pk)
			builder.AddOp(byte(txscript.OP_CHECKSIGADD))
		}
	}
	builder.AddOp(byte(txscript.OP_1 - 1 + m))
	builder.AddOp(byte(txscript.OP_GREATERTHANOREQUAL))
	leafScript, err := builder.Script()
	if err != nil {
		return nil, err
	}

	// internalPubkey is a non-expendable public key whose private key no one knows and is the value recommended by the relevant standards
	// Reference: bitcoinjs-lib/test/integration/taproot.spec.ts:761
	// The prefix 02 has been added here to indicate that it is not a pk of XOnly
	var internalPubkeyBz, _ = hex.DecodeString("0250929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0") // unspendableInternalKey
	internalPubkey, err := btcec.ParsePubKey(internalPubkeyBz)
	if err != nil {
		return nil, err
	}

	// taproot tree
	scriptTree := txscript.AssembleTaprootScriptTree(txscript.TapLeaf{
		LeafVersion: txscript.BaseLeafVersion,
		Script:      leafScript,
	})
	tapleafHash := scriptTree.RootNode.TapHash()
	// Public
	taprootOutputKey := txscript.ComputeTaprootOutputKey(internalPubkey, tapleafHash.CloneBytes())

	// address
	witnessProg := schnorr.SerializePubKey(taprootOutputKey)
	tapAddr, err := btcutil.NewAddressTaproot(witnessProg, network)
	if err != nil {
		panic(err)
	}

	// Output scripts
	p2trScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_1).AddData(witnessProg).Script()
	if err != nil {
		return nil, err
	}

	// Determine if the last digit of the y coordinate is an odd number
	yIsOdd := taprootOutputKey.Y().Bit(0) == 1

	// Generate control blocks
	controlBlock := txscript.ControlBlock{
		LeafVersion:     txscript.BaseLeafVersion,
		InternalKey:     internalPubkey,
		OutputKeyYIsOdd: yIsOdd,
	}
	controlBlockBytes, err := controlBlock.ToBytes()
	if err != nil {
		return nil, err
	}
	return &MultisigTaprootScript{
		pks:          pks,
		m:            m,
		n:            n,
		network:      network,
		leafScript:   leafScript,
		tapleafHash:  tapleafHash.CloneBytes(),
		publicKey:    taprootOutputKey,
		address:      tapAddr.EncodeAddress(),
		output:       p2trScript,
		controlBlock: controlBlockBytes,
	}, nil
}

func (ms *MultisigTaprootScript) RedeemScript() []byte {
	return ms.output
}

func (ms *MultisigTaprootScript) DisasmString() string {
	redeemStr, err := txscript.DisasmString(ms.leafScript)
	if err != nil {
		panic(err)
	}
	return redeemStr
}

func (ms *MultisigTaprootScript) PubKey() []byte {
	return ms.publicKey.SerializeCompressed()
}

func (ms *MultisigTaprootScript) Address() string {
	return ms.address
}

func (ms *MultisigTaprootScript) Output() []byte {
	return ms.output
}

func (ms *MultisigTaprootScript) ControlBlock() []byte {
	return ms.controlBlock
}

func (ms *MultisigTaprootScript) LeafScript() []byte {
	return ms.leafScript
}

func (ms *MultisigTaprootScript) TapleafHash() []byte {
	return ms.tapleafHash
}
