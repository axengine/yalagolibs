package bitcoinlib

import (
	"context"
	"github.com/axengine/utils"
	"testing"
)

const URL_MEMPOOL = "https://mempool.space:443/testnet/api/"

func TestGetTipHeight(t *testing.T) {
	height, err := NewMempool(URL_MEMPOOL).GetTipHeight(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(height)
}

func TestGetFeeRate(t *testing.T) {
	cli := NewMempool(URL_MEMPOOL)
	feeRate, err := cli.GetFeeRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("feeRate", feeRate)

	if _, err := cli.GetFeeEstimates(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGetFeeEstimates(t *testing.T) {
	cli := NewMempool(URL_MEMPOOL)

	if _, err := cli.GetFeeEstimates(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGetTransaction(t *testing.T) {
	rsp, err := NewMempool(URL_MEMPOOL).GetTransaction(context.Background(), "060ababdc23c14ec5d5dc606f790b9f594365f4ab5876684c5695b1d2af20e2d")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

// tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp
func TestGetAddressTransactions(t *testing.T) {
	rsp, err := NewMempool(URL_MEMPOOL).GetAddressTransactions(context.Background(),
		"tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(rsp))
	for _, v := range rsp {
		t.Log(v.Txid, v.Status)
	}

	//t.Log(utils.JsonPretty(rsp))
}

func TestGetAddressTransactionsChain(t *testing.T) {
	rsp, err := NewMempool(URL_MEMPOOL).GetAddressTransactionsChain(context.Background(),
		"tb1ph8jzyv68rwx436wtnr760tsg3ut07ml8zueuqcwww04fdmrp8ntqaqjmsp",
		"b1b10f8cb96a561798996108130d91380c07fe5ab8cb3cf03514748d8430e0bc")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(rsp))
	for _, v := range rsp {
		t.Log(v.Txid, v.Status)
	}
}

func TestGetAddressUtxos(t *testing.T) {
	rsp, err := NewMempool(URL_MEMPOOL).GetAddressUtxos(context.Background(), "tb1qq40hh9f68qjtahe3nm5q2tvskaxgefwr5mysm2gac5get8qaalfsyxqar4")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestPostTransaction(t *testing.T) {
	rsp, err := NewMempool(URL_MEMPOOL).PostTransaction(context.Background(), "010000000001012d0ef22a1d5b69c5846687b54a5f3694f5b990f706c65d5dec143cc2bdba0a060100000000fdffffff021027000000000000225120b9e42233471b8d58e9cb98fda7ae088f16ff6fe71733c061ce73ea96ec613cd6e817000000000000220020dbef95cd020bcc48fd0062524e673c21d59da1d2e872963fa9e2ed35e0b5a39902473044022011dcbab3bb1e20e419b948739f11822aa9f1fd9d1bd4185573286de203430dad02204192b0fddc13c05c4de239b92721b247a25aa2810329c9c9dc6058c477fc827901232102c8e1863b7c11867bcd261c8646cb2c27eefc50b15a2d901bb63c8a9305da2308ac00000000")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}
