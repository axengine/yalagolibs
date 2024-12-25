package bitcoinlib

import "testing"

func TestCalculateP2WSHVSize(t *testing.T) {
	t.Log(CalculateP2WSHVSize(1, 2))
}

func TestCalculateP2WSHVSizeV2(t *testing.T) {
	inputCount, outputCount := 1, 2

	t.Log(CalculateP2WSHVSize(inputCount, outputCount))
	t.Log(CalculateP2WSHVSizeV2(inputCount, outputCount, 2, 3, 105))
}
