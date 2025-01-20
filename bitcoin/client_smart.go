package bitcoinlib

import (
	"context"
	"fmt"
	"github.com/axengine/utils/log"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math"
	"strconv"
	"strings"
	"sync"
)

const ContentTypeJson = "application/json"
const ContentTypeText = "text/plain"

type SmartClient struct {
	clients []*resty.Client
	errors  map[string]int
	mu      sync.Mutex
}

func NewSmartClient(baseUrls []string) *SmartClient {
	c := &SmartClient{
		errors: make(map[string]int),
	}
	for _, v := range baseUrls {
		cli := resty.New().SetBaseURL(v)
		cli.SetCloseConnection(false)
		c.clients = append(c.clients, cli)
	}
	return c
}

func (m *SmartClient) Name() string {
	return "smart_client"
}

func (m *SmartClient) addError(baseURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[baseURL] += 1

	count := m.errors[baseURL]
	if count%100 == 0 {
		log.Logger.Warn("SmartClient", zap.String("baseURL", baseURL), zap.Int("errors", count))
	}
}

func (m *SmartClient) get(ctx context.Context, uri string, query, contentType string, result interface{}) (string, error) {
	for _, cli := range m.clients {
		r := cli.R().SetContext(ctx)
		if result != nil {
			r.SetResult(result)
		}
		if query != "" {
			r.SetQueryString(query)
		}
		if contentType != "" {
			r.ForceContentType(contentType)
		}
		rsp, err := r.Get(uri)
		if err != nil {
			m.addError(cli.BaseURL)
			log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.Error(err))
			continue
		}
		if rsp.IsError() {
			m.addError(cli.BaseURL)
			log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.String("status", rsp.Status()), zap.String("response", rsp.String()))
			continue
		}
		return string(rsp.Body()), nil
	}

	return "", errors.New("no api available")
}

func (m *SmartClient) post(ctx context.Context, uri string, body interface{}, reqContentType, rspContentType string, result interface{}) (string, error) {
	var finalErr = errors.New("post failed")
	for _, cli := range m.clients {
		r := cli.R().SetContext(ctx)

		if body != "" {
			r.SetBody(body)
		}
		if reqContentType != "" {
			r.SetHeader("Content-Type", reqContentType)
		}
		if rspContentType != "" {
			r.ForceContentType(rspContentType)
		}
		if result != nil {
			r.SetResult(result)
		}
		rsp, err := r.Post(uri)
		if err != nil {
			finalErr = errors.WithMessage(finalErr, err.Error())
			m.addError(cli.BaseURL)
			log.Logger.Debug("SmartClient:post", zap.String("baseURL", cli.BaseURL), zap.Error(err))
			continue
		}
		if rsp.IsError() {
			m.addError(cli.BaseURL)
			finalErr = errors.WithMessage(finalErr, rsp.String())
			log.Logger.Debug("SmartClient:post", zap.String("baseURL", cli.BaseURL), zap.String("status", rsp.Status()), zap.String("response", rsp.String()))
			continue
		}
		return string(rsp.Body()), nil
	}

	return "", finalErr
}

func (m *SmartClient) GetTipHeight(ctx context.Context) (int64, error) {
	body, err := m.get(ctx, "/blocks/tip/height", "", "", nil)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(body, 10, 64)
}

func (m *SmartClient) BlockHash(ctx context.Context, height int64) (string, error) {
	body, err := m.get(ctx, "/block-height/"+fmt.Sprintf("%d", height), "", "", nil)
	if err != nil {
		return "", err
	}
	return body, nil
}

func (m *SmartClient) BlockTxs(ctx context.Context, blockHash string, index int) ([]Transaction, error) {
	url := fmt.Sprintf("/block/%s/txs/%d", blockHash, index)
	var txs []Transaction
	for _, cli := range m.clients {
		r := cli.R().SetContext(ctx)
		r.SetResult(&txs)
		rsp, err := r.Get(url)
		if err != nil {
			m.addError(cli.BaseURL)
			log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.Error(err))
			continue
		}
		if rsp.IsError() && rsp.StatusCode() != 404 {
			m.addError(cli.BaseURL)
			log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.String("status", rsp.Status()), zap.String("response", rsp.String()))
			continue
		}
		// When index out of range
		if rsp.StatusCode() == 404 {
			return nil, nil
		}
		return txs, nil
	}
	return nil, errors.New("no api available")
}

// getFeeRecommended To get the referral rate, only mempool has this interface
func (m *SmartClient) getFeeRecommended(ctx context.Context) (int64, error) {
	var feeRsp struct {
		FastestFee  float32 `json:"fastestFee"`
		HalfHourFee float32 `json:"halfHourFee"`
		HourFee     float32 `json:"hourFee"`
		EconomyFee  float32 `json:"economyFee"`
		MinimumFee  float32 `json:"minimumFee"`
	}

	for _, cli := range m.clients {
		if strings.Contains(cli.BaseURL, "mempool.space") {
			r := cli.R().SetContext(ctx)
			r.SetResult(&feeRsp)
			rsp, err := r.Get("/v1/fees/recommended")
			if err != nil {
				m.addError(cli.BaseURL)
				log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.Error(err))
				continue
			}
			if rsp.IsError() {
				m.addError(cli.BaseURL)
				log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.String("status", rsp.Status()), zap.String("response", rsp.String()))
				continue
			}
			return int64(feeRsp.HalfHourFee), nil
		}
	}

	return 0, errors.New("no api available")
}

