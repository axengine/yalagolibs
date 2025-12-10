package bitcoinlib

import (
	"context"
)

type IBitcoinClient interface {
	Name() string

	GetTipHeight(ctx context.Context) (int64, error)
	BlockHash(ctx context.Context, height int64) (string, error)
	BlockTxs(ctx context.Context, blockHash string, index int) ([]Transaction, error)
	GetFeeRate(ctx context.Context) (int64, error)
	GetTransaction(ctx context.Context, txid string) (*Transaction, error)
	GetAddress(ctx context.Context, address string) (*AddressStats, error)
	GetAddressTransactionsMempool(ctx context.Context, address string) ([]Transaction, error)
	// GetAddressTransactions Returns up to 50 mempool transactions plus the first 25 confirmed transactions.
	GetAddressTransactions(ctx context.Context, address string) ([]Transaction, error)
	// GetAddressTransactionsChain Returns 25 transactions per page.
	GetAddressTransactionsChain(ctx context.Context, address string, last_seen_txid string) ([]Transaction, error)
	GetAddressUtxos(ctx context.Context, address string) ([]UTXO, error)
	GetTransactionOutspend(ctx context.Context, txid string, vout int) (*Outspend, error)

	PostTransaction(ctx context.Context, txHex string) (string, error)
}
