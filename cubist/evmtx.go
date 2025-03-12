package cubist

type TransactionV1 struct {
	From     string `json:"from"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Nonce    string `json:"nonce"`
	To       string `json:"to"`
	Type     string `json:"type"` // 0x00
	Value    string `json:"value"`
	Data     string `json:"data"`
}

type TransactionV2 struct {
	ChainID              string      `json:"chain_id"`
	From                 string      `json:"from"`
	Gas                  string      `json:"gas"`
	MaxFeePerGas         string      `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string      `json:"maxPriorityFeePerGas"`
	Nonce                string      `json:"nonce"`
	To                   string      `json:"to"`
	Type                 string      `json:"type"`
	Value                string      `json:"value"`
	Data                 string      `json:"data,omitempty"`
	AccessList           interface{} `json:"accessList,omitempty"`
}

type TransactionData struct {
	Metadata interface{} `json:"metadata,omitempty"`
	ChainID  int         `json:"chain_id"`
	Tx       interface{} `json:"tx"`
}