// getFeeEstimates blockstream/eletrs are the same, but the data returned by the mempool interface is in a different format
func (m *SmartClient) getFeeEstimates(ctx context.Context) (int64, error) {
	var fees = make(map[string]float32)
	for _, cli := range m.clients {
		if !strings.Contains(cli.BaseURL, "mempool.space") {
			r := cli.R().SetContext(ctx)
			r.SetResult(&fees)
			rsp, err := r.Get("fee-estimates")
			if err != nil {
				m.addError(cli.BaseURL)
				log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.Error(err))
				continue
			}
			if rsp.IsError() {
				m.addError(cli.BaseURL)
				log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.String("status", rsp.Status()), zap.String("response", rsp.String()))
				continue
			}
			log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.Any("fees", fees))
			// https://github.com/Blockstream/esplora/blob/master/API.md#get-fee-estimates
			// The data is likely to be incorrect, take the lowest rate of 1-25 blocks
			var minFeeRate = int64(math.Ceil(float64(fees["10"])))
			if minFeeRate == 0 {
				for i := 1; i <= 25; i++ {
					value := fees[fmt.Sprintf("%d", i)]
					feeRate := int64(math.Ceil(float64(value)))
					if minFeeRate == 0 {
						minFeeRate = feeRate
					} else {
						if feeRate < minFeeRate {
							minFeeRate = feeRate
						}
					}
				}
			}

			if minFeeRate > 0 {
				return minFeeRate, nil
			}
			// If there is none, 1 will be returned
			return 1, nil
		}
	}

	return 0, errors.New("no api available")
}

// GetFeeRate preferentially uses the mempool interface, otherwise the blockstream/electrs interface is used
func (m *SmartClient) GetFeeRate(ctx context.Context) (int64, error) {
	feeRate, _ := m.getFeeRecommended(ctx)
	if feeRate > 0 {
		return feeRate, nil
	}
	return m.getFeeEstimates(ctx)
}

func (m *SmartClient) GetTransaction(ctx context.Context, txid string) (*Transaction, error) {
	var tx = new(Transaction)

	for _, cli := range m.clients {
		r := cli.R().SetContext(ctx)
		r.SetResult(&tx)
		rsp, err := r.Get("/tx/" + txid)
		if err != nil {
			m.addError(cli.BaseURL)
			log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.Error(err))
			continue
		}
		if rsp.IsError() && rsp.StatusCode() != 404 {
			m.addError(cli.BaseURL)
			log.Logger.Debug("SmartClient:get", zap.String("baseURL", cli.BaseURL), zap.String("status", rsp.Status()), zap.String("response", rsp.String()))
			continue
		}
		// When the txid does not exist, the server returns a 404
		if rsp.StatusCode() == 404 {
			return nil, nil
		}
		return tx, nil
	}
	return nil, errors.New("no api available")
}

func (m *SmartClient) GetAddress(ctx context.Context, address string) (*AddressStats, error) {
	var addressStats = new(AddressStats)
	_, err := m.get(ctx, "/address/"+address, "", ContentTypeJson, addressStats)
	if err != nil {
		return nil, err
	}
	return addressStats, nil
}

func (m *SmartClient) GetAddressTransactions(ctx context.Context, address string) ([]Transaction, error) {
	var txs = make([]Transaction, 0)
	_, err := m.get(ctx, "/address/"+address+"/txs", "", ContentTypeJson, &txs)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (m *SmartClient) GetAddressTransactionsMempool(ctx context.Context, address string) ([]Transaction, error) {
	var txs = make([]Transaction, 0)
	uri := "/address/" + address + "/txs/mempool"
	_, err := m.get(ctx, uri, "", ContentTypeJson, &txs)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (m *SmartClient) GetAddressTransactionsChain(ctx context.Context, address string, lastSeenTxid string) ([]Transaction, error) {
	var txs = make([]Transaction, 0)
	uri := "/address/" + address + "/txs/chain"
	if lastSeenTxid != "" {
		uri = uri + "/" + lastSeenTxid
	}
	_, err := m.get(ctx, uri, "", ContentTypeJson, &txs)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (m *SmartClient) GetAddressUtxos(ctx context.Context, address string) ([]UTXO, error) {
	var utxos = make([]UTXO, 0)
	_, err := m.get(ctx, "/address/"+address+"/utxo", "", ContentTypeJson, &utxos)
	if err != nil {
		return nil, err
	}
	return utxos, nil
}

func (m *SmartClient) GetTransactionOutspend(ctx context.Context, txid string, vout int) (*Outspend, error) {
	var outspend = new(Outspend)
	_, err := m.get(ctx, fmt.Sprintf("/tx/%s/outspend/%d", txid, vout), "", ContentTypeJson, outspend)
	if err != nil {
		return nil, err
	}
	return outspend, nil
}

func (m *SmartClient) PostTransaction(ctx context.Context, txHex string) (string, error) {
	rsp, err := m.post(ctx, "/tx", txHex, ContentTypeText, ContentTypeJson, nil)
	if err != nil {
		return "", err
	}
	return rsp, nil
}
