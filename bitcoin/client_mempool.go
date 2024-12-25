package bitcoinlib

import (
	"context"
	"errors"
	"fmt"
	"github.com/axengine/utils"
	"github.com/go-resty/resty/v2"
	"math"
	"strconv"
)

type Mempool struct {
	client  *resty.Client
	baseURL string
}

func NewMempool(baseURL string) *Mempool {
	cli := resty.New().SetBaseURL(baseURL)
	cli.SetCloseConnection(false)
	return &Mempool{
		client:  cli,
		baseURL: baseURL,
	}
}

func (m *Mempool) Name() string {
	return "mempool"
}

func (m *Mempool) GetTipHeight(ctx context.Context) (int64, error) {
	rsp, err := m.client.R().SetContext(ctx).Get("/blocks/tip/height")
	if err != nil {
		return 0, err
	}
	if rsp.IsError() {
		return 0, errors.New(rsp.Status() + ":" + rsp.String())
	}
	return strconv.ParseInt(rsp.String(), 10, 64)
}

func (m *Mempool) GetFeeRate(ctx context.Context) (int64, error) {
	var feeRsp struct {
		FastestFee  float32 `json:"fastestFee"`
		HalfHourFee float32 `json:"halfHourFee"`
		HourFee     float32 `json:"hourFee"`
		EconomyFee  float32 `json:"economyFee"`
		MinimumFee  float32 `json:"minimumFee"`
	}
	rsp, err := m.client.R().SetContext(ctx).SetResult(&feeRsp).Get("/v1/fees/recommended")
	if err != nil {
		return 0, err
	}
	if rsp.IsError() {
		return 0, errors.New(rsp.Status() + ":" + rsp.String())
	}
	fmt.Println(utils.JsonPretty(feeRsp))
	return int64(feeRsp.HalfHourFee), err
}

func (m *Mempool) GetFeeEstimates(ctx context.Context) (int, error) {
	var fees = make(map[string]float32)
	_, err := m.client.R().SetContext(ctx).SetResult(&fees).Get("fee-estimates")
	if err != nil {
		return 0, err
	}
	if _, ok := fees["3"]; ok {
		// Rounding up the handling fee can be a big one
		return int(math.Ceil(float64(fees["3"]))), nil
	}
	return 1, nil
}

func (m *Mempool) GetTransaction(ctx context.Context, txid string) (*Transaction, error) {
	var tx = new(Transaction)
	rsp, err := m.client.R().ForceContentType("application/json").SetResult(tx).SetContext(ctx).Get("/tx/" + txid)
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, errors.New(rsp.Status() + ":" + rsp.String())
	}
	return tx, nil
}

func (m *Mempool) GetAddressTransactions(ctx context.Context, address string) ([]Transaction, error) {
	var txs = make([]Transaction, 0)
	r := m.client.R().ForceContentType("application/json").SetResult(&txs).SetContext(ctx)
	rsp, err := r.Get("/address/" + address + "/txs")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, errors.New(rsp.Status() + ":" + rsp.String())
	}
	return txs, nil
}

// GetAddressTransactionsMempool Returns up to 50 transactions (no paging).
func (m *Mempool) GetAddressTransactionsMempool(ctx context.Context, address string) ([]Transaction, error) {
	var txs = make([]Transaction, 0)
	r := m.client.R().ForceContentType("application/json").SetResult(&txs).SetContext(ctx)
	uri := "/address/" + address + "/txs/mempool"
	rsp, err := r.Get(uri)
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, errors.New(rsp.Status() + ":" + rsp.String())
	}
	return txs, nil
}

func (m *Mempool) GetAddressTransactionsChain(ctx context.Context, address string, last_seen_txid string) ([]Transaction, error) {
	var txs = make([]Transaction, 0)
	r := m.client.R().ForceContentType("application/json").SetResult(&txs).SetContext(ctx)
	uri := "/address/" + address + "/txs/chain"
	if last_seen_txid != "" {
		uri = uri + "/" + last_seen_txid
	}
	rsp, err := r.Get(uri)
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, errors.New(rsp.Status() + ":" + rsp.String())
	}
	return txs, nil
}

func (m *Mempool) GetAddressUtxos(ctx context.Context, address string) ([]UTXO, error) {
	var utxos = make([]UTXO, 0)
	rsp, err := m.client.R().ForceContentType("application/json").SetResult(&utxos).SetContext(ctx).Get("/address/" + address + "/utxo")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, errors.New(rsp.Status() + ":" + rsp.String())
	}
	return utxos, nil
}

func (m *Mempool) GetTransactionOutspend(ctx context.Context, txid string, vout int) (*Outspend, error) {
	var outspend = new(Outspend)
	rsp, err := m.client.R().ForceContentType("application/json").SetResult(outspend).SetContext(ctx).Get(fmt.Sprintf("/tx/%s/outspend/%d", txid, vout))
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, errors.New(rsp.Status() + ":" + rsp.String())
	}
	return outspend, nil
}

func (m *Mempool) PostTransaction(ctx context.Context, txHex string) (string, error) {
	rsp, err := m.client.R().SetContext(ctx).SetHeader("Content-Type", "text/plain").SetBody(txHex).Post("/tx")
	if err != nil {
		return "", err
	}
	if rsp.IsError() {
		return "", errors.New(rsp.Status() + ":" + rsp.String())
	}
	return rsp.String(), nil
}
