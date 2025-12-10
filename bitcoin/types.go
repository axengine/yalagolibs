package bitcoinlib

type UTXO struct {
	Txid   string `json:"txid"`
	Vout   uint32 `json:"vout"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int64  `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
	Value int64 `json:"value"`
}

type TransactionStatus struct {
	BlockHash   string `json:"block_hash"`
	BlockHeight int64  `json:"block_height"`
	BlockTime   int64  `json:"block_time"`
	Confirmed   bool   `json:"confirmed"`
}
type Prevout struct {
	ScriptPubkey        string `json:"scriptpubkey"`
	ScriptPubkeyAddress string `json:"scriptpubkey_address"`
	ScriptPubkeyASM     string `json:"scriptpubkey_asm"`
	ScriptPubkeyType    string `json:"scriptpubkey_type"`
	Value               int64  `json:"value"`
}
type Vin struct {
	InnerWitnessScriptASM string   `json:"inner_witnessscript_asm"`
	IsCoinbase            bool     `json:"is_coinbase"`
	Prevout               Prevout  `json:"prevout"`
	ScriptSig             string   `json:"scriptsig"`
	ScriptSigASM          string   `json:"scriptsig_asm"`
	Sequence              uint32   `json:"sequence"`
	Txid                  string   `json:"txid"`
	Vout                  uint32   `json:"vout"`
	Witness               []string `json:"witness"`
}

type Vout struct {
	ScriptPubkey        string `json:"scriptpubkey"`
	ScriptPubkeyAddress string `json:"scriptpubkey_address"`
	ScriptPubkeyASM     string `json:"scriptpubkey_asm"`
	ScriptPubkeyType    string `json:"scriptpubkey_type"`
	Value               int64  `json:"value"`
}

type Transaction struct {
	Fee      int64              `json:"fee"`
	Locktime uint32             `json:"locktime"`
	Sigops   int                `json:"sigops"`
	Size     int                `json:"size"`
	Status   *TransactionStatus `json:"status,omitempty"`
	Txid     string             `json:"txid"`
	Version  int                `json:"version"`
	Vin      []Vin              `json:"vin"`
	Vout     []Vout             `json:"vout"`
	Weight   int                `json:"weight"`
}

type Outspend struct {
	Spent  bool   `json:"spent"`
	Txid   string `json:"txid"`
	Vin    int    `json:"vin"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int64  `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
}

type ChainStats struct {
	FundedTxoCount int64 `json:"funded_txo_count"`
	FundedTxoSum   int64 `json:"funded_txo_sum"`
	SpentTxoCount  int64 `json:"spent_txo_count"`
	SpentTxoSum    int64 `json:"spent_txo_sum"`
	TxCount        int64 `json:"tx_count"`
}

type MempoolStats struct {
	FundedTxoCount int64 `json:"funded_txo_count"`
	FundedTxoSum   int64 `json:"funded_txo_sum"`
	SpentTxoCount  int64 `json:"spent_txo_count"`
	SpentTxoSum    int64 `json:"spent_txo_sum"`
	TxCount        int64 `json:"tx_count"`
}

type AddressStats struct {
	Address      string       `json:"address"`
	ChainStats   ChainStats   `json:"chain_stats"`
	MempoolStats MempoolStats `json:"mempool_stats"`
}
