package bitcoinlib

import (
	"context"
	"github.com/axengine/utils"
	"testing"
)

var multi_expporer_api = []string{"http://192.168.1.51:3001/"}

func TestSmartClient_GetTipHeight(t *testing.T) {
	height, err := NewSmartClient(multi_expporer_api).GetTipHeight(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(height)
}

func TestSmartClient_GetFeeRate(t *testing.T) {
	feeRate, err := NewSmartClient(multi_expporer_api).GetFeeRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(feeRate)
}

func TestSmartClient_getFeeEstimates(t *testing.T) {
	feeRate, err := NewSmartClient(multi_expporer_api).getFeeEstimates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(feeRate)
}

func TestSmartClient_GetAddress(t *testing.T) {
	rsp, err := NewSmartClient(multi_expporer_api).GetAddress(context.Background(), "tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestSmartClient_GetTransaction(t *testing.T) {
	tx, err := NewSmartClient(multi_expporer_api).GetTransaction(context.Background(), "e6cd51731c0b876eb32209d076d8242cc09ada0e23806944900d1e1c015d883c")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tx)
}

func TestSmartClient_GetAddressTransactions(t *testing.T) {
	tx, err := NewSmartClient(multi_expporer_api).GetAddressTransactions(context.Background(), "tb1q8f8rjp83055wky6hdnp3se6wefadedxws3f3zh5ee5vjkrt3m5xqgy7uay")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tx)
}

func TestSmartClient_GetAddressTransactionsMempool(t *testing.T) {
	tx, err := NewSmartClient(multi_expporer_api).GetAddressTransactionsMempool(context.Background(), "tb1q8f8rjp83055wky6hdnp3se6wefadedxws3f3zh5ee5vjkrt3m5xqgy7uay")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tx)
}

func TestSmartClient_GetAddressUtxos(t *testing.T) {
	rsp, err := NewSmartClient(multi_expporer_api).GetAddressUtxos(context.Background(), "tb1pwv6xu5xlv0puqdpxr2xjslh0mezm00cpcpnaq2ux77nqlxered9skmgty7")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestSmartClient_GetTransactionOutspend(t *testing.T) {
	rlt, err := NewSmartClient(multi_expporer_api).GetTransactionOutspend(context.Background(), "77a0f65725bc79a2f9d4bb40730991be95a744d629e15a43f36afebf3b2df0c1", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rlt)
}

func TestSmartClient_PostTransaction(t *testing.T) {
	txid, err := NewSmartClient(multi_expporer_api).PostTransaction(context.Background(), "0x111")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(txid)
}

func TestSmartClient_BlockHash(t *testing.T) {
	hash, err := NewSmartClient(multi_expporer_api).BlockHash(context.Background(), 3520900)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hash)
}

func TestSmartClient_BlockTxs(t *testing.T) {
	var index = 0
	var total int
	for {
		txs, err := NewSmartClient(multi_expporer_api).BlockTxs(context.Background(), "00000000040f05b4b7042e382a2a3cf65694bf137d1b43c054b113afd089b6e6", index)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(len(txs))
		total += len(txs)
		for idx, tx := range txs {
			t.Log(idx, ":", tx.Txid)
		}
		if len(txs) < 25 {
			break
		}

		index += len(txs)
	}
	t.Log("total:", total)
}
