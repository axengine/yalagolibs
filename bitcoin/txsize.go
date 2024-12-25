package bitcoinlib

import (
	"bytes"

	"github.com/btcsuite/btcd/wire"
)

func CalculateVSize(tx *wire.MsgTx) int {
	// Serialize the transaction without witness data
	var buf bytes.Buffer
	_ = tx.SerializeNoWitness(&buf)
	realSize := buf.Len()

	// Serialize the transaction with witness data
	buf.Reset()
	_ = tx.Serialize(&buf)
	totalSize := buf.Len()

	// Calculate the witness size
	witnessSize := totalSize - realSize

	// Calculate the virtual size
	vSize := (realSize*3 + witnessSize) / 4
	return vSize
}

func CalculateP2WSHVSize(inputCount, outputCount int) int {
	return inputCount*109 + outputCount*43 + 10
}

func CalculateP2WSHVSizeV2(inputCount, outputCount int, mOfMusig, nOfMusig int, scriptLen int) int {
	var basesize int
	basesize += 4 // version

	basesize += 1 // input count assumes inputCount<=255
	for i := 0; i < inputCount; i++ {
		inputsize := 32 + 4 // outpoint(hash+index)
		inputsize += 1      // len SignatureScript
		inputsize += 0      // SignatureScript
		inputsize += 4      // sequence
		basesize += inputsize
	}

	basesize += 1 // output count assumes outputCount<=255
	for i := 0; i < outputCount; i++ {
		outputsize := 8  // value
		outputsize += 1  // pk len
		outputsize += 34 // pk assumes that pk length is 34
		basesize += outputsize
	}

	basesize += 4 // nlocktime

	witnesssize := 0
	witnesssize += 2 // flag
	for i := 0; i < inputCount; i++ {
		witnesssizePerInput := 0
		witnesssizePerInput += 1 // witness len, assuming nOfMusig<=255

		for j := 0; j < nOfMusig-mOfMusig; j++ { // There is no signed public key
			witnesssizePerInput += 1 // witness buf len
			witnesssizePerInput += 0
		}
		for j := 0; j < mOfMusig; j++ { // The public key of the signatureublic key of the signature
			witnesssizePerInput += 1 // witness buf len
			if j%2 == 0 {
				witnesssizePerInput += 71 // There are 71 and 72 here fake random
			} else {
				witnesssizePerInput += 72 // There are 71 and 72
			}

		}

		witnesssizePerInput += 1 // witness buf len, assuming scriptLen《=255
		witnesssizePerInput += scriptLen

		witnesssize += witnesssizePerInput
	}

	// doc：https://bitcoin.stackexchange.com/questions/114375/how to accurately calculate the vsize of a transaction
	totalsize := basesize + witnesssize

	//fmt.Println("basesize=", basesize, "witnesssize=", witnesssize)
	weight := 3*basesize + totalsize // vsize = basesize + (totalsize-basesize)/4
	vsize := (weight + 4 - 1) / 4    // Not so precise
	return vsize
}
